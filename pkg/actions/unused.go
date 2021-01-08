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
	usedConfigs, usedSecrets, err := kubernetes.ResourcesInUse(c, u.Namespace)
	if err != nil {
		return err
	}

	for _, resource := range u.Resources {
		var gvr schema.GroupVersionResource
		var ref map[string]bool

		switch resource {
		case "configmap","configmaps":
			gvr = kubernetes.ConfigMapSchema
			ref = usedConfigs
		case "secret","secrets":
			gvr = kubernetes.SecretSchema
			ref = usedSecrets
		default:
			fmt.Fprintf(o, "Unsupported resource: %s, skipping.", resource)
			continue
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
