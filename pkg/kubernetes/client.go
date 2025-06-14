package kubernetes

import (
	"context"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Client struct {
	clientset *kubernetes.Clientset
}

type Resource struct {
	Name   string
	Type   string
	Status string
}

func NewClient() (*Client, string, error) {
	clientset, currentNamespace, err := loadKubeConfig()
	if err != nil {
		return nil, "", err
	}

	return &Client{clientset: clientset}, currentNamespace, nil
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

func loadKubeConfig() (*kubernetes.Clientset, string, error) {
	var kubeconfig string
	
	if kubeconfigEnv := os.Getenv("KUBECONFIG"); kubeconfigEnv != "" {
		kubeconfig = kubeconfigEnv
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, "", err
		}
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, "", err
	}

	// Get current namespace from kubeconfig
	configAccess := clientcmd.NewDefaultPathOptions()
	configAccess.GlobalFile = kubeconfig
	rawConfig, err := configAccess.GetStartingConfig()
	if err != nil {
		return nil, "", err
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
		return nil, "", err
	}

	return clientset, currentNamespace, nil
}