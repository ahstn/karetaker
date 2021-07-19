package kubernetes

import (
	"context"
	"strings"
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
	Name   string
	Kind   string
	Age    time.Duration
	Status Status
}

type Status int

const (
	Running   Status = iota
	Completed Status = iota
	Failed    Status = iota
	Unknown   Status = iota
)

// Resources returns all the existing objects for a given resource type.
func Resources(c dynamic.Interface, r schema.GroupVersionResource, n string, a []string) ([]string, error) {
	list, err := c.Resource(r).Namespace(n).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var resources []string
	for _, resource := range list.Items {
		name, found, err := unstructured.NestedString(resource.Object, "metadata", "name")
		if err != nil || !found {
			return nil, err
		}

		if !stringContainsArrayElement(name, a) {
			resources = append(resources, name)
		}
	}

	return resources, nil
}

// ResourcesOlderThan returns a list of the resources older than the duration 'd'.
func ResourcesOlderThan(c dynamic.Interface, r schema.GroupVersionResource, n string, d time.Duration, a []string) ([]Resource, error) {
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

			if !stringContainsArrayElement(name, a) {
				resource = append(resource, Resource{
					Name: name,
					Kind: r.Resource,
					Age:  age.Round(time.Minute),
				})
			}
		}
	}

	return resource, nil
}

func DeleteResource(c dynamic.Interface, r schema.GroupVersionResource, ns, n string) error {
	deletePolicy := meta_v1.DeletePropagationForeground
	deleteOptions := meta_v1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	return c.Resource(r).Namespace(ns).Delete(context.TODO(), n, deleteOptions)
}

func stringContainsArrayElement(s string, t []string) bool {
	for _, e := range t {
		if strings.Contains(s, e) {
			return true
		}
	}
	return false
}
