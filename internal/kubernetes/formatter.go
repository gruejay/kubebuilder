package kubernetes

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

func CleanData(obj unstructured.Unstructured) unstructured.Unstructured {

	delete(obj.Object["metadata"].(map[string]any), "managedFields")
	return obj
}
