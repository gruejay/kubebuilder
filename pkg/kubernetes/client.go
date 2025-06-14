package kubernetes

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ResourceInfo struct {
	GVR        schema.GroupVersionResource
	GVK        schema.GroupVersionKind
	Namespaced bool
	IsCustom   bool
}

type UnifiedClient struct {
	typedClient     kubernetes.Interface
	dynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	crdClient       apiextclient.Interface
	config          *rest.Config

	// Resource discovery cache
	resourceCache map[schema.GroupVersionResource]*ResourceInfo
	cacheMutex    sync.RWMutex
	lastDiscovery time.Time
	cacheTimeout  time.Duration
}

func NewUnifiedClient() (*UnifiedClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	typedClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	crdClient, err := apiextclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	client := &UnifiedClient{
		typedClient:     typedClient,
		dynamicClient:   dynamicClient,
		discoveryClient: discoveryClient,
		crdClient:       crdClient,
		config:          config,
		resourceCache:   make(map[schema.GroupVersionResource]*ResourceInfo),
		cacheTimeout:    5 * time.Minute, // Cache for 5 minutes
	}

	// Initial discovery
	if err := client.discoverResources(); err != nil {
		return nil, fmt.Errorf("initial resource discovery failed: %w", err)
	}

	return client, nil
}

// Discover all available resources (core + custom)
func (c *UnifiedClient) discoverResources() error {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	// Clear existing cache
	c.resourceCache = make(map[schema.GroupVersionResource]*ResourceInfo)

	// Discover core resources
	if err := c.discoverCoreResources(); err != nil {
		return fmt.Errorf("failed to discover core resources: %w", err)
	}

	// Discover custom resources
	if err := c.discoverCustomResources(); err != nil {
		return fmt.Errorf("failed to discover custom resources: %w", err)
	}

	c.lastDiscovery = time.Now()
	return nil
}

// Discover core Kubernetes resources
func (c *UnifiedClient) discoverCoreResources() error {
	coreResources := []ResourceInfo{
		{GVR: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			GVK:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
			Namespaced: true, IsCustom: false},
		{GVR: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"},
			GVK:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
			Namespaced: true, IsCustom: false},
		{GVR: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
			GVK:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
			Namespaced: true, IsCustom: false},
		{GVR: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"},
			GVK:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
			Namespaced: true, IsCustom: false},
		{GVR: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"},
			GVK:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
			Namespaced: false, IsCustom: false},
		{GVR: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
			GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
			Namespaced: true, IsCustom: false},
		{GVR: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicasets"},
			GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
			Namespaced: true, IsCustom: false},
		{GVR: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"},
			GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DaemonSet"},
			Namespaced: true, IsCustom: false},
		{GVR: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"},
			GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
			Namespaced: true, IsCustom: false},
	}

	for _, resource := range coreResources {
		resourceCopy := resource
		c.resourceCache[resource.GVR] = &resourceCopy
	}

	return nil
}

