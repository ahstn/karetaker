package actions

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ahstn/karetaker/pkg/domain"
	"github.com/ahstn/karetaker/pkg/kubernetes"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	"strings"
	"testing"
	"time"
)

func TestAgeLogOutputAndDeletion(t *testing.T) {
	scheme := runtime.NewScheme()
	client := fake.NewSimpleDynamicClient(scheme,
		newDeploymentWithTime("two-hours", time.Now().Add(-2*time.Hour)),
		newDeploymentWithTime("eight-hours", time.Now().Add(-8*time.Hour)),
		newDeploymentWithTime("seventy-hours", time.Now().Add(-70*time.Hour)),
	)

	tests := []struct {
		name      string
		config    domain.Age
		expected  []string
		remaining int
	}{
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
				fmt.Sprintf("eight-hours\t8h0m0s\tUN-CHANGED (dry-run)"),
				fmt.Sprintf("seventy-hours\t70h0m0s\tUN-CHANGED (dry-run)"),
			},
			remaining: 3,
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
				fmt.Sprintf("eight-hours\t8h0m0s\tDELETED"),
				fmt.Sprintf("seventy-hours\t70h0m0s\tDELETED"),
			},
			remaining: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &bytes.Buffer{}
			err := Age(client, tt.config, o)
			if err != nil {
				t.Errorf("Unused() error = %v", err)
				return
			}

			for _, string := range tt.expected {
				if !strings.Contains(o.String(), string) {
					t.Errorf("Output error, \nexpected: %s \ngot: %s", string, o.String())
					return
				}
			}

			result, _ := client.Resource(kubernetes.DeploymentSchema).Namespace("default").List(context.TODO(), meta_v1.ListOptions{})
			if len(result.Items) != tt.remaining {
				t.Errorf("expected number of configmaps is %d, got: %d", tt.remaining, len(result.Items))
				return
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
