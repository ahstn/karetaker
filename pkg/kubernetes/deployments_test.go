package kubernetes

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
