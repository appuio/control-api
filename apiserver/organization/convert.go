package organization

import (
	"fmt"
	"strings"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
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
	org := &orgv1.Organization{
		ObjectMeta: *ns.ObjectMeta.DeepCopy(),
		Spec: orgv1.OrganizationSpec{
			DisplayName: displayName,
		},
	}
	org.Name = namespaceNameToOrgName(ns.Name)
	if org.Annotations == nil {
		org.Annotations = map[string]string{}
	}
	org.Annotations[namespaceKey] = ns.Name
	delete(org.Annotations, displayNameKey)
	delete(org.Labels, typeKey)
	delete(org.Labels, nameKey)
	return org
}

func namespaceNameToOrgName(ns string) string {
	return strings.TrimPrefix(ns, "org-")
}

func orgToNamespace(org *orgv1.Organization) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: *org.ObjectMeta.DeepCopy(),
	}
	ns.Name = orgNameToNamespaceName(org.Name)
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}

	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}
	ns.Labels[typeKey] = "organization"
	ns.Labels[nameKey] = org.Name
	ns.Annotations[displayNameKey] = org.Spec.DisplayName
	return ns
}
func orgNameToNamespaceName(org string) string {
	return fmt.Sprintf("org-%s", org)
}
