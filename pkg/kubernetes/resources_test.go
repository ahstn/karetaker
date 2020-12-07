package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	fake "k8s.io/client-go/dynamic/fake"
)

var (
	deployResource = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	configResource = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
)

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

func newDeploymentWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("apps/v1", "deployment", name, t)
}

func newConfigmapWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("", "configmap", name, t)
}

func TestResourcesOlderThan(t *testing.T) {
	scheme := runtime.NewScheme()

	var tests = []struct {
		duration time.Duration
		expected []Resource
		client   dynamic.Interface
	}{
		{
			duration: 5 * time.Hour,
			expected: []Resource{
				{"eight-hours", "deployments", 5 * time.Hour},
				{"seventy-hours", "deployments", 70 * time.Hour},
			},
			client: fake.NewSimpleDynamicClient(scheme,
				newDeploymentWithTime("two-hours", time.Now().Add(-2 * time.Hour)),
				newResourceWithTime("apps/v1", "deployment", "eight-hours", time.Now().Add(-5 * time.Hour)),
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