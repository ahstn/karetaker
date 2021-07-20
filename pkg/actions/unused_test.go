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

const (
	unused         = "unused-config"
	usedConfigName = "properties-config"
	usedSecretName = "tokens-secret"
	failedJob      = "failed-job"
	completedJob   = "completed-job"
)

var (
	defaultUnusedObjects = []runtime.Object{
		newPodWithVolumes("config-pod", usedConfigName, usedSecretName),
		newConfigmap(unused),
		newConfigmap(usedConfigName),
		newSecret(usedSecretName),
		newFailedJob(failedJob),
		newCompletedJobWithTime(completedJob, time.Now().Add(-1*time.Hour)),
	}
)

func TestUnusedLogOutputAndDeletion(t *testing.T) {
	tests := []struct {
		name      string
		config    domain.Unused
		expected  []string
		remaining int
	}{
		{
			name: "Error is printed on invalid resource type",
			config: domain.Unused{
				Resources: []string{"invalid-resource"},
				Namespace: "default",
				Allow:     []string{},
				DryRun:    false,
			},
			expected: []string{
				"Unsupported resource: invalid-resource, skipping.",
			},
		},
		{
			name: "On dry-run, objects are printed and not deleted",
			config: domain.Unused{
				Resources: []string{"configmap", "job"},
				Namespace: "default",
				Allow:     []string{},
				DryRun:    true,
			},
			expected: []string{
				fmt.Sprintf("%s\tUN-CHANGED (dry-run)", unused),
				fmt.Sprintf("%s\tIN-USE", usedConfigName),
				fmt.Sprintf("%s\tUN-CHANGED (dry-run)", failedJob),
				fmt.Sprintf("%s\tUN-CHANGED (dry-run)", completedJob),
			},
		},
		{
			name: "With age filter, certain objects are skipped",
			config: domain.Unused{
				Resources: []string{"configmap", "job"},
				Namespace: "default",
				Age:       24 * time.Hour,
				Allow:     []string{},
				DryRun:    false,
			},
			expected: []string{
				fmt.Sprintf("%s\tUN-CHANGED (age)", completedJob),
			},
		},
		{
			name: "Objects are printed and deleted",
			config: domain.Unused{
				Resources: []string{"configmap", "job"},
				Namespace: "default",
				Allow:     []string{},
				DryRun:    false,
			},
			expected: []string{
				fmt.Sprintf("%s\tDELETED", unused),
				fmt.Sprintf("%s\tIN-USE", usedConfigName),
				fmt.Sprintf("%s\tDELETED (was: 2)", failedJob),
			},
		},
	}
	for _, tt := range tests {
		client := fake.NewSimpleDynamicClient(defaultScheme, defaultUnusedObjects...)

		t.Run(tt.name, func(t *testing.T) {
			o := &bytes.Buffer{}
			err := Unused(client, tt.config, o)
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

func newCompletedJob(name string) *unstructured.Unstructured {
	return newJobWithStatus(name, map[string]interface{}{
		"succeeded": int64(1),
	})
}

func newCompletedJobWithTime(name string, t time.Time) *unstructured.Unstructured {
	return newJobWithStatusAndTime(name, map[string]interface{}{
		"succeeded": int64(1),
	}, t)
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

func newJobWithStatusAndTime(name string, status map[string]interface{}, t time.Time) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "batch/v1",
			"kind":       "job",
			"metadata": map[string]interface{}{
				"namespace":         "default",
				"name":              name,
				"creationTimestamp": t.Format(time.RFC3339),
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
