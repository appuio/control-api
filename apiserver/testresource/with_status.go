// +kubebuilder:object:generate=true
// +kubebuilder:skip
// +groupName=test
package testresource

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

	"github.com/appuio/control-api/apiserver/secretstorage/status"
)

// +kubebuilder:object:root=true
// TestResourceWithStatus implements resource.Object
type TestResourceWithStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Field1 string `json:"field1"`

	Status TestResourceWithStatusStatus `json:"status"`
}

type TestResourceWithStatusStatus struct {
	Num int `json:"num"`
}

// TestResourceWithStatus needs to implement the builder resource interface
var _ status.ObjectWithStatusSubResource = &TestResourceWithStatus{}

// GetObjectMeta returns the objects meta reference.
func (o *TestResourceWithStatus) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
// The resource should be the all lowercase and pluralized kind
func (o *TestResourceWithStatus) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    GroupVersion.Group,
		Version:  GroupVersion.Version,
		Resource: "testresourceswithstatus",
	}
}

// IsStorageVersion returns true if the object is also the internal version -- i.e. is the type defined for the API group or an alias to this object.
// If false, the resource is expected to implement MultiVersionObject interface.
func (o *TestResourceWithStatus) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true if the object is namespaced
func (o *TestResourceWithStatus) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (o *TestResourceWithStatus) New() runtime.Object {
	return &TestResourceWithStatus{}
}

// NewList return a new list instance of the resource
func (o *TestResourceWithStatus) NewList() runtime.Object {
	return &TestResourceWithStatusList{}
}

// GetStatus returns the status of the resource
func (o *TestResourceWithStatus) SecretStorageGetStatus() status.StatusSubResource {
	return &o.Status
}

// CopyTo copies the status to the given parent resource
func (s *TestResourceWithStatusStatus) SecretStorageCopyTo(parent status.ObjectWithStatusSubResource) {
	parent.(*TestResourceWithStatus).Status = *s.DeepCopy()
}

func (s TestResourceWithStatusStatus) SubResourceName() string {
	return "status"
}

// +kubebuilder:object:root=true
// TestResourceWithStatusList contains a list of TestResourceWithStatuss
type TestResourceWithStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []TestResourceWithStatus `json:"items"`
}

// TestResourceWithStatusList needs to implement the builder resource interface
var _ resource.ObjectList = &TestResourceWithStatusList{}

// GetListMeta returns the list meta reference.
func (in *TestResourceWithStatusList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

func init() {
	SchemeBuilder.Register(&TestResourceWithStatus{}, &TestResourceWithStatusList{})
}
