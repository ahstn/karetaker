package kubernetes

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"testing"
	"time"
)

var (
	deployResource = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	configResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
)

func TestResourcesOlderThan(t *testing.T) {
	scheme := runtime.NewScheme()

	var tests = []struct {
		name string
		duration time.Duration
		expected []Resource
		client   dynamic.Interface
	}{
		{
			name: "returns two deployments older than the target duration",
			duration: 5 * time.Hour,
			expected: []Resource{
				{"eight-hours", "deployments", 5 * time.Hour},
				{"seventy-hours", "deployments", 70 * time.Hour},
			},
			client: fake.NewSimpleDynamicClient(scheme,
				newDeploymentWithTime("two-hours", time.Now().Add(-2 * time.Hour)),
				newDeploymentWithTime("eight-hours", time.Now().Add(-8 * time.Hour)),
				newDeploymentWithTime("seventy-hours", time.Now().Add(-70 * time.Hour)),
			),
		},
	}

	for _, test := range tests {
		t.Run("Test", func(t *testing.T) {
			actual, err := ResourcesOlderThan(test.client, deployResource, "default", test.duration)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(actual, test.expected); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
				return
			}
		})
	}
}

func TestResourcesInUse(t *testing.T) {
	scheme := runtime.NewScheme()

	configs := []string{"properties", "env-vars"}
	secrets := []string{"tokens"}

	client := fake.NewSimpleDynamicClient(scheme, newPodWithVolumes("config-pod", "properties", "tokens"))

	t.Run("returns used configmap and secret", func(t *testing.T) {
		actualConfigs, actualSecrets, err := ResourcesInUse(client, "default")
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
			return
		}

		if diff := cmp.Diff(actualSecrets, secrets); diff != "" {
			t.Errorf("Secrets differ, (-got +want): %s", diff)
		}

		if diff := cmp.Diff(actualConfigs, configs); diff != "" {
			t.Errorf("Configs differ, (-got +want): %s", diff)
		}
	})
}

func newResourceWithTime(api, kind, name string, t time.Time) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": api,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace": "default",
				"name":      name,
				"creationTimestamp": t.Format(time.RFC3339),
			},
		},
	}
}


// Adding to "fake" Kubernetes client fails on the following:
// DeepCopyJSON - cannot deep copy []map[string]interface
// TODO: GH Issue in kubernetes/client-go
func newPodWithVolumes(name, config, secret string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "pod",
			"metadata": map[string]interface{}{
				"namespace": "default",
				"name":      name,
				"creationTimestamp": time.Now().Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"containers": []map[string]interface{}{
					{
						"name": name,
					},
				},
			},
			"volumes": []map[string]interface{}{
				{
					"name": secret,
					"secret": map[string]interface{}{
						"secretName": secret,
					},
				},
				{
					"name": config,
					"configMap": map[string]interface{}{
						"name": config,
					},
				},
			},
		},
	}
}

// Same issue as above
func newPodWithConfigMapEnv(name, config string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "pod",
			"metadata": map[string]interface{}{
				"namespace": "default",
				"name":      name,
				"creationTimestamp": time.Now().Format(time.RFC3339),
			},
			"spec": map[string]interface{}{
				"containers": []map[string]interface{}{
					{
						"name": name,
						"envFrom": []map[string]interface{}{
							{
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

func newDeploymentWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("apps/v1", "deployment", name, t)
}

func newConfigmapWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("v1", "configmap", name, t)
}

func newSecretWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("v1", "secret", name, t)
}