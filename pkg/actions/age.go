package actions

import (
	"fmt"
	"github.com/ahstn/karetaker/pkg/domain"
	"github.com/ahstn/karetaker/pkg/kubernetes"
	"io"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Age for each resource type in 'u.Resources', find objects older than 'u.Age' and delete them.
func Age(c dynamic.Interface, u domain.Age, o io.Writer) error {
	for _, resource := range u.Resources {
		var gvr schema.GroupVersionResource

		switch resource {
		case "configmap","configmaps":
			gvr = kubernetes.ConfigMapSchema
		case "secret","secrets":
			gvr = kubernetes.SecretSchema
		case "deploy","deployment","deployments":
			gvr = kubernetes.DeploymentSchema
		default:
			fmt.Fprintf(o, "Unsupported resource: %s, skipping.", resource)
			continue
		}

		list, err := kubernetes.ResourcesOlderThan(c, gvr, u.Namespace, u.Age, u.Allow)
		if err != nil {
			return err
		}

		fmt.Fprint(o, "RESOURCE\tAGE\tSTATUS\n")
		for _, item := range list {
			if u.DryRun {
				fmt.Fprintf(o, "%s\t%v\tUN-CHANGED (dry-run)\n", item.Name, item.Age)
			} else {
				fmt.Fprintf(o, "%s\t%v\tDELETED\n", item.Name, item.Age)
			}
		}
	}
}