// Discover custom resources from CRDs
func (c *UnifiedClient) discoverCustomResources() error {
	// Get all CRDs
	crdList, err := c.crdClient.ApiextensionsV1().CustomResourceDefinitions().List(
		context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, crd := range crdList.Items {
		// Process each version of the CRD
		for _, version := range crd.Spec.Versions {
			if !version.Served {
				continue // Skip versions that aren't served
			}

			gvr := schema.GroupVersionResource{
				Group:    crd.Spec.Group,
				Version:  version.Name,
				Resource: crd.Spec.Names.Plural,
			}

			gvk := schema.GroupVersionKind{
				Group:   crd.Spec.Group,
				Version: version.Name,
				Kind:    crd.Spec.Names.Kind,
			}

			resourceInfo := &ResourceInfo{
				GVR:        gvr,
				GVK:        gvk,
				Namespaced: crd.Spec.Scope == apiextv1.NamespaceScoped,
				IsCustom:   true,
			}

			c.resourceCache[gvr] = resourceInfo
		}
	}

	return nil
}

// Check if cache needs refresh and refresh if needed
func (c *UnifiedClient) ensureFreshCache() error {
	c.cacheMutex.RLock()
	needsRefresh := time.Since(c.lastDiscovery) > c.cacheTimeout
	c.cacheMutex.RUnlock()

	if needsRefresh {
		return c.discoverResources()
	}
	return nil
}

// Get resource info from cache
func (c *UnifiedClient) getResourceInfo(gvr schema.GroupVersionResource) (*ResourceInfo, error) {
	if err := c.ensureFreshCache(); err != nil {
		return nil, err
	}

	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	info, exists := c.resourceCache[gvr]
	if !exists {
		return nil, fmt.Errorf("resource %v not found in cluster", gvr)
	}

	return info, nil
}

// Generic operations that work with discovered resources
func (c *UnifiedClient) Get(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, obj any) error {
	resourceInfo, err := c.getResourceInfo(gvr)
	if err != nil {
		return err
	}

	// Validate namespace usage
	if namespace != "" && !resourceInfo.Namespaced {
		return fmt.Errorf("resource %v is cluster-scoped, cannot specify namespace", gvr)
	}

	if !resourceInfo.IsCustom {
		return c.getTypedResource(ctx, gvr, namespace, name, obj)
	}
	return c.getDynamicResource(ctx, gvr, namespace, name, obj)
}

func (c *UnifiedClient) List(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj any) error {
	resourceInfo, err := c.getResourceInfo(gvr)
	if err != nil {
		return err
	}

	// Validate namespace usage
	if namespace != "" && !resourceInfo.Namespaced {
		return fmt.Errorf("resource %v is cluster-scoped, cannot specify namespace", gvr)
	}

	if !resourceInfo.IsCustom {
		return c.listTypedResource(ctx, gvr, namespace, obj)
	}
	return c.listDynamicResource(ctx, gvr, namespace, obj)
}

// func (c *UnifiedClient) Create(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj any) error {
// 	resourceInfo, err := c.getResourceInfo(gvr)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Validate namespace usage
// 	if namespace != "" && !resourceInfo.Namespaced {
// 		return fmt.Errorf("resource %v is cluster-scoped, cannot specify namespace", gvr)
// 	}
//
// 	if !resourceInfo.IsCustom {
// 		return c.createTypedResource(ctx, gvr, namespace, obj)
// 	}
// 	return c.createDynamicResource(ctx, gvr, namespace, obj)
// }

// List all available resources in the cluster
func (c *UnifiedClient) ListAvailableResources() ([]ResourceInfo, error) {
	if err := c.ensureFreshCache(); err != nil {
		return nil, err
	}

	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	var resources []ResourceInfo
	for _, info := range c.resourceCache {
		resources = append(resources, *info)
	}

	return resources, nil
}

// List only custom resources
func (c *UnifiedClient) ListCustomResources() ([]ResourceInfo, error) {
	allResources, err := c.ListAvailableResources()
	if err != nil {
		return nil, err
	}

	var customResources []ResourceInfo
	for _, resource := range allResources {
		if resource.IsCustom {
			customResources = append(customResources, resource)
		}
	}

	return customResources, nil
}

// Check if a resource exists in the cluster
func (c *UnifiedClient) ResourceExists(gvr schema.GroupVersionResource) bool {
	_, err := c.getResourceInfo(gvr)
	return err == nil
}

// Get GVK from GVR
func (c *UnifiedClient) GetGVK(gvr schema.GroupVersionResource) (schema.GroupVersionKind, error) {
	resourceInfo, err := c.getResourceInfo(gvr)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}
	return resourceInfo.GVK, nil
}

// Force refresh of resource cache
func (c *UnifiedClient) RefreshResourceCache() error {
	return c.discoverResources()
}

// Handle typed resource operations
func (c *UnifiedClient) getTypedResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, obj any) error {
	// Check if target is unstructured, if so use dynamic client for consistency
	if _, ok := obj.(*unstructured.Unstructured); ok {
		return c.getDynamicResource(ctx, gvr, namespace, name, obj)
	}

	switch gvr {
	case schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}:
		pod, err := c.typedClient.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(*pod))

	case schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}:
		svc, err := c.typedClient.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(*svc))

	case schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}:
		deploy, err := c.typedClient.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(*deploy))

	default:
		return fmt.Errorf("unsupported core resource: %v", gvr)
	}

	return nil
}

func (c *UnifiedClient) listTypedResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj any) error {
	// Check if target is unstructured, if so use dynamic client for consistency
	if _, ok := obj.(*unstructured.UnstructuredList); ok {
		return c.listDynamicResource(ctx, gvr, namespace, obj)
	}

	switch gvr {
	case schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}:
		pods, err := c.typedClient.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(*pods))

	case schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}:
		deployments, err := c.typedClient.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(*deployments))

	default:
		return fmt.Errorf("unsupported core resource: %v", gvr)
	}

	return nil
}

