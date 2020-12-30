package actions

import (
	"fmt"
	"github.com/ahstn/karetaker/pkg/kubernetes"
	"github.com/ahstn/karetaker/pkg/log"
	"github.com/thatisuday/commando"
	"os"
	"text/tabwriter"
)

func Duplicate(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	namespace, _ := flags["namespace"].GetString()
	filter, _ := flags["filter"].GetString()
	targetLabel := args["target"].Value

	s := log.Print("Connecting to Kubernetes Cluster")
	clientset, err := kubernetes.Config("")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	s.Stop()

	s = log.Print(fmt.Sprintf("Fetching Deployments (namespace: %s)", namespace))
	deployments, err := kubernetes.ListDuplicateDeployments(clientset, namespace, filter, targetLabel)
	s.Stop()

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	fmt.Fprintf(w, "%s\t%s\n", "DEPLOYMENT", "MATCHES")
	for deployment, matches := range deployments {
		fmt.Fprintf(w, "%s\t%v\t\n", deployment, matches)
	}
}
