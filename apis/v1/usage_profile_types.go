package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// UsageProfile is a representation of an APPUiO Cloud usage profile
type UsageProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UsageProfileSpec   `json:"spec,omitempty"`
	Status UsageProfileStatus `json:"status,omitempty"`
}

// UsageProfileSpec contains the desired state of the usage profile
type UsageProfileSpec struct {
	// NamespaceCount is the number of namespaces an organization with this usage profile can create per zone.
	NamespaceCount int `json:"namespaceCount,omitempty"`
	// Resources is the set of resources which are created in each namespace for which the usage profile is applied.
	// The key is used as the name of the resource and the value is the resource definition.
	Resources map[string]runtime.RawExtension `json:"resources,omitempty"`
}

// UsageProfileStatus contains the actual state of the usage profile
type UsageProfileStatus struct {
}

// +kubebuilder:object:root=true

// UsageProfileList contains a list of UsageProfiles.
type UsageProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []UsageProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&UsageProfile{}, &UsageProfileList{})
}
