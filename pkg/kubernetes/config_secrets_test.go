package kubernetes

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"testing"
	"time"
)

var (
	configResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
)

func TestResourcesInUse(t *testing.T) {
	scheme := runtime.NewScheme()
	expectedConfigs := map[string]bool{"properties": true, "env-vars": true}
	expectedSecrets := map[string]bool{"tokens": true}

	client := fake.NewSimpleDynamicClient(scheme,
		newPodWithVolumes("config-pod", "properties", "tokens"),
		newPodWithConfigMapEnv("env-pod", "env-vars"),
		newConfigmap("unused-config"),
		newSecret("unused-secret"),
	)

	configs, secrets, err := UsedConfigAndSecrets(client, "default")
	if err != nil {
		t.Error(err)
	}

	if diff := cmp.Diff(configs, expectedConfigs); diff != "" {
		t.Errorf("%T differ (-got, +want): %s", []string{"properties"}, diff)
		return
	}

	if diff := cmp.Diff(secrets, expectedSecrets); diff != "" {
		t.Errorf("%T differ (-got, +want): %s", []string{"properties"}, diff)
		return
	}

}

func newConfigmap(name string) *unstructured.Unstructured {
	return newResource("v1", "configmap", name)
}

func newSecret(name string) *unstructured.Unstructured {
	return newResource("v1", "secret", name)
}

func newPodWithVolumes(name, config, secret string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "pod",
			"metadata": map[string]interface{}{
				"namespace": "default",
				"name":      name,
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name": name,
					},
				},
				"volumes": []interface{}{
					map[string]interface{}{
						"name": secret,
						"secret": map[string]interface{}{
							"secretName": secret,
						},
					},
					map[string]interface{}{
						"name": config,
						"configMap": map[string]interface{}{
							"name": config,
						},
					},
				},
			},
		},
	}
}

func newPodWithConfigMapEnv(name, config string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "pod",
			"metadata": map[string]interface{}{
				"namespace":         "default",
				"name":              name,
				"creationTimestamp": time.Now().Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name": name,
						"envFrom": []interface{}{
							map[string]interface{}{
								"configMapRef": map[string]interface{}{
									"name": config,
								},
							},
						},
					},
				},
			},
		},
	}
}