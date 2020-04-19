package kubernetes

import (
	"time"

	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Deployment struct {
	Name string
	Age  time.Duration
}

// ListDeployments returns a list of the deployments in the current cluster
func ListDeployments(clientset kubernetes.Interface) ([]Deployment, error) {
	list, err := clientset.AppsV1().Deployments("default").List(meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "getting deployments")
	}

	now := time.Now()
	// TODO: Parse ObjectMeta.CreationTimestamp
	deployments := []Deployment{}
	for _, deployment := range list.Items {
		deployments = append(deployments, Deployment{
			Name: deployment.ObjectMeta.Name,
			Age:  now.Sub(deployment.ObjectMeta.CreationTimestamp.Time),
		})
	}

	return deployments, nil
}
