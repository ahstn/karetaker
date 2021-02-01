package domain

import (
	"github.com/pkg/errors"
	"strings"
	"time"
)

type Unused struct {
	// Resources are all the types to act on, i.e. ("deployment", "configmap")
	Resources []string

	// Namespace is the Kubernetes namespace to operate in
	Namespace string

	// Allow is a list of patterns to ignore when operating (i.e. don't delete objects containing these)
	Allow     []string

	// DryRun controls if the deletion occurs or not
	DryRun    bool
}

type Age struct {
	// Resources are all the types to act on, i.e. ("deployment", "configmap")
	Resources []string

	// Age is the target to filter on
	Age       time.Duration

	// Namespace is the Kubernetes namespace to operate in
	Namespace string

	// Allow is a list of patterns to ignore when operating (i.e. don't delete objects containing these)
	Allow     []string

	// DryRun controls if the deletion occurs or not
	DryRun    bool
}

func NewAgeConfig(r, a, n string, d bool) (Age, error) {
	age, err := time.ParseDuration(a)
	if err != nil {
		return Age{}, errors.Wrap(err, "unsupported duration")
	}

	return Age{
		Resources: strings.Split(r, ","),
		Age:       age,
		DryRun:    d,
		Namespace: n,
	}, nil
}