// Handle dynamic resource operations
func (c *UnifiedClient) getDynamicResource(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string, obj any) error {
	resourceInterface := c.getResourceInterface(gvr, namespace)
	unstructuredObj, err := resourceInterface.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Convert unstructured to the target type if it's not already unstructured
	if _, ok := obj.(*unstructured.Unstructured); ok {
		reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(*unstructuredObj))
	} else {
		// Convert unstructured to typed object
		return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.Object, obj)
	}

	return nil
}

func (c *UnifiedClient) listDynamicResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj any) error {
	resourceInterface := c.getResourceInterface(gvr, namespace)
	unstructuredList, err := resourceInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Handle the result based on the target type
	if _, ok := obj.(*unstructured.UnstructuredList); ok {
		reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(*unstructuredList))
	} else {
		// Convert to typed list if needed
		return runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredList.Object, obj)
	}

	return nil
}

func (c *UnifiedClient) createDynamicResource(ctx context.Context, gvr schema.GroupVersionResource, namespace string, obj any) error {
	resourceInterface := c.getResourceInterface(gvr, namespace)

	var unstructuredObj *unstructured.Unstructured
	var err error

	if u, ok := obj.(*unstructured.Unstructured); ok {
		unstructuredObj = u
	} else {
		// Convert typed object to unstructured
		unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}
		unstructuredObj = &unstructured.Unstructured{Object: unstructuredMap}
	}

	_, err = resourceInterface.Create(ctx, unstructuredObj, metav1.CreateOptions{})
	return err
}

func (c *UnifiedClient) getResourceInterface(gvr schema.GroupVersionResource, namespace string) dynamic.ResourceInterface {
	if namespace != "" {
		return c.dynamicClient.Resource(gvr).Namespace(namespace)
	}
	return c.dynamicClient.Resource(gvr)
}

// func main() {
// 	client, err := NewUnifiedClient()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	ctx := context.Background()
//
// 	// List all available resources (core + custom)
// 	resources, err := client.ListAvailableResources()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	fmt.Printf("Found %d total resources in cluster:\n", len(resources))
// 	for _, resource := range resources {
// 		scope := "Namespaced"
// 		if !resource.Namespaced {
// 			scope = "Cluster-scoped"
// 		}
// 		resourceType := "Core"
// 		if resource.IsCustom {
// 			resourceType = "Custom"
// 		}
// 		fmt.Printf("- %s (%s, %s): %s\n",
// 			resource.GVR.Resource,
// 			resourceType,
// 			scope,
// 			resource.GVK.Kind)
// 	}
//
// 	// List only custom resources
// 	customResources, err := client.ListCustomResources()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	fmt.Printf("\nFound %d custom resources:\n", len(customResources))
// 	for _, resource := range customResources {
// 		fmt.Printf("- %s.%s/%s (Kind: %s)\n",
// 			resource.GVR.Resource,
// 			resource.GVR.Group,
// 			resource.GVR.Version,
// 			resource.GVK.Kind)
// 	}
//
// 	// Try to work with a custom resource if it exists
// 	exampleGVR := schema.GroupVersionResource{
// 		Group:    "kubevirt.io",
// 		Version:  "v1",
// 		Resource: "virtualmachineinstances",
// 	}
//
// 	if client.ResourceExists(exampleGVR) {
// 		fmt.Printf("\nCustom resource %v exists, attempting to list...\n", exampleGVR)
//
// 		var customList unstructured.UnstructuredList
// 		err = client.List(ctx, exampleGVR, "cri-proj", &customList)
// 		if err != nil {
// 			fmt.Printf("Error listing custom resources: %v\n", err)
// 		} else {
// 			fmt.Printf("Found %d custom resources\n", len(customList.Items))
// 		}
// 	} else {
// 		fmt.Printf("\nCustom resource %v not found in cluster\n", exampleGVR)
// 	}
//
// 	// Work with core resources normally
// 	podGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
// 	var podList unstructured.UnstructuredList
// 	err = client.List(ctx, podGVR, "default", &podList)
// 	if err != nil {
// 		fmt.Printf("Error listing pods: %v\n", err)
// 	} else {
// 		fmt.Printf("Found %d pods in default namespace\n", len(podList.Items))
// 	}
// }
