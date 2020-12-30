package actions

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ahstn/karetaker/pkg/kubernetes"
	"github.com/ahstn/karetaker/pkg/log"
	"github.com/thatisuday/commando"
)

func Unused(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	n, _ := flags["namespace"].GetString()
	d, _ := flags["dry-run"].GetBool()
	a, _ := flags["allow"].GetString()
	allowlist = append(allowlist, strings.Split(a, ",")[:]...)

	fmt.Printf("Using Allow List of: %s\n\n", allowlist)

	s := log.Print("Connecting to Kubernetes Cluster")
	client, err := kubernetes.DynamicConfig("")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	s.Stop()

	s = log.Print(fmt.Sprintf("Fetching Deployments (namespace: %s)", n))
	usedConfigs, usedSecrets, err := kubernetes.ResourcesInUse(client, n)
	s.Stop()

	fmt.Printf("Configs in use: %v\n", usedConfigs)
	fmt.Printf("Secrets in use: %v\n", usedSecrets)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()


	configs, err := kubernetes.Resources(client, kubernetes.ConfigMapSchema, n, allowlist)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "%s\t%s\n", "RESOURCE (CONFIGMAP)", "STATUS")
	for _, config := range configs {
		if _, isPresent := usedConfigs[config]; isPresent {
			fmt.Fprintf(w, "%s\tIN-USE\t\n", config)
		} else if d {
			fmt.Fprintf(w, "%s\tUN-CHANGED (dry-run)\t\n", config)
		} else {
			fmt.Fprintf(w, "%s\tDELETED\t\n", config)
		}
	}

	secrets, err := kubernetes.Resources(client, kubernetes.ConfigMapSchema, n, allowlist)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(w, "%s\t%s\n", "RESOURCE (SECRETS)", "STATUS")
	for _, secret := range secrets {
		if _, isPresent := usedSecrets[secret]; isPresent {
			fmt.Fprintf(w, "%s\tIN-USE\t\n", secret)
		} else if d {
			fmt.Fprintf(w, "%s\tUN-CHANGED (dry-run)\t\n", secret)
		} else {
			fmt.Fprintf(w, "%s\tDELETED\t\n", secret)
		}
	}

}