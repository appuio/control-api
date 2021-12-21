package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

// +kubebuilder:object:root=true

// Organization is a representation of an APPUiO Cloud Organization
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Data holds the cluster specific metadata.
	Spec OrganizationSpec `json:"spec,omitempty"`
}

// OrganizationSpec defines the desired state of the Organization
type OrganizationSpec struct {
	// DisplayName is a human-friendly name
	DisplayName string `json:"displayName,omitempty"`
}

// Organization needs to implement the builder resource interface
var _ resource.Object = &Organization{}

// GetObjectMeta returns the objects meta reference.
func (n *Organization) GetObjectMeta() *metav1.ObjectMeta {
	return &n.ObjectMeta
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
// The resource should be the all lowercase and pluralized kind
func (n *Organization) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    GroupVersion.Group,
		Version:  GroupVersion.Version,
		Resource: "organizations",
	}
}

// IsStorageVersion returns true if the object is also the internal version -- i.e. is the type defined for the API group or an alias to this object.
// If false, the resource is expected to implement MultiVersionObject interface.
func (n *Organization) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true if the object is namespaced
func (n *Organization) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (n *Organization) New() runtime.Object {
	return &Organization{}
}

// NewList return a new list instance of the resource
func (n *Organization) NewList() runtime.Object {
	return &OrganizationList{}
}

// +kubebuilder:object:root=true

// OrganizationList contains a list of Organizations
type OrganizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Organization `json:"items"`
}

// OrganizationList needs to implement the builder resource interface
var _ resource.ObjectList = &OrganizationList{}

// GetListMeta returns the list meta reference.
func (in *OrganizationList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}
