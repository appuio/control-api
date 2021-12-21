package v1

import (
	"fmt"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// Zone is the Schema for the Zone API
type Zone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Data holds the cluster specific metadata.
	Data ZoneData `json:"data,omitempty"`
}

// ZoneData holds all the Zone specific properties
type ZoneData struct {
	// DisplayName is a human-friendly name for the Zone.
	DisplayName string `json:"displayName,omitempty"`
	// Features holds a key-value dict with keys being a feature name and values being a property of that feature.
	// Some features may hold a version string as property.
	Features Features `json:"features,omitempty"`
	// URLs holds a key-value dict with keys being a name of the URL and the values publicly accessible links.
	URLs URLMap `json:"urls,omitempty"`
	// CNAME is the DNS record where custom application DNS hostnames shall be pointing to when exposing an application.
	CNAME string `json:"cname,omitempty"`
	// DefaultAppDomain is the base DNS record where OpenShift Routes without specific hostnames are exposed.
	DefaultAppDomain string `json:"defaultAppDomain,omitempty"`
	// GatewayIPs holds the outgoing IP addresses of the cluster.
	GatewayIPs []string `json:"gatewayIPs,omitempty"`
	// CloudProvider identifies the infrastructure provider which the Zone is running on.
	CloudProvider CloudProvider `json:"cloudProvider,omitempty"`
}

// Features is a key-value dict with keys being a feature name and values being a property of that feature.
type Features map[string]string

// URLMap is a key-value dict with keys being a name of the URL and the values publicly accessible links.
type URLMap map[string]string

// CloudProvider identifies an infrastructure provider.
type CloudProvider struct {
	// Name identifies the cloud provider.
	Name string `json:"name,omitempty"`
	// Zones is cloud-provider-specific zone aliases within a Region.
	// If multiple entries are present, the cluster may be spanning multiple zones.
	Zones []string `json:"zones,omitempty"`
	// Region is the geographic location of the Zone.
	Region string `json:"region,omitempty"`
}

// +kubebuilder:object:root=true

// ZoneList contains a list of Zone.
type ZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Zone `json:"items"`
}

// GetURL invokes url.Parse for the raw string from given key, if found.
func (in URLMap) GetURL(key string) (*url.URL, error) {
	if in == nil {
		return nil, fmt.Errorf("map is nil")
	}
	if raw, found := in[key]; found {
		return url.Parse(raw)
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

func init() {
	SchemeBuilder.Register(&Zone{}, &ZoneList{})
}
