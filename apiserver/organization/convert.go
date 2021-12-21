package organization

import (
	"fmt"
	"strings"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	typeKey        = "appuio.io/resource.type"
	nameKey        = "appuio.io/metadata.name"
	namespaceKey   = "organization.appuio.io/namespace"
	displayNameKey = "organization.appuio.io/display-name"
)

func namespaceToOrg(ns *corev1.Namespace) *orgv1.Organization {
	displayName := ""
	if ns.Annotations != nil {
		displayName = ns.Annotations[displayNameKey]
	}
	return &orgv1.Organization{
		ObjectMeta: metav1.ObjectMeta{
			Name:              namespaceNameToOrgName(ns.Name),
			CreationTimestamp: ns.CreationTimestamp,
			Annotations: map[string]string{
				namespaceKey: ns.Name,
			},
		},
		Spec: orgv1.OrganizationSpec{
			DisplayName: displayName,
		},
	}
}

func namespaceNameToOrgName(ns string) string {
	return strings.TrimPrefix(ns, "org-")
}

func orgToNamespace(org *orgv1.Organization) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: orgNameToNamespaceName(org.Name),
			Labels: map[string]string{
				typeKey: "organization",
				nameKey: org.Name,
			},
			Annotations: map[string]string{
				displayNameKey: org.Spec.DisplayName,
			},
		},
	}
}
func orgNameToNamespaceName(org string) string {
	return fmt.Sprintf("org-%s", org)
}
