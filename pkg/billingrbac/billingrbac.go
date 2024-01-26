// billingrbac is the common definition of the RBAC rules for billing entities.
package billingrbac

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterRolesParams struct {
	AdminUsers  []string
	ViewerUsers []string

	// AllowSubjectsToViewRole controls whether the admin and viewer roles are allowed to view the role itself.
	// Workaround for the portal crashing when trying to load the role.
	AllowSubjectsToViewRole bool
}

// ClusterRoles returns the ClusterRoles and ClusterRoleBindings for the given BillingEntity.
func ClusterRoles(beName string, p ClusterRolesParams) (ar *rbacv1.ClusterRole, arBinding *rbacv1.ClusterRoleBinding, vr *rbacv1.ClusterRole, vrBinding *rbacv1.ClusterRoleBinding) {
	viewRoleName := fmt.Sprintf("billingentities-%s-viewer", beName)
	viewRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: viewRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"rbac.appuio.io"},
				Resources:     []string{"billingentities"},
				Verbs:         []string{"get"},
				ResourceNames: []string{beName},
			},
		},
	}
	if p.AllowSubjectsToViewRole {
		viewRole.Rules = append(viewRole.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{"rbac.authorization.k8s.io"},
			Resources:     []string{"clusterroles"},
			Verbs:         []string{"get"},
			ResourceNames: []string{viewRoleName},
		})
	}
	viewRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: viewRoleName,
		},
		Subjects: userSubjects(p.ViewerUsers),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     viewRoleName,
		},
	}
	adminRoleName := fmt.Sprintf("billingentities-%s-admin", beName)
	adminRole := &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: adminRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"rbac.appuio.io", "billing.appuio.io"},
				Resources:     []string{"billingentities"},
				Verbs:         []string{"get", "patch", "update", "edit"},
				ResourceNames: []string{beName},
			},
			{
				APIGroups:     []string{"rbac.authorization.k8s.io"},
				Resources:     []string{"clusterrolebindings"},
				Verbs:         []string{"get", "edit", "update", "patch"},
				ResourceNames: []string{viewRoleName, adminRoleName},
			},
		},
	}
	if p.AllowSubjectsToViewRole {
		adminRole.Rules = append(adminRole.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{"rbac.authorization.k8s.io"},
			Resources:     []string{"clusterroles"},
			Verbs:         []string{"get"},
			ResourceNames: []string{viewRoleName, adminRoleName},
		})
	}
	adminRoleBinding := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: adminRoleName,
		},
		Subjects: userSubjects(p.AdminUsers),
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     adminRoleName,
		},
	}

	return adminRole, adminRoleBinding, viewRole, viewRoleBinding
}

func userSubjects(users []string) []rbacv1.Subject {
	subjects := make([]rbacv1.Subject, 0, len(users))
	for _, u := range users {
		subjects = append(subjects, rbacv1.Subject{
			Kind:     "User",
			APIGroup: "rbac.authorization.k8s.io",
			Name:     u,
		})
	}
	return subjects
}
