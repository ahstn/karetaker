package kubernetes

import (
	"context"
	"github.com/pkg/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"time"
)

//
func JobsNotRunning(c dynamic.Interface, n string, a []string) ([]Resource, error) {
	list, err := c.Resource(JobSchema).Namespace(n).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "getting resource")
	}

	var resource []Resource
	for _, job := range list.Items {
		age, err := objectAge(job)
		if err != nil {
			return nil, err
		}

		name, found, err := unstructured.NestedString(job.Object, "metadata", "name")
		if err != nil || !found {
			return nil, err
		}

		status, err := objectStatus(job)
		if err != nil {
			return nil, err
		}

		if !stringContainsArrayElement(name, a) && (status == Completed || status == Failed) {
			resource = append(resource, Resource{
				Name:   name,
				Kind:   JobSchema.Resource,
				Age:    age.Round(time.Minute),
				Status: status,
			})
		}
	}

	return resource, nil
}

// TODO: Test dis
// TODO: Cover deploy status also
func objectStatus(job unstructured.Unstructured) (Status, error) {
	var status Status
	succeeded, found, err := unstructured.NestedInt64(job.Object, "status", "succeeded")
	if err != nil {
		return Unknown, err
	} else if succeeded == 1 {
		status = Completed
	} else if !found {
		failed, _, err := unstructured.NestedInt64(job.Object, "status", "failed")
		if err != nil {
			return Unknown, err
		} else if failed == 1 {
			status = Failed
		} else {
			status = Unknown
		}
	}
	return status, nil
}

// TODO: Test dis
func objectAge(job unstructured.Unstructured) (time.Duration, error) {
	t, found, err := unstructured.NestedString(job.Object, "metadata", "creationTimestamp")
	if err != nil || !found {
		return 0, err
	}

	creation, err := time.Parse(time.RFC3339, t)
	return time.Now().Sub(creation), nil
}
