package kubernetes

import (
	"context"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// UsedConfigAndSecrets returns two maps of configmaps and secrets currently in use by existing pods.
// NB: This needs broken up to limit the amount of nesting.
// After retrieving the existing pods, it looks at containers.envFrom and volumes for configmap/secret references.
// Any found references are placed into the respective maps (with the key as their metadata.name)
func UsedConfigAndSecrets(c dynamic.Interface, n string) (map[string]bool, map[string]bool, error) {
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