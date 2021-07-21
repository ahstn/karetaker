package actions

import (
	"fmt"
	"github.com/ahstn/karetaker/pkg/domain"
	"github.com/ahstn/karetaker/pkg/kubernetes"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Unused retrieves the resources in use (i.e. referenced configmaps) and all the existing resources.
// It then cross-references those to determine which are not currently in use.
func Unused(c dynamic.Interface, u domain.Unused, o io.Writer) error {
	for _, resource := range u.Resources {
		var err error
		switch resource {
		case "configmap","configmaps","secret","secrets":
			err = handleConfigSecrets(c, u, o)
		case "job","jobs":
			err = handleJobs(c, u, o)
		default:
			fmt.Fprintf(o, "Unsupported resource: %s, skipping.", resource)
			continue
		}

		if err != nil {
			fmt.Printf("error executing for resource type (%s), continuing...", resource)
		}
	}

	return nil
}

func handleConfigSecrets(c dynamic.Interface, u domain.Unused, o io.Writer) error {
	usedConfigs, usedSecrets, err := kubernetes.UsedConfigAndSecrets(c, u.Namespace)
	if err != nil {
		return err
	}
	for _, resource := range u.Resources {
		var gvr schema.GroupVersionResource
		var ref map[string]bool

		switch resource {
		case "configmap", "configmaps":
			gvr = kubernetes.ConfigMapSchema
			ref = usedConfigs
		case "secret", "secrets":
			gvr = kubernetes.SecretSchema
			ref = usedSecrets
		default:
			return nil
		}

		list, err := kubernetes.Resources(c, gvr, u.Namespace, u.Allow)
		if err != nil {
			return err
		}

		fmt.Fprintf(o, "RESOURCE (%s)\tSTATUS\n", resource)
		for _, item := range list {
			if _, isPresent := ref[item]; isPresent {
				fmt.Fprintf(o, "%s\tIN-USE\t\n", item)
			} else if u.DryRun {
				fmt.Fprintf(o, "%s\tUN-CHANGED (dry-run)\t\n", item)
			} else {
				fmt.Fprintf(o, "%s\tDELETED\t\n", item)
				err = kubernetes.DeleteResource(c, gvr, u.Namespace, item)
				if err != nil {
					fmt.Printf("error deleting %s, continuing...", item)
				}
			}
		}
	}
	return nil
}

func handleJobs(c dynamic.Interface, u domain.Unused, o io.Writer) error {
	jobs, err := kubernetes.JobsNotRunning(c, u.Namespace, u.Allow)
	if err != nil {
		return err
	}

	fmt.Fprintf(o, "RESOURCE (jobs)\tSTATUS\n")
	for _, job := range jobs {
		if u.DryRun {
			fmt.Fprintf(o, "%s\tUN-CHANGED (dry-run)\t\n", job.Name)
		} else if u.Age != 0 && (job.Age < u.Age) {
			fmt.Fprintf(o, "%s\tUN-CHANGED (age)\t\n", job.Name)
		}else {
			fmt.Fprintf(o, "%s\tDELETED (was: %v)\t\n", job.Name, job.Status)
			err = kubernetes.DeleteResource(c, kubernetes.JobSchema, u.Namespace, job.Name)
			if err != nil {
				fmt.Printf("error deleting %s, continuing...", job.Name)
			}
		}
	}
	return nil
}