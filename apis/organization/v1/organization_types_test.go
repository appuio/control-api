package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewOrganizationFromNS(t *testing.T) {
	tests := map[string]struct {
		namespace    *corev1.Namespace
		organization *Organization
	}{
		"GivenNil_ThenNil": {},
		"GivenNonOrgNs_ThenNil": {
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			},
		},
		"GivenNonOrgNsWithLabel_ThenNil": {
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
					Labels: map[string]string{
						"foo": "bar",
					},
				},
			},
		},
		"GivenOrgNs_ThenOrg": {
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fooBar",
					Labels: map[string]string{
						TypeKey: OrgType,
					},
					Annotations: map[string]string{
						DisplayNameKey: "Foo Bar Inc.",
					},
				},
			},
			organization: &Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "fooBar",
					Labels:      map[string]string{},
					Annotations: map[string]string{},
				},
				Spec: OrganizationSpec{
					DisplayName: "Foo Bar Inc.",
				},
			},
		},
		"GivenOrgNsWithLabelAndAnnot_ThenOrgWithLabelAndAnnot": {
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fooBar",
					Labels: map[string]string{
						TypeKey: OrgType,
						"foo":   "bar",
					},
					Annotations: map[string]string{
						DisplayNameKey: "Foo Bar Inc.",
						"bar":          "buzz",
					},
				},
			},
			organization: &Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fooBar",
					Labels: map[string]string{
						"foo": "bar",
					},
					Annotations: map[string]string{
						"bar": "buzz",
					},
				},
				Spec: OrganizationSpec{
					DisplayName: "Foo Bar Inc.",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.organization, NewOrganizationFromNS(tt.namespace))
		})
	}
}

func TestOrganization_ToNamespace(t *testing.T) {
	tests := map[string]struct {
		organization *Organization
		namespace    *corev1.Namespace
	}{
		"GivenOrg_ThenOrgNs": {
			organization: &Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fooBar",
				},
				Spec: OrganizationSpec{
					DisplayName: "Foo Bar Inc.",
				},
			},
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fooBar",
					Labels: map[string]string{
						TypeKey: OrgType,
					},
					Annotations: map[string]string{
						DisplayNameKey: "Foo Bar Inc.",
					},
				},
			},
		},
		"GivenOrgWithLabelAndAnnot_ThenOrgNsWithLabelAndAnnot": {
			organization: &Organization{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fooBar",
					Labels: map[string]string{
						"foo": "bar",
					},
					Annotations: map[string]string{
						"bar": "buzz",
					},
				},
				Spec: OrganizationSpec{
					DisplayName: "Foo Bar Inc.",
				},
			},
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "fooBar",
					Labels: map[string]string{
						TypeKey: OrgType,
						"foo":   "bar",
					},
					Annotations: map[string]string{
						DisplayNameKey: "Foo Bar Inc.",
						"bar":          "buzz",
					},
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.namespace, tt.organization.ToNamespace())
		})
	}
}
