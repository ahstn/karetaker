package actions

import (
	"bytes"
	"fmt"
	"github.com/ahstn/karetaker/pkg/domain"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	"strings"
	"testing"
	"time"
)

var (
	defaultScheme = runtime.NewScheme()
	defaultObjects = []runtime.Object {
		newDeploymentWithTime("two-hours-deploy", time.Now().Add(-2*time.Hour)),
		newDeploymentWithTime("eight-hours-deploy", time.Now().Add(-8*time.Hour)),
		newDeploymentWithTime("seventy-hours-deploy", time.Now().Add(-70*time.Hour)),
		newStatefulSetWithTime("seventy-hours-ss", time.Now().Add(-70*time.Hour)),
		newJobWithTime("seventy-hours-job", time.Now().Add(-70*time.Hour)),
		newServiceWithTime("seventy-hours-svc", time.Now().Add(-70*time.Hour)),
		newConfigMapWithTime("seventy-hours-cm", time.Now().Add(-70*time.Hour)),
		newSecretWithTime("seventy-hours-secret", time.Now().Add(-70*time.Hour)),
	}
)

func TestAgeLogOutputAndDeletionUsingDeployments(t *testing.T) {
	tests := []struct {
		name      string
		config    domain.Age
		expected  []string
		remaining int
	}{
		{
			name: "Error is printed on invalid resource type",
			config: domain.Age{
				Resources: []string{"invalid-resource"},
				Namespace: "default",
				Age:       5 * time.Hour,
				Allow:     []string{},
				DryRun:    false,
			},
			expected: []string{
				"Unsupported resource: invalid-resource, skipping.",
			},
		},
		{
			name: "On dry-run, objects are printed and not deleted",
			config: domain.Age{
				Resources: []string{"deployment"},
				Namespace: "default",
				Age:       5 * time.Hour,
				Allow:     []string{},
				DryRun:    true,
			},
			expected: []string{
				fmt.Sprintf("eight-hours-deploy\t8h0m0s\tUN-CHANGED (dry-run)"),
				fmt.Sprintf("seventy-hours-deploy\t70h0m0s\tUN-CHANGED (dry-run)"),
			},
		},
		{
			name: "Objects are printed and deleted",
			config: domain.Age{
				Resources: []string{"deployment"},
				Namespace: "default",
				Age:       5 * time.Hour,
				Allow:     []string{},
				DryRun:    false,
			},
			expected: []string{
				fmt.Sprintf("eight-hours-deploy\t8h0m0s\tDELETED"),
				fmt.Sprintf("seventy-hours-deploy\t70h0m0s\tDELETED"),
			},
		},
		{
			name: "Deletes multiple resource types",
			config: domain.Age{
				Resources: []string{"deploy","svc","ss","job","configmap","secret"},
				Namespace: "default",
				Age:       5 * time.Hour,
				Allow:     []string{},
				DryRun:    false,
			},
			expected: []string{
				fmt.Sprintf("eight-hours-deploy\t8h0m0s\tDELETED"),
				fmt.Sprintf("seventy-hours-deploy\t70h0m0s\tDELETED"),
				fmt.Sprintf("seventy-hours-job\t70h0m0s\tDELETED"),
				fmt.Sprintf("seventy-hours-ss\t70h0m0s\tDELETED"),
				fmt.Sprintf("seventy-hours-cm\t70h0m0s\tDELETED"),
				fmt.Sprintf("seventy-hours-secret\t70h0m0s\tDELETED"),
			},
		},
	}
	for _, tt := range tests {
		client := fake.NewSimpleDynamicClient(defaultScheme, defaultObjects...)

		t.Run(tt.name, func(t *testing.T) {
			o := &bytes.Buffer{}
			err := Age(client, tt.config, o)
			if err != nil {
				t.Errorf("Unused() error = %v", err)
				return
			}

			for _, expected := range tt.expected {
				if !strings.Contains(o.String(), expected) {
					t.Errorf("Output error, \nexpected: %s \ngot: %s", expected, o.String())
					return
				}
			}
		})
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

func newStatefulSetWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("apps/v1", "statefulset", name, t)
}

func newServiceWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("v1", "service", name, t)
}

func newJobWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("batch/v1", "job", name, t)
}

func newConfigMapWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("v1", "configmap", name, t)
}

func newSecretWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newResourceWithTime("v1", "secret", name, t)
}


