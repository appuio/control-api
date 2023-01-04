package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

// +kubebuilder:object:root=true

// BillingEntity is a representation of an APPUiO Cloud BillingEntity
type BillingEntity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the cluster specific metadata.
	Spec BillingEntitySpec `json:"spec,omitempty"`
}

// BillingEntitySpec defines the desired state of the BillingEntity
type BillingEntitySpec struct {
	// Name is the human-readable name of the BillingEntity
	Name string `json:"name"`
	// Phone is the phone number of the BillingEntity
	Phone string `json:"phone"`
	// Emails is a list of email addresses of the BillingEntity
	Emails []string `json:"emails"`
	// Address is the postal address of the BillingEntity
	Address BillingEntityAddress `json:"address"`

	// AccountingContact is the contact person for accounting
	AccountingContact BillingEntityContact `json:"accountingContact"`

	// LanguagePreference is the preferred language of the BillingEntity
	LanguagePreference string `json:"languagePreference"`
}

type BillingEntityAddress struct {
	// Line1 is the first line of the address
	Line1 string `json:"line1"`
	// Line2 is the second line of the address
	Line2 string `json:"line2,omitempty"`
	// City is the city of the address
	City string `json:"city"`
	// PostalCode is the postal code of the address
	PostalCode string `json:"postalCode"`
	// Country is the country of the address
	Country string `json:"country"`
}

type BillingEntityContact struct {
	// Name is the name of the contact person
	Name string `json:"name"`
	// Emails is a list of email addresses of the contact person
	Emails []string `json:"emails"`
}

// BillingEntity needs to implement the builder resource interface
var _ resource.Object = &BillingEntity{}

// GetObjectMeta returns the objects meta reference.
func (o *BillingEntity) GetObjectMeta() *metav1.ObjectMeta {
	return &o.ObjectMeta
}

// GetGroupVersionResource returns the GroupVersionResource for this resource.
// The resource should be the all lowercase and pluralized kind
func (o *BillingEntity) GetGroupVersionResource() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    GroupVersion.Group,
		Version:  GroupVersion.Version,
		Resource: "billingentities",
	}
}

// IsStorageVersion returns true if the object is also the internal version -- i.e. is the type defined for the API group or an alias to this object.
// If false, the resource is expected to implement MultiVersionObject interface.
func (o *BillingEntity) IsStorageVersion() bool {
	return true
}

// NamespaceScoped returns true if the object is namespaced
func (o *BillingEntity) NamespaceScoped() bool {
	return false
}

// New returns a new instance of the resource
func (o *BillingEntity) New() runtime.Object {
	return &BillingEntity{}
}

// NewList return a new list instance of the resource
func (o *BillingEntity) NewList() runtime.Object {
	return &BillingEntityList{}
}

// +kubebuilder:object:root=true

// BillingEntityList contains a list of BillingEntitys
type BillingEntityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []BillingEntity `json:"items"`
}

// BillingEntityList needs to implement the builder resource interface
var _ resource.ObjectList = &BillingEntityList{}

// GetListMeta returns the list meta reference.
func (in *BillingEntityList) GetListMeta() *metav1.ListMeta {
	return &in.ListMeta
}

func init() {
	SchemeBuilder.Register(&BillingEntity{}, &BillingEntityList{})
}
