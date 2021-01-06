package actions

import (
	"fmt"
	"github.com/ahstn/karetaker/pkg/actions"
	"github.com/ahstn/karetaker/pkg/domain"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/ahstn/karetaker/pkg/kubernetes"
	"github.com/thatisuday/commando"
)

func Unused(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	n, _ := flags["namespace"].GetString()
	d, _ := flags["dry-run"].GetBool()
	a, _ := flags["allow"].GetString()
	t := args["type"].Value
	allowlist = append(allowlist, strings.Split(a, ",")[:]...)

	config := domain.Unused{
		Resources: strings.Split(t, ","),
		Namespace: n,
		Allow:     allowlist,
		DryRun:    d,
	}

	fmt.Printf("Using Allow List of: %s\n\n", allowlist)

	fmt.Println("Connecting to Kubernetes Cluster")
	client, err := kubernetes.DynamicConfig("")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	err = actions.Unused(client, config, w)
	if err != nil {
		panic(err)
	}
}