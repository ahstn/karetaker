package main

import (
	"fmt"
	"os"
	"text/tabwriter"

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
		Register("batch").
		SetDescription("Execute a batch run using pre-existing clean-up logic").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			clientset, err := kubernetes.Config("")
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			deployments, err := kubernetes.ListDeployments(clientset)
			for _, deployment := range deployments {
				fmt.Printf("deploy/%s (Age: %s)\n", deployment.Name, deployment.Age)
			}
		})

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
	commando.Parse(nil)
}
