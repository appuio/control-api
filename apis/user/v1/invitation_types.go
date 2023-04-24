package v1

import (
	"net/url"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

	"github.com/appuio/control-api/apiserver/secretstorage/status"
)

const (
	// ConditionRedeemed is set when the invitation has been redeemed
	ConditionRedeemed = "Redeemed"
	// ConditionEmailSent is set when the invitation email has been sent
	ConditionEmailSent        = "EmailSent"
	ConditionReasonSendFailed = "SendFailed"
)

// +kubebuilder:object:root=true

// Invitation is a representation of an APPUiO Cloud Invitation
type Invitation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired invitation state
	Spec InvitationSpec `json:"spec,omitempty"`
	// Status holds the invitation specific status
	Status InvitationStatus `json:"status,omitempty"`
}

// InvitationSpec defines the desired state of the Invitation
type InvitationSpec struct {
	// Note is a free-form text field to add a note to the invitation
	Note string `json:"note,omitempty"`
	// Email is the email address of the invited user, used to send the invitation
	Email string `json:"email,omitempty"`
	// TargetRefs is a list of references to the target resources
	TargetRefs []TargetRef `json:"targetRefs,omitempty"`
}

// TargetRef is a reference to a target resource
type TargetRef struct {
	// APIGroup is the API group of the target resource
	APIGroup string `json:"apiGroup,omitempty"`
	// Kind is the kind of the target resource
	Kind string `json:"kind,omitempty"`
	// Name is the name of the target resource
	Name string `json:"name,omitempty"`
	// Namespace is the namespace of the target resource
	Namespace string `json:"namespace,omitempty"`
}

// InvitationStatus defines the observed state of the Invitation
type InvitationStatus struct {
	// Token is the invitation token
	Token string `json:"token"`
	// ValidUntil is the time when the invitation expires
	ValidUntil metav1.Time `json:"validUntil"`
	// Conditions is a list of conditions for the invitation
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// TargetStatuses is a list of statuses for the target resources
	TargetStatuses []TargetStatus `json:"targetStatuses,omitempty"`
	// RedeemedBy is the user who redeemed the invitation
	RedeemedBy string `json:"redeemedBy,omitempty"`
}

// TargetStatus is the status of a target resource
type TargetStatus struct {
	Condition metav1.Condition `json:"condition"`
	TargetRef TargetRef        `json:"targetRef"`
}

// Invitation needs to implement the builder resource interface
var _ status.ObjectWithStatusSubResource = &Invitation{}

// GetObjectMeta returns the objects meta reference.
func (o *Invitation) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
// The resource should be the all lowercase and pluralized kind
func (o *Invitation) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    GroupVersion.Group,
		Version:  GroupVersion.Version,
		Resource: "invitations",
	}
}

// IsStorageVersion returns true if the object is also the internal version -- i.e. is the type defined for the API group or an alias to this object.
// If false, the resource is expected to implement MultiVersionObject interface.
func (o *Invitation) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true if the object is namespaced
func (o *Invitation) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (o *Invitation) New() runtime.Object {
	return &Invitation{}
}

// NewList return a new list instance of the resource
func (o *Invitation) NewList() runtime.Object {
	return &InvitationList{}
}

// IsRedeemed returns true if the invitation has been redeemed
func (o *Invitation) IsRedeemed() bool {
	return apimeta.IsStatusConditionTrue(o.Status.Conditions, ConditionRedeemed)
}

// SecretStorageGetStatus returns the status of the resource
func (o *Invitation) SecretStorageGetStatus() status.StatusSubResource {
	return &o.Status
}

// CopyTo copies the status to the given parent resource
func (s *InvitationStatus) SecretStorageCopyTo(parent status.ObjectWithStatusSubResource) {
	parent.(*Invitation).Status = *s.DeepCopy()
}

func (s InvitationStatus) SubResourceName() string {
	return "status"
}

// +kubebuilder:object:root=true

// InvitationList contains a list of Invitations
type InvitationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Invitation `json:"items"`
}

// InvitationList needs to implement the builder resource interface
var _ resource.ObjectList = &InvitationList{}

// GetListMeta returns the list meta reference.
func (in *InvitationList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type RedeemOptions struct {
	metav1.TypeMeta `json:",inline"`

	Token string `json:"token"`
}

var _ resource.QueryParameterObject = &RedeemOptions{}

// apiserver-runtime tracks its own version of the query parameter scheme.
// Just registering the conversion functions to the scheme is not enough.
// It also needs to implement the resource.QueryParameterObject interface.
func (o *RedeemOptions) ConvertFromUrlValues(values *url.Values) error {
	return convert_url_Values_To__RedeemOptions(values, o)
}

var _ resource.Object = &InvitationRedeemRequest{}

// +kubebuilder:object:root=true
// InvitationRedeemRequest is a request to redeem an invitation
type InvitationRedeemRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Token is the token to redeem the invitation
	Token string `json:"token"`
}

// GetObjectMeta returns the objects meta reference.
func (o *InvitationRedeemRequest) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
// The resource should be the all lowercase and pluralized kind
func (o *InvitationRedeemRequest) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    GroupVersion.Group,
		Version:  GroupVersion.Version,
		Resource: "invitationredeemrequests",
	}
}

// IsStorageVersion returns true if the object is also the internal version -- i.e. is the type defined for the API group or an alias to this object.
// If false, the resource is expected to implement MultiVersionObject interface.
func (o *InvitationRedeemRequest) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true if the object is namespaced
func (o *InvitationRedeemRequest) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (o *InvitationRedeemRequest) New() runtime.Object {
	return &InvitationRedeemRequest{}
}

// NewList return a new list instance of the resource
func (o *InvitationRedeemRequest) NewList() runtime.Object {
	return &InvitationRedeemRequestList{}
}

var _ resource.ObjectList = &InvitationRedeemRequestList{}

// +kubebuilder:object:root=true
// InvitationList contains a list of Invitations
type InvitationRedeemRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []InvitationRedeemRequest `json:"items"`
}

// GetListMeta returns the list meta reference.
func (in *InvitationRedeemRequestList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

func init() {
	SchemeBuilder.Register(&Invitation{}, &InvitationList{}, &InvitationRedeemRequest{})
}
