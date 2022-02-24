package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Team is the collection of members of a team.
type Team struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TeamSpec   `json:"spec,omitempty"`
	Status TeamStatus `json:"status,omitempty"`
}

// TeamSpec contains the desired members of a team.
type TeamSpec struct {
	DisplayName string    `json:"displayName,omitempty"`
	UserRefs    []UserRef `json:"userRefs"`
}

// TeamStatus contains the actual members of a team.
type TeamStatus struct {
	ResolvedUserRefs []UserRef `json:"resolvedUserRefs,omitempty"`
}

// +kubebuilder:object:root=true

// TeamList contains a list of Team resources
type TeamList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Team `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Team{}, &TeamList{})
}
