package organization

import (
	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
)

var (
	typeKey        = "appuio.io/resource.type"
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
	if org.Annotations != nil {
		delete(org.Annotations, displayNameKey)
		delete(org.Labels, typeKey)
	}
	return org
}

func orgToNamespace(org *orgv1.Organization) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: *org.ObjectMeta.DeepCopy(),
	}
	if ns.Labels == nil {
		ns.Labels = map[string]string{}
	}
	if ns.Annotations == nil {
		ns.Annotations = map[string]string{}
	}
	ns.Labels[typeKey] = "organization"
	ns.Annotations[displayNameKey] = org.Spec.DisplayName
	return ns
}
