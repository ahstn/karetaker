package kubernetes

import (
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListDeployments returns a list of the deployments in the current cluster
func ListDeployments(clientset kubernetes.Interface) ([]string, error) {
	list, err := clientset.AppsV1().Deployments("default").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "getting deployments")
	}

	// TODO: Parse ObjectMeta.CreationTimestamp
	deployments := []string{}
	for _, deployment := range list.Items {
		deployments = append(deployments, deployment.ObjectMeta.Name)
	}

	return deployments, nil
}
