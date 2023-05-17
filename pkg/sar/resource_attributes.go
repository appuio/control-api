package sar

// ResourceAttributes includes the authorization attributes available for resource requests to the Authorizer interface.
// From https://github.com/kubernetes/api/blob/2f9553831ec24dc60e3e1c3a374fb63ca091688f/authorization/v1/types.go#L92-L118.
// Importing the whole package confuses go mod.
type ResourceAttributes struct {
	// Namespace is the namespace of the action being requested.  Currently, there is no distinction between no namespace and all namespaces
	// "" (empty) is defaulted for LocalSubjectAccessReviews
	// "" (empty) is empty for cluster-scoped resources
	// "" (empty) means "all" for namespace scoped resources from a SubjectAccessReview or SelfSubjectAccessReview
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	// Verb is a kubernetes resource API verb, like: get, list, watch, create, update, delete, proxy.  "*" means all.
	// +optional
	Verb string `json:"verb,omitempty" protobuf:"bytes,2,opt,name=verb"`
	// Group is the API Group of the Resource.  "*" means all.
	// +optional
	Group string `json:"group,omitempty" protobuf:"bytes,3,opt,name=group"`
	// Version is the API Version of the Resource.  "*" means all.
	// +optional
	Version string `json:"version,omitempty" protobuf:"bytes,4,opt,name=version"`
	// Resource is one of the existing resource types.  "*" means all.
	// +optional
	Resource string `json:"resource,omitempty" protobuf:"bytes,5,opt,name=resource"`
	// Subresource is one of the existing resource types.  "" means none.
	// +optional
	Subresource string `json:"subresource,omitempty" protobuf:"bytes,6,opt,name=subresource"`
	// Name is the name of the resource being requested for a "get" or deleted for a "delete". "" (empty) means all.
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,7,opt,name=name"`
}
