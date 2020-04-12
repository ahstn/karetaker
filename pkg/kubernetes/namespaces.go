package kubernetes

import (
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListNamespaces returns a list of the namespaces in the current cluster
func ListNamespaces(clientset kubernetes.Interface) ([]string, error) {
	list, err := clientset.CoreV1().Namespaces().List(meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "getting namespaces")
	}

	namespaces := []string{}
	for _, namespace := range list.Items {
		namespaces = append(namespaces, namespace.ObjectMeta.Name)
	}

	return namespaces, nil
}
