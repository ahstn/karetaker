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

func TestUnusedLogOutputAndDeletion(t *testing.T) {
	unused := "unused-config"
	scheme := runtime.NewScheme()
	client := fake.NewSimpleDynamicClient(scheme,
		newPodWithVolumes("config-pod", "properties", "tokens"),
		newConfigmap(unused),
		newConfigmap("properties"),
	)

	tests := []struct {
		name      string
		config    domain.Unused
		expected  []string
		remaining int
	}{
		{
			name: "On dry-run, objects are printed and not deleted",
			config: domain.Unused{
				Resources: []string{"configmap"},
				Namespace: "default",
				Allow:     []string{},
				DryRun:    true,
			},
			expected: []string{
				fmt.Sprintf("%s\tUN-CHANGED (dry-run)", unused),
				fmt.Sprintf("properties\tIN-USE"),
			},
			remaining: 2,
		},
		{
			name: "Objects are printed and deleted",
			config: domain.Unused{
				Resources: []string{"configmap"},
				Namespace: "default",
				Allow:     []string{},
				DryRun: false,
			},
			expected: []string{
				fmt.Sprintf("%s\tDELETED", unused),
				fmt.Sprintf("properties\tIN-USE"),
			},
			remaining: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &bytes.Buffer{}
			err := Unused(client, tt.config, o)
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

			result, _ := client.Resource(kubernetes.ConfigMapSchema).Namespace("default").List(context.TODO(), meta_v1.ListOptions{})
			if len(result.Items) != tt.remaining {
				t.Errorf("expected number of configmaps is %d, got: %d", tt.remaining, len(result.Items))
				return
			}
		})
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

func newConfigmap(name string) *unstructured.Unstructured {
	return newResource("v1", "configmap", name)
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
