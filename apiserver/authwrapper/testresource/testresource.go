// +kubebuilder:object:generate=true
// +kubebuilder:skip
// +groupName=test
package testresource

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "test", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// +kubebuilder:object:root=true

// TestResource implements resource.Object
type TestResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Field1 string `json:"field1"`
}

// TestResource needs to implement the builder resource interface
var _ resource.Object = &TestResource{}

// GetObjectMeta returns the objects meta reference.
func (o *TestResource) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
// The resource should be the all lowercase and pluralized kind
func (o *TestResource) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    GroupVersion.Group,
		Version:  GroupVersion.Version,
		Resource: "testresources",
	}
}

// IsStorageVersion returns true if the object is also the internal version -- i.e. is the type defined for the API group or an alias to this object.
// If false, the resource is expected to implement MultiVersionObject interface.
func (o *TestResource) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true if the object is namespaced
func (o *TestResource) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (o *TestResource) New() runtime.Object {
	return &TestResource{}
}

// NewList return a new list instance of the resource
func (o *TestResource) NewList() runtime.Object {
	return &TestResourceList{}
}

// +kubebuilder:object:root=true

// TestResourceList contains a list of TestResources
type TestResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []TestResource `json:"items"`
}

// TestResourceList needs to implement the builder resource interface
var _ resource.ObjectList = &TestResourceList{}

// GetListMeta returns the list meta reference.
func (in *TestResourceList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

func init() {
	SchemeBuilder.Register(&TestResource{}, &TestResourceList{})
}
