package kubernetes

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/xrash/smetrics"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Deployment is a stripped down version of the Kubernetes Resource.
// It only holds the name and age of the resource.
type Deployment struct {
	Name string
	Age  time.Duration
}

var (
	// DeploymentResource is a helper schema for interacting with Kubernetes Deployments
	DeploymentResource = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
)

// ListDeploymentsOlderThan returns a list of the deployments older than the duration 'd'
func ListDeploymentsOlderThan(c dynamic.Interface, n string, d time.Duration) ([]Deployment, error) {
	list, err := c.Resource(DeploymentResource).Namespace(n).List(context.TODO(), meta_v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "getting deployments")
	}

	now := time.Now()
	deployments := []Deployment{}
	for _, deployment := range list.Items {
		t, found, err := unstructured.NestedString(deployment.Object, "metadata", "creationTimestamp")
		if err != nil || !found {
			return nil, err
		}

		creation, err := time.Parse(time.RFC3339, t)

		age := now.Sub(creation)
		if age > d {
			name, found, err := unstructured.NestedString(deployment.Object, "metadata", "name")
			if err != nil || !found {
				return nil, err
			}
			deployments = append(deployments, Deployment{
				Name: name,
				Age:  age.Round(time.Minute),
			})
		}
	}

	return deployments, nil
}

// ListDuplicateDeployments finds potential duplicate deployments from similar labels
// i.e. [dev, release, john4, dev2, john5] should match [dev, dev2] and [john4, john5]
func ListDuplicateDeployments(clientset kubernetes.Interface,
	namespace string,
	appLabel string,
	instanceLabel string) (map[string][]string, error) {
	listopt := meta_v1.ListOptions{
		LabelSelector: appLabel,
	}

	list, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), listopt)
	if err != nil {
		return nil, errors.Wrap(err, "getting deployments")
	}

	// Copy all 'instance' labels into a string slice for iteration
	// Copy all 'instance' labels into a map to store with their matches
	var instances = []string{}
	similar := make(map[string][]string)
	for _, deployment := range list.Items {
		instances = append(instances, deployment.ObjectMeta.Labels[instanceLabel])
		similar[deployment.ObjectMeta.Labels[instanceLabel]] = []string{}
	}

	// Regex to remove anything but characters
	reg, err := regexp.Compile("[^a-zA-Z]+")
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("Found Instances: %s \n", instances)

	// Copy []instances into a map (key: name, value: similar names).
	// For each "version" iterate through the []instances to find matches.
	// If a match is found, add it to the coresponding map[target] value slice
	// and remove the 'match' (as a key) from the map.
	// If a match hasn't been found, remove the 'target' (as a key) from the map.
	for _, v := range instances {
		if _, ok := similar[v]; !ok {
			continue
		}

		matched := false
		for n, j := range instances {
			if v != j && smetrics.JaroWinkler(reg.ReplaceAllString(v, ""), j, 0.7, 4) > 0.9 {
				similar[v] = append(similar[v], j)
				delete(similar, j)
				matched = true
			}
			if n == len(instances)-1 && !matched {
				// fmt.Printf("End of list and no matches - removing %s\n", v)
				delete(similar, v)
			}
		}
	}

	// fmt.Printf("Result Map: %s\n", similar)

	// NB: If we make this concurrent by taking chucks of []instances
	// It'll still need a final last to ensure all the chucks are filtered together

	return similar, nil
}
