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

	now := time.Now()
	var resource []Resource
	for _, job := range list.Items {
		age, err := determineObjectAge(job, now)
		if err != nil {
			return nil, err
		}

		name, found, err := unstructured.NestedString(job.Object, "metadata", "name")
		if err != nil || !found {
			return nil, err
		}

		// TODO: Extract this, enum status and potentially cover pod statuses also
		var status string
		succeeded, found, err := unstructured.NestedInt64(job.Object, "status", "succeeded")
		if err != nil {
			return nil, err
		} else if !found {
			failed, found, err := unstructured.NestedInt64(job.Object, "status", "failed")
			if err != nil {
				return nil, err
			} else if !found {
				break
			} else if failed == 1 {
				status = "failed"
			}
		} else if succeeded == 1 {
			status = "success"
		}

		if !stringContainsArrayElement(name, a) {
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

func determineObjectAge(job unstructured.Unstructured, now time.Time) (time.Duration, error) {
	t, found, err := unstructured.NestedString(job.Object, "metadata", "creationTimestamp")
	if err != nil || !found {
		return 0, err
	}

	creation, err := time.Parse(time.RFC3339, t)
	return now.Sub(creation), nil
}
