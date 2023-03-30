package controllers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
	. "github.com/appuio/control-api/controllers"
)

func Test_InvitationRedeemReconciler_Reconcile_Success(t *testing.T) {
	const redeemedBy = "example-user-01"
	ctx := context.Background()

	team := &controlv1.Team{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "appuio.io/v1",
			Kind:       "Team",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-team-01",
			Namespace: "example-organization-01",
		},
	}
	orgMem := &controlv1.OrganizationMembers{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "appuio.io/v1",
			Kind:       "OrganizationMembers",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "members",
			Namespace: "example-organization-01",
		},
		Spec: controlv1.OrganizationMembersSpec{
			UserRefs: []controlv1.UserRef{
				{
					Name: redeemedBy,
				},
			},
		},
	}
	rb := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "organization-admin",
			Namespace: "example-organization-01",
		},
	}
	crb := &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "organization-admin",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     rbacv1.UserKind,
				APIGroup: rbacv1.GroupName,
				Name:     redeemedBy,
			},
		},
	}

	subject := &userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Status: userv1.InvitationStatus{
			RedeemedBy: redeemedBy,
			TargetStatuses: []userv1.TargetStatus{
				{TargetRef: targetRefFromObject(team)},
				{TargetRef: targetRefFromObject(orgMem)},
				{TargetRef: targetRefFromObject(rb)},
				{TargetRef: targetRefFromObject(crb)},
			},
		},
	}
	apimeta.SetStatusCondition(&subject.Status.Conditions, metav1.Condition{
		Type:   userv1.ConditionRedeemed,
		Status: metav1.ConditionTrue,
	})

	c := prepareTest(t, team, orgMem, rb, crb, subject)

	r := invitationRedeemReconciler(c)
	_, err := r.Reconcile(ctx, requestFor(subject))
	require.NoError(t, err)

	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(subject), subject))
	for _, targetStatus := range subject.Status.TargetStatuses {
		assert.Equal(t, metav1.ConditionTrue, targetStatus.Condition.Status)
	}

	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(team), team))
	assert.Equal(t, []controlv1.UserRef{{Name: redeemedBy}}, team.Spec.UserRefs)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(orgMem), orgMem))
	assert.Equal(t, []controlv1.UserRef{{Name: redeemedBy}}, orgMem.Spec.UserRefs)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(rb), rb))
	assert.Equal(t, []rbacv1.Subject{{Kind: rbacv1.UserKind, APIGroup: rbacv1.GroupName, Name: redeemedBy}}, rb.Subjects)
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(crb), crb))
	assert.Equal(t, []rbacv1.Subject{{Kind: rbacv1.UserKind, APIGroup: rbacv1.GroupName, Name: redeemedBy}}, crb.Subjects)

	role := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "invitations-subject-redeemer",
		},
	}
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(role), role), "Redeeming person should have read access to the invitation")
	rolebinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "invitations-subject-redeemer",
		},
	}
	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(rolebinding), rolebinding), "Redeeming person should have read access to the invitation")
	require.Len(t, rolebinding.Subjects, 1)
	require.Equal(t, redeemedBy, rolebinding.Subjects[0].Name)

	_, err = r.Reconcile(ctx, requestFor(subject))
	require.NoError(t, err)
}

func Test_InvitationRedeemReconciler_Reconcile_NotRedeemed_Success(t *testing.T) {
	ctx := context.Background()

	subject := &userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}

	c := prepareTest(t, subject)
	_, err := invitationRedeemReconciler(c).Reconcile(ctx, requestFor(subject))
	require.NoError(t, err)
}

func Test_InvitationRedeemReconciler_Reconcile_Fail_NoUser(t *testing.T) {
	ctx := context.Background()

	subject := &userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
	}
	apimeta.SetStatusCondition(&subject.Status.Conditions, metav1.Condition{
		Type:   userv1.ConditionRedeemed,
		Status: metav1.ConditionTrue,
	})

	c := prepareTest(t, subject)
	_, err := invitationRedeemReconciler(c).Reconcile(ctx, requestFor(subject))
	require.ErrorContains(t, err, "invitation has no user")
}

func Test_InvitationRedeemReconciler_Reconcile_Fail_UnsupportedTarget(t *testing.T) {
	ctx := context.Background()

	subject := &userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Status: userv1.InvitationStatus{
			RedeemedBy: "user",
			TargetStatuses: []userv1.TargetStatus{
				{TargetRef: userv1.TargetRef{Kind: "Unsupported"}},
			},
		},
	}
	apimeta.SetStatusCondition(&subject.Status.Conditions, metav1.Condition{
		Type:   userv1.ConditionRedeemed,
		Status: metav1.ConditionTrue,
	})

	c := prepareTest(t, subject)
	_, err := invitationRedeemReconciler(c).Reconcile(ctx, requestFor(subject))
	require.ErrorContains(t, err, "unsupported target")
}

func Test_InvitationRedeemReconciler_Reconcile_Fail_TargetNotFound(t *testing.T) {
	ctx := context.Background()

	team := &controlv1.Team{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "appuio.io/v1",
			Kind:       "Team",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-team-01",
			Namespace: "example-organization-01",
		},
	}

	subject := &userv1.Invitation{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject",
		},
		Status: userv1.InvitationStatus{
			RedeemedBy: "user",
			TargetStatuses: []userv1.TargetStatus{
				{TargetRef: targetRefFromObject(team)},
			},
		},
	}
	apimeta.SetStatusCondition(&subject.Status.Conditions, metav1.Condition{
		Type:   userv1.ConditionRedeemed,
		Status: metav1.ConditionTrue,
	})

	c := prepareTest(t, subject)
	_, err := invitationRedeemReconciler(c).Reconcile(ctx, requestFor(subject))
	require.ErrorContains(t, err, "not found")

	require.NoError(t, c.Get(ctx, client.ObjectKeyFromObject(subject), subject))
	assert.Equal(t, metav1.ConditionFalse, subject.Status.TargetStatuses[0].Condition.Status)
}

func invitationRedeemReconciler(c client.WithWatch) *InvitationRedeemReconciler {
	return &InvitationRedeemReconciler{
		Client:   c,
		Scheme:   c.Scheme(),
		Recorder: record.NewFakeRecorder(3),
	}
}

func targetRefFromObject(obj client.Object) userv1.TargetRef {
	return userv1.TargetRef{
		APIGroup:  obj.GetObjectKind().GroupVersionKind().Group,
		Kind:      obj.GetObjectKind().GroupVersionKind().Kind,
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
}
