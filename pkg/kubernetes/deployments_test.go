package kubernetes

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/dynamic/fake"
)

func newDeploymentWithTime(name string, t time.Time) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"namespace": "default",
				"name":      name,
				"creationTimestamp": t.Format(time.RFC3339),
			},
		},
	}
}

func TestListDeploymentsOlderThan(t *testing.T) {
	scheme := runtime.NewScheme()

	var tests = []struct {
		duration time.Duration
		expected []Deployment
		client   dynamic.Interface
	}{
		{
			duration: 5 * time.Hour,
			expected: []Deployment{
				{"eight-hours", 5 * time.Hour},
				{"seventy-hours", 70 * time.Hour},
			},
			client: fake.NewSimpleDynamicClient(scheme,
				newDeploymentWithTime("two-hours", time.Now().Add(-2 * time.Hour)),
				newDeploymentWithTime("eight-hours", time.Now().Add(-5 * time.Hour)),
				newDeploymentWithTime("seventy-hours", time.Now().Add(-70 * time.Hour)),
			),
		},
	}

	for _, test := range tests {
		t.Run("Test", func(t *testing.T) {
			actual, err := ListDeploymentsOlderThan(test.client, "default", test.duration)
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
//
//func TestListDuplicateDeployments(t *testing.T) {
//	var tests = []struct {
//		namespace     string
//		appLabel      string
//		instanceLabel string
//		expected      map[string][]string
//		client     kubernetes.Interface
//	}{
//		{
//			namespace:     "default",
//			appLabel:      "kubernetes.io/name=auth",
//			instanceLabel: "kubernetes.io/instance",
//			expected: map[string][]string{
//				"dev": {"dev1"},
//				"qa1": {"qa2", "qa3"},
//			},
//			client: fake.NewSimpleClientset(&appsv1.Deployment{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      "test-dev",
//					Namespace: "default",
//					Labels: map[string]string{
//						"kubernetes.io/name":     "auth",
//						"kubernetes.io/instance": "dev",
//					},
//				},
//			}, &appsv1.Deployment{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      "test-dev1",
//					Namespace: "default",
//					Labels: map[string]string{
//						"kubernetes.io/name":     "auth",
//						"kubernetes.io/instance": "dev1",
//					},
//				},
//			}, &appsv1.Deployment{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      "test-false",
//					Namespace: "default",
//					Labels: map[string]string{
//						"kubernetes.io/name":     "auth",
//						"kubernetes.io/instance": "false",
//					},
//				},
//			}, &appsv1.Deployment{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      "test-qa1",
//					Namespace: "default",
//					Labels: map[string]string{
//						"kubernetes.io/name":     "auth",
//						"kubernetes.io/instance": "qa1",
//					},
//				},
//			}, &appsv1.Deployment{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      "test-qa2",
//					Namespace: "default",
//					Labels: map[string]string{
//						"kubernetes.io/name":     "auth",
//						"kubernetes.io/instance": "qa2",
//					},
//				},
//			}, &appsv1.Deployment{
//				ObjectMeta: metav1.ObjectMeta{
//					Name:      "test-qa3",
//					Namespace: "default",
//					Labels: map[string]string{
//						"kubernetes.io/name":     "auth",
//						"kubernetes.io/instance": "qa3",
//					},
//				},
//			}, &v1.Namespace{
//				ObjectMeta: metav1.ObjectMeta{
//					Name: "default",
//				},
//			}),
//		},
//	}
//
//	for _, test := range tests {
//		t.Run("Test", func(t *testing.T) {
//			actual, err := ListDuplicateDeployments(test.client, test.namespace, test.appLabel, test.instanceLabel)
//			if err != nil {
//				t.Errorf("Unexpected error: %s", err)
//				return
//			}
//			if diff := cmp.Diff(actual, test.expected); diff != "" {
//				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
//				return
//			}
//		})
//	}
//}

func namespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: name},
		Spec:       v1.NamespaceSpec{},
	}
}
