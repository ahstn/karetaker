package kubernetes

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	fake "k8s.io/client-go/kubernetes/fake"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func TestListImages(t *testing.T) {
	var tests = []struct {
		expected  []string
		clientset kubernetes.Interface
	}{
		{
			expected: []string{"default", "billing"},
			clientset: fake.NewSimpleClientset(&v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-pod",
				},
			}, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			}, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "billing",
				},
			}),
		},
	}

	for _, test := range tests {
		t.Run("Test", func(t *testing.T) {
			actual, err := ListNamespaces(test.clientset)
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

func namespace(name string) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: name},
		Spec:       v1.NamespaceSpec{},
	}
}
