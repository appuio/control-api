package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status

// User is a representation of a APPUiO Cloud user
type User struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserSpec   `json:"spec,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}
type UserSpec struct {
	Preferences UserPreferences `json:"Preferences,omitempty"`
}
type UserPreferences struct {
	DefaultOrganizationRef string `json:"defaultOrganizationRef,omitempty"`
}
type UserStatus struct {
	DefaultOrganizationRef string `json:"defaultOrganization,omitempty"`
	DisplayName            string `json:"displayName,omitempty"`
	Username               string `json:"username,omitempty"`
	Email                  string `json:"email,omitempty"`
}

// +kubebuilder:object:root=true

// UserList contains a list of Users.
type UserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []User `json:"items"`
}

func init() {
	SchemeBuilder.Register(&User{}, &UserList{})
}
