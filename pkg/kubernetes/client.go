package kubernetes

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Client struct {
	clientset     *kubernetes.Clientset
	dynamicClient dynamic.Interface
}

type Resource struct {
	Name   string
	Type   string
	Status string
}

func NewClient() (*Client, string, error) {
	clientset, dynamicClient, currentNamespace, err := loadKubeConfig()
	if err != nil {
		return nil, "", err
	}

	return &Client{
		clientset:     clientset,
		dynamicClient: dynamicClient,
	}, currentNamespace, nil
}

func (c *Client) GetNamespaces() ([]string, error) {
	ctx := context.Background()
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var nsNames []string
	for _, ns := range namespaces.Items {
		nsNames = append(nsNames, ns.Name)
	}

	return nsNames, nil
}

func (c *Client) GetPodsInNamespace(namespace string) ([]Resource, error) {
	ctx := context.Background()
	pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []Resource
	for _, pod := range pods.Items {
		resources = append(resources, Resource{
			Name:   pod.Name,
			Type:   "Pod",
			Status: string(pod.Status.Phase),
		})
	}

	return resources, nil
}

func (c *Client) GetServicesInNamespace(namespace string) ([]Resource, error) {
	ctx := context.Background()
	services, err := c.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []Resource
	for _, svc := range services.Items {
		resources = append(resources, Resource{
			Name:   svc.Name,
			Type:   "Service",
			Status: "Active",
		})
	}

	return resources, nil
}

func loadKubeConfig() (*kubernetes.Clientset, dynamic.Interface, string, error) {
	var kubeconfig string
	
	if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
		kubeconfig = kubeconfigEnv
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, nil, "", err
		}
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, "", err
	}

	// Get current namespace from kubeconfig
	configAccess := clientcmd.NewDefaultPathOptions()
	configAccess.GlobalFile = kubeconfig
	rawConfig, err := configAccess.GetStartingConfig()
	if err != nil {
		return nil, nil, "", err
	}

	currentContext := rawConfig.CurrentContext
	currentNamespace := "default" // fallback
	if context, exists := rawConfig.Contexts[currentContext]; exists {
		if context.Namespace != "" {
			currentNamespace = context.Namespace
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, "", err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, "", err
	}

	return clientset, dynamicClient, currentNamespace, nil
}

func (c *Client) GetResourceDetails(resourceType string, resourceName string, namespace string) (string, error) {
	ctx := context.Background()
	
	gvr, err := c.getGVRForResourceType(resourceType)
	if err != nil {
		return "", err
	}
	
	resource, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, resourceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	
	yamlData, err := yaml.Marshal(resource.Object)
	if err != nil {
		return "", err
	}
	
	return string(yamlData), nil
}

func (c *Client) getGVRForResourceType(resourceType string) (schema.GroupVersionResource, error) {
	resourceTypeLower := strings.ToLower(resourceType)
	
	switch resourceTypeLower {
	case "pod":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}, nil
	case "service":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}, nil
	case "deployment":
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, nil
	case "configmap":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}, nil
	case "secret":
		return schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}, nil
	case "ingress":
		return schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}, nil
	case "daemonset":
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}, nil
	case "statefulset":
		return schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}, nil
	case "job":
		return schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}, nil
	case "cronjob":
		return schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjobs"}, nil
	default:
		return schema.GroupVersionResource{}, nil
	}
}