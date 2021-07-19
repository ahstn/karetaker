package kubernetes

import (
	"github.com/google/go-cmp/cmp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"testing"
	"time"
)

func TestJobsNotRunning(t *testing.T) {
	scheme := runtime.NewScheme()

	tests := []struct {
		name     string
		allow    []string
		client   dynamic.Interface
		expected []Resource
	}{
		{
			name:     "Testing ConfigMaps with complete match allow-list",
			allow:    []string{"allowed-job"},
			client: fake.NewSimpleDynamicClient(scheme,
				newCompletedJob("completed-job"),
				newCompletedJob("allowed-job"),
				newFailedJob("failed-job"),
				newRunningJob("running-job"),
			),
			expected: []Resource{
				{Name: "completed-job", Kind: "jobs", Age: 0, Status: Completed},
				{Name: "failed-job", Kind: "jobs", Age: 0, Status: Failed},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := JobsNotRunning(test.client, "default", test.allow)
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


func newCompletedJob(name string) *unstructured.Unstructured {
	return newJobWithStatus(name, map[string]interface{}{
		"succeeded": int64(1),
	})
}

func newRunningJob(name string) *unstructured.Unstructured {
	return newJobWithStatus(name, map[string]interface{}{
		"running": int64(1),
	})
}

func newFailedJob(name string) *unstructured.Unstructured {
	return newJobWithStatus(name, map[string]interface{}{
		"failed": int64(1),
	})
}

func newJobWithStatus(name string, status map[string]interface{}) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "job",
			"metadata": map[string]interface{}{
				"namespace":         "default",
				"name":              name,
				"creationTimestamp": time.Now().Format(time.RFC3339),
			},
			"status": status,
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name": name,
					},
				},
			},
		},
	}
}