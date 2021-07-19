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

// ResourcesInUse returns two maps of configmaps and secrets currently in use by existing pods.
// NB: This needs broken up to limit the amount of nesting.
// After retrieving the existing pods, it looks at containers.envFrom and volumes for configmap/secret references.
// Any found references are placed into the respective maps (with the key as their metadata.name)
func ResourcesInUse(c dynamic.Interface, n string) (map[string]bool, map[string]bool, error) {
	list, err := c.Resource(PodSchema).Namespace(n).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, nil, errors.Wrap(err, "error getting pods")
	}

	secrets := make(map[string]bool)
	configs := make(map[string]bool)

	for _, pod := range list.Items {
		containers, found, err := unstructured.NestedSlice(pod.Object, "spec", "containers")
		if err != nil || !found {
			return nil, nil, err
		}

		for _, container := range containers {
			envs, found, err := unstructured.NestedSlice(container.(map[string]interface{}), "envFrom")
			if err != nil {
				return nil, nil, err
			} else if found {
				for _, env := range envs {
					config, found, err := unstructured.NestedString(env.(map[string]interface{}), "configMapRef", "name")
					if err != nil {
						return nil, nil, err
					} else if found {
						configs[config] = true
					}
				}
			}
		}

		volumes, found, err := unstructured.NestedSlice(pod.Object, "spec", "volumes")
		if err != nil {
			return nil, nil, err
		} else if !found {
			continue
		}

		for _, volume := range volumes {
			config, found, err := unstructured.NestedString(volume.(map[string]interface{}), "configMap", "name")
			if err != nil {
				return nil, nil, err
			} else if found {
				configs[config] = true
			}

			secret, found, err := unstructured.NestedString(volume.(map[string]interface{}), "secret", "secretName")
			if err != nil {
				return nil, nil, err
			} else if found {
				secrets[secret] = true
			}
		}
	}
	return configs, secrets, nil
}

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
