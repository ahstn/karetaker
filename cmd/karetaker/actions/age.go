package actions

import (
	"fmt"
	"github.com/ahstn/karetaker/pkg/actions"
	"github.com/ahstn/karetaker/pkg/domain"
	"github.com/ahstn/karetaker/pkg/kubernetes"
	"github.com/thatisuday/commando"
	"os"
	"strings"
	"text/tabwriter"
)

var allowlist = []string{"default-token", "istio-ca", "sh.helm.release"}

func Age(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
	n, _ := flags["namespace"].GetString()
	d, _ := flags["dry-run"].GetBool()
	a, _ := flags["age"].GetString()
	al, _ := flags["allow"].GetString()
	t := args["type"].Value
	allowlist = append(allowlist, strings.Split(al, ",")[:]...)

	config, err := domain.NewAgeConfig(t, a, n, allowlist, d)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Using Allow List of: %s\n", allowlist)
	fmt.Println("Connecting to Kubernetes Cluster")
	client, err := kubernetes.DynamicConfig("")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)
	defer w.Flush()

	err = actions.Age(client, config, w)
	if err != nil {
		panic(err)
	}
}
