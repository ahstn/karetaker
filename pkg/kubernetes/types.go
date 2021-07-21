package kubernetes

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	PodSchema = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	ConfigMapSchema = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	SecretSchema = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	ServiceSchema = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}

	DeploymentSchema = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	StatefulSetSchema = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}

	JobSchema = schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}
)
