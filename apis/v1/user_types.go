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

// UserSpec contains the desired state of the user
type UserSpec struct {
	Preferences UserPreferences `json:"preferences,omitempty"`
}

// UserPreferences contains the Preferences of the user
type UserPreferences struct {
	DefaultOrganizationRef string `json:"defaultOrganizationRef,omitempty"`
}

// UserStatus contains the acutal state of the user
type UserStatus struct {
	DefaultOrganizationRef string `json:"defaultOrganization,omitempty"`
	ID                     string `json:"id,omitempty"`
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
