package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OrganizationMembers is the collection of members of an organization
type OrganizationMembers struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrganizationMembersSpec   `json:"spec,omitempty"`
	Status OrganizationMembersStatus `json:"status,omitempty"`
}

type OrganizationMembersSpec struct {
	UserRefs []UserRef `json:"userRefs,omitempty"`
}
type OrganizationMembersStatus struct {
	UserRefs []UserRef `json:"resolvedUserRefs,omitempty"`
}

type UserRef struct {
	ID       string `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
}

// +kubebuilder:object:root=true

// OrganizationMembersList contains a list of OrganizationMembers resources
type OrganizationMembersList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []OrganizationMembers `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OrganizationMembers{}, &OrganizationMembersList{})
}
