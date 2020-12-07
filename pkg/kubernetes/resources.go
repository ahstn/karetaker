package kubernetes

import (
	"context"
	"time"

	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Resource is a stripped down version of a Kubernetes Resource.
// It only holds the name and age of the resource.
type Resource struct {
	Name string
	Kind string
	Age  time.Duration
}


// ResourcesOlderThan returns a list of the resources older than the duration 'd'
func ResourcesOlderThan(c dynamic.Interface, r schema.GroupVersionResource, n string, d time.Duration) ([]Resource, error) {
	list, err := c.Resource(r).Namespace(n).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "getting resource")
	}

	now := time.Now()
	var resource []Resource
	for _, deployment := range list.Items {
		t, found, err := unstructured.NestedString(deployment.Object, "metadata", "creationTimestamp")
		if err != nil || !found {
			return nil, err
		}

		creation, err := time.Parse(time.RFC3339, t)

		age := now.Sub(creation)
		if age > d {
			name, found, err := unstructured.NestedString(deployment.Object, "metadata", "name")
			if err != nil || !found {
				return nil, err
			}

			resource = append(resource, Resource{
				Name: name,
				Kind: r.Resource,
				Age:  age.Round(time.Minute),
			})
		}
	}

	return resource, nil
}