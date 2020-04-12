package main

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
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

	commando.Parse(nil)
}
