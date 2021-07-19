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
)

func TestResourcesOlderThan(t *testing.T) {
	scheme := runtime.NewScheme()

	var tests = []struct {
		name     string
		duration time.Duration
		allow    []string
		client   dynamic.Interface
		expected []Resource
	}{
		{
			name:     "returns deployments older than the target duration with no allow-list",
			duration: 5 * time.Hour,
			allow: []string{},
			client: fake.NewSimpleDynamicClient(scheme,
				newDeploymentWithTime("two-hours", time.Now().Add(-2*time.Hour)),
				newDeploymentWithTime("eight-hours", time.Now().Add(-8*time.Hour)),
				newDeploymentWithTime("seventy-hours", time.Now().Add(-70*time.Hour)),
			),
			expected: []Resource{
				{Name: "eight-hours", Kind: "deployments", Age: 8 * time.Hour},
				{Name: "seventy-hours", Kind: "deployments", Age: 70 * time.Hour},
			},
		},
		{
			name:     "returns deployments older than the target duration with allow-list",
			duration: 5 * time.Hour,
			allow: []string{"certs", "monitoring"},
			client: fake.NewSimpleDynamicClient(scheme,
				newDeploymentWithTime("seventy-hours", time.Now().Add(-70*time.Hour)),
				newDeploymentWithTime("certs-controller", time.Now().Add(-120*time.Hour)),
				newDeploymentWithTime("monitoring", time.Now().Add(-120*time.Hour)),
			),
			expected: []Resource{
				{Name: "seventy-hours", Kind: "deployments", Age: 70 * time.Hour},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ResourcesOlderThan(test.client, deployResource, "default", test.duration, test.allow)
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

func TestResources(t *testing.T) {
	scheme := runtime.NewScheme()

	tests := []struct {
		name     string
		resource schema.GroupVersionResource
		allow    []string
		client   dynamic.Interface
		expected []string
	}{
		{
			name:     "Testing ConfigMaps with complete match allow-list",
			resource: ConfigMapSchema,
			allow:    []string{"allowed-config"},
			client: fake.NewSimpleDynamicClient(scheme,
				newConfigmap("properties"),
				newConfigmap("env-vars"),
				newConfigmap("allowed-config"),
			),
			expected: []string{"properties", "env-vars"},
		},
		{
			name:     "Testing Secrets with contains match allow-list",
			resource: SecretSchema,
			allow:    []string{"allowed"},
			client: fake.NewSimpleDynamicClient(scheme,
				newSecret("tokens"),
				newSecret("certs"),
				newSecret("allowed-secret"),
			),
			expected: []string{"tokens", "certs"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := Resources(test.client, test.resource, "default", test.allow)
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

func TestDeleteResource(t *testing.T) {
	scheme := runtime.NewScheme()

	client := fake.NewSimpleDynamicClient(scheme,
		newConfigmap("unused-config"),
		newSecret("unused-secret"),
	)

	err := DeleteResource(client, ConfigMapSchema, "default", "unused-config")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	err = DeleteResource(client, SecretSchema, "default", "unused-secret")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	err = DeleteResource(client, PodSchema, "default", "invalid-should-err")
	if err == nil {
		t.Errorf("Expected error, but got: %s", err)
	}
}

func newResource(api, kind, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": api,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace":         "default",
				"name":              name,
				"creationTimestamp": time.Now().Format(time.RFC3339),
			},
		},
	}
}

func newResourceWithTime(api, kind, name string, t time.Time) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": api,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"namespace":         "default",
				"name":              name,
				"creationTimestamp": t.Format(time.RFC3339),
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


