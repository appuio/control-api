package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;delete;update

var (
	// TypeKey is the label key to identify organization namespaces
	TypeKey = "appuio.io/resource.type"
	// OrgType is the label value to identify organization namespaces
	OrgType = "organization"
	// DisplayNameKey is the annotation key that stores the display name
	DisplayNameKey = "organization.appuio.io/display-name"
	// BillingEntityRefKey is the annotation key that stores the billing entity reference
	BillingEntityRefKey = "organization.appuio.io/billing-entity-ref"
)

// NewOrganizationFromNS returns an Organization based on the given namespace
// If the namespace does not represent an organization it will return nil
func NewOrganizationFromNS(ns *corev1.Namespace) *Organization {
	if ns == nil || ns.Labels == nil || ns.Labels[TypeKey] != OrgType {
		return nil
	}
	var displayName, billingEntityRef string
	if ns.Annotations != nil {
		displayName = ns.Annotations[DisplayNameKey]
		billingEntityRef = ns.Annotations[BillingEntityRefKey]
	}
	org := &Organization{
		ObjectMeta: *ns.ObjectMeta.DeepCopy(),
		Spec: OrganizationSpec{
			DisplayName:      displayName,
			BillingEntityRef: billingEntityRef,
		},
	}
	if org.Annotations != nil {
		delete(org.Annotations, DisplayNameKey)
		delete(org.Annotations, BillingEntityRefKey)
		delete(org.Labels, TypeKey)
	}
	return org
}

// +kubebuilder:object:root=true

// Organization is a representation of an APPUiO Cloud Organization
type Organization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the cluster specific metadata.
	Spec OrganizationSpec `json:"spec,omitempty"`
}

// OrganizationSpec defines the desired state of the Organization
type OrganizationSpec struct {
	// DisplayName is a human-friendly name
	DisplayName string `json:"displayName,omitempty"`

	// BillingEntityRef is the reference to the billing entity
	BillingEntityRef string `json:"billingEntityRef,omitempty"`
}

// Organization needs to implement the builder resource interface
var _ resource.Object = &Organization{}

// GetObjectMeta returns the objects meta reference.
func (o *Organization) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
// The resource should be the all lowercase and pluralized kind
func (o *Organization) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    GroupVersion.Group,
		Version:  GroupVersion.Version,
		Resource: "organizations",
	}
}

// IsStorageVersion returns true if the object is also the internal version -- i.e. is the type defined for the API group or an alias to this object.
// If false, the resource is expected to implement MultiVersionObject interface.
func (o *Organization) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true if the object is namespaced
func (o *Organization) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (o *Organization) New() runtime.Object {
	return &Organization{}
}

// NewList return a new list instance of the resource
func (o *Organization) NewList() runtime.Object {
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

// ToNamespace translates an Organization to the underlying namespace representation
func (o *Organization) ToNamespace() *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: *o.ObjectMeta.DeepCopy(),
	}
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}
	ns.Labels[TypeKey] = OrgType
	ns.Annotations[DisplayNameKey] = o.Spec.DisplayName
	ns.Annotations[BillingEntityRefKey] = o.Spec.BillingEntityRef
	return ns
}

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}
