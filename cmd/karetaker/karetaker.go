package main

import (
	"github.com/ahstn/karetaker/cmd/karetaker/actions"
	"github.com/thatisuday/commando"
)

var allowlist = []string{"default-token", "istio-ca", "sh.helm.release"}

func main() {
	commando.
		SetExecutableName("karetaker").
		SetVersion("1.0.0")

	commando.
		Register("duplicate").
		SetDescription("Find similar or duplicate Kubernetes' deployments").
		AddArgument("target", "label to target similarities and duplicates", "kubernetes.io/instance").
		AddFlag("filter,f", "deployments label filter (i.e. app=auth)", commando.String, nil).
		AddFlag("namespace,n", "kubernetes namespace", commando.String, "default").
		SetAction(actions.Duplicate)

	commando.
		Register("age").
		SetDescription("Find resources older than a certain age").
		AddArgument("type", "type of resource", "deployment").
		AddFlag("age,a", "age boundary to filter on", commando.String, "48h").
		AddFlag("namespace,n", "kubernetes namespace", commando.String, "default").
		AddFlag("dry-run,d", "if true, only show the resources", commando.Bool, true).
		AddFlag("allow,A", "allow list (CSV) of name patterns to ignore (i.e. 'istio')", commando.String, "").
		SetAction(actions.Age)

	commando.
		Register("unused").
		SetDescription("Find resources not in use by another object").
		AddArgument("type", "type of resource", "configmap").
		AddFlag("age,a", "age boundary to filter on (only for certain resources)", commando.String, "24h").
		AddFlag("namespace,n", "kubernetes namespace", commando.String, "default").
		AddFlag("dry-run,d", "if true, only show the resources", commando.Bool, true).
		AddFlag("allow,A", "allow list (CSV) of name patterns to ignore (i.e. 'istio')", commando.String, "").
		SetAction(actions.Unused)

	commando.Parse(nil)
}
