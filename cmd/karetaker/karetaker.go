package main

import (
	"fmt"
	"github.com/ahstn/karetaker/cmd/karetaker/actions"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ahstn/karetaker/pkg/kubernetes"
	"github.com/ahstn/karetaker/pkg/log"
	"github.com/thatisuday/commando"
)

var questions = []*survey.Question{
	{
		Name: "namespace",
		Prompt: &survey.Select{
			Message: "Which Namespace?",
			Options: []string{"authentication", "billing", "infrastucture", "all"},
		},
		Validate: survey.Required,
	},
	{
		Name: "resources",
		Prompt: &survey.MultiSelect{
			Message: "Which Kubernetes Resources?",
			Options: []string{
				"Deployments",
				"Jobs",
				"Services",
				"StatefulSets",
			},
		},
	},
	{
		Name: "age",
		Prompt: &survey.Input{
			Message: "Age at which resources should be deleted? (i.e. 5d, 72hr)",
		},
		Validate: survey.Required,
	},
}

var allowlist = []string{"default-token", "istio-ca", "sh.helm.release"}

func main() {
	answers := struct {
		Namespace string
		Resources []string
		Age       string
	}{}

	commando.
		SetExecutableName("karetaker").
		SetVersion("1.0.0")

	commando.
		Register("interactive").
		SetDescription("Interactively build your clean-up logic").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			err := survey.Ask(questions, &answers)
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			fmt.Printf("Selected %s in Namespace %s for %s.\n", answers.Resources, answers.Namespace, answers.Age)
		})

	commando.
		Register("duplicate").
		SetDescription("Find similar or duplicate Kubernetes' deployments").
		AddArgument("target", "label to target similarities and duplicates", "kubernetes.io/instance").
		AddFlag("filter,f", "deployments label filter (i.e. app=auth)", commando.String, nil).
		AddFlag("namespace,n", "kubernetes namespace", commando.String, "default").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
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
		})

	commando.
		Register("age").
		SetDescription("Find resources older than a certain age").
		AddArgument("type", "type of resource", "deployment").
		AddFlag("age,a", "age boundary to filter on", commando.String, "48h").
		AddFlag("namespace,n", "kubernetes namespace", commando.String, "default").
		AddFlag("allow,A", "allow list (CSV) of name patterns to ignore (i.e. 'istio')", commando.String, "").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
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

		})

	commando.
		Register("unused").
		SetDescription("Find resources not in use by another object").
		AddArgument("type", "type of resource", "configmap").
		AddFlag("namespace,n", "kubernetes namespace", commando.String, "default").
		AddFlag("dry-run,d", "if true, only show the resources", commando.Bool, true).
		AddFlag("allow,A", "allow list (CSV) of name patterns to ignore (i.e. 'istio')", commando.String, "").
		SetAction(actions.UnusedAction)

	commando.Parse(nil)
}
