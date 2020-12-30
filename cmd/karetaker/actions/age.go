package actions

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ahstn/karetaker/pkg/kubernetes"
	"github.com/ahstn/karetaker/pkg/log"
	"github.com/thatisuday/commando"
)

var allowlist = []string{"default-token", "istio-ca", "sh.helm.release"}

func Age(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	r := schema.GroupVersionResource{}
	n, _ := flags["namespace"].GetString()
	a, _ := flags["age"].GetString()
	al, _ := flags["allow"].GetString()
	allowlist = append(allowlist, strings.Split(al, ",")[:]...)

	age, err := time.ParseDuration(a)
	if err != nil {
		panic("Unsupported Duration")
	}

	resource := args["type"].Value

	if resource == "deployment" {
		r = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	} else if resource == "configmap" {
		r = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	} else {
		panic("Unsupported resource")
	}

	s := log.Print("Connecting to Kubernetes Cluster")
	client, err := kubernetes.DynamicConfig("")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	s.Stop()

	s = log.Print(fmt.Sprintf("Fetching Deployments (namespace: %s)", n))
	resources, err := kubernetes.ResourcesOlderThan(client, r, n, age, allowlist)
	s.Stop()

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	fmt.Fprintf(w, "%s\t%s\n", "RESOURCE", "AGE")
	for _, resource := range resources {
		fmt.Fprintf(w, "%s\t%v\t\n", resource.Name, resource.Age)
	}
}
