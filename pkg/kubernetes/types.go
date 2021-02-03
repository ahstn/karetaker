package kubernetes

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	PodSchema = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	ConfigMapSchema = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	SecretSchema = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}

	DeploymentSchema = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	StatefulSetSchema = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}
)
