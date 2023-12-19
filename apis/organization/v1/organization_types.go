package v1

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

const (
	// SaleOrderCreated is set when the Sale Order has been created
	ConditionSaleOrderCreated = "SaleOrderCreated"

	// SaleOrderNameUpdated is set when the Sale Order's name has been added to the Status
	ConditionSaleOrderNameUpdated = "SaleOrderNameUpdated"

	ConditionReasonCreateFailed = "CreateFailed"

	ConditionReasonGetNameFailed = "GetNameFailed"
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
	// BillingEntityNameKey is the annotation key that stores the billing entity name
	BillingEntityNameKey = "status.organization.appuio.io/billing-entity-name"
	// SalesOrderIdKey is the annotation key that stores the sale order ID
	SalesOrderIdKey = "status.organization.appuio.io/sales-order-id"
	// SalesOrderNameKey is the annotation key that stores the sale order name
	SalesOrderNameKey = "status.organization.appuio.io/sales-order-name"
	// StatusConditionsKey is the annotation key that stores the serialized status conditions
	StatusConditionsKey = "status.organization.appuio.io/conditions"
)

// NewOrganizationFromNS returns an Organization based on the given namespace
// If the namespace does not represent an organization it will return nil
func NewOrganizationFromNS(ns *corev1.Namespace) *Organization {
	if ns == nil || ns.Labels == nil || ns.Labels[TypeKey] != OrgType {
		return nil
	}
	var displayName, billingEntityRef, billingEntityName, saleOrderId, saleOrderName, statusConditionsString string
	if ns.Annotations != nil {
		displayName = ns.Annotations[DisplayNameKey]
		billingEntityRef = ns.Annotations[BillingEntityRefKey]
		billingEntityName = ns.Annotations[BillingEntityNameKey]
		statusConditionsString = ns.Annotations[StatusConditionsKey]
		saleOrderId = ns.Annotations[SalesOrderIdKey]
		saleOrderName = ns.Annotations[SalesOrderNameKey]
	}
	var conditions []metav1.Condition
	err := json.Unmarshal([]byte(statusConditionsString), &conditions)
	if err != nil {
		conditions = nil
	}
	org := &Organization{
		ObjectMeta: *ns.ObjectMeta.DeepCopy(),
		Spec: OrganizationSpec{
			DisplayName:      displayName,
			BillingEntityRef: billingEntityRef,
		},
		Status: OrganizationStatus{
			BillingEntityName: billingEntityName,
			SalesOrderID:      saleOrderId,
			SalesOrderName:    saleOrderName,
			Conditions:        conditions,
		},
	}
	if org.Annotations != nil {
		delete(org.Annotations, DisplayNameKey)
		delete(org.Annotations, BillingEntityRefKey)
		delete(org.Annotations, BillingEntityNameKey)
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
	// Status holds the organization specific status
	Status OrganizationStatus `json:"status,omitempty"`
}

// OrganizationSpec defines the desired state of the Organization
type OrganizationSpec struct {
	// DisplayName is a human-friendly name
	DisplayName string `json:"displayName,omitempty"`

	// BillingEntityRef is the reference to the billing entity
	BillingEntityRef string `json:"billingEntityRef,omitempty"`
}

type OrganizationStatus struct {
	// BillingEntityName is the name of the billing entity
	BillingEntityName string `json:"billingEntityName,omitempty"`

	// SalesOrderID is the ID of the sale order
	SalesOrderID string `json:"salesOrderId,omitempty"`

	// SalesOrderName is the name of the sale order
	SalesOrderName string `json:"salesOrderName,omitempty"`

	// Conditions is a list of conditions for the invitation
	Conditions []metav1.Condition `json:"conditions,omitempty"`
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
	var statusString string
	if o.Status.Conditions != nil {
		statusBytes, err := json.Marshal(o.Status.Conditions)
		if err == nil {
			statusString = string(statusBytes)
		}
	}

	ns.Labels[TypeKey] = OrgType
	ns.Annotations[DisplayNameKey] = o.Spec.DisplayName
	ns.Annotations[BillingEntityRefKey] = o.Spec.BillingEntityRef
	ns.Annotations[BillingEntityNameKey] = o.Status.BillingEntityName
	if o.Status.SalesOrderID != "" {
		ns.Annotations[SalesOrderIdKey] = o.Status.SalesOrderID
	}
	if o.Status.SalesOrderName != "" {
		ns.Annotations[SalesOrderNameKey] = o.Status.SalesOrderName
	}
	if statusString != "" {
		ns.Annotations[StatusConditionsKey] = statusString
	}
	return ns
}

func init() {
	SchemeBuilder.Register(&Organization{}, &OrganizationList{})
}
