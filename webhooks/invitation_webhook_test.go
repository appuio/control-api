package webhooks

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/apis/authorization"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
)

func TestInvitationValidator_Handle(t *testing.T) {
	// allowed user is a member of possible targets
	const allowedUser = "allowed-user"
	// denied user is not a member of possible targets
	const deniedUser = "denied-user"

	const (
		testOrg      = "foo-org"
		testTeam     = "foo-team"
		testRoleName = "foo-role"
	)

	tests := map[string]struct {
		requestUser string
		targets     []userv1.TargetRef

		allowed bool
		errcode int32
	}{
		"empty is allowed": {
			requestUser: deniedUser,
			targets:     []userv1.TargetRef{},
			allowed:     true,
			errcode:     http.StatusOK,
		},

		"unknown target is denied": {
			requestUser: allowedUser,
			targets: []userv1.TargetRef{
				{
					APIGroup:  "custom.io",
					Kind:      "MyCustomKind",
					Namespace: testOrg,
					Name:      "custom",
				},
			},
			allowed: false,
			errcode: http.StatusForbidden,
		},

		"invalid request": {
			requestUser: allowedUser,
			targets:     []userv1.TargetRef{},
			allowed:     false,
			errcode:     http.StatusBadRequest,
		},

		"OrganizationMembers allowed": {
			requestUser: allowedUser,
			targets: []userv1.TargetRef{
				{
					APIGroup:  "appuio.io",
					Kind:      "OrganizationMembers",
					Namespace: testOrg,
					Name:      "members",
				},
			},
			allowed: true,
			errcode: http.StatusOK,
		},

		"OrganizationMembers denied": {
			requestUser: deniedUser,
			targets: []userv1.TargetRef{
				{
					APIGroup:  "appuio.io",
					Kind:      "OrganizationMembers",
					Namespace: testOrg,
					Name:      "members",
				},
			},
			allowed: false,
			errcode: http.StatusForbidden,
		},

		"Team allowed": {
			requestUser: allowedUser,
			targets: []userv1.TargetRef{
				{
					APIGroup:  "appuio.io",
					Kind:      "Team",
					Namespace: testOrg,
					Name:      testTeam,
				},
			},
			allowed: true,
			errcode: http.StatusOK,
		},

		"Team denied": {
			requestUser: deniedUser,
			targets: []userv1.TargetRef{
				{
					APIGroup:  "appuio.io",
					Kind:      "Team",
					Namespace: testOrg,
					Name:      testTeam,
				},
			},
			allowed: false,
			errcode: http.StatusForbidden,
		},

		"ClusterRoleBinding allowed": {
			requestUser: allowedUser,
			targets: []userv1.TargetRef{
				{
					APIGroup: rbacv1.GroupName,
					Kind:     "ClusterRoleBinding",
					Name:     testRoleName,
				},
			},
			allowed: true,
			errcode: http.StatusOK,
		},

		"ClusterRoleBinding denied": {
			requestUser: deniedUser,
			targets: []userv1.TargetRef{
				{
					APIGroup: rbacv1.GroupName,
					Kind:     "ClusterRoleBinding",
					Name:     testRoleName,
				},
			},
			allowed: false,
			errcode: http.StatusForbidden,
		},

		"RoleBinding allowed": {
			requestUser: allowedUser,
			targets: []userv1.TargetRef{
				{
					APIGroup:  rbacv1.GroupName,
					Kind:      "RoleBinding",
					Namespace: testOrg,
					Name:      testRoleName,
				},
			},
			allowed: true,
			errcode: http.StatusOK,
		},

		"RoleBinding denied": {
			requestUser: deniedUser,
			targets: []userv1.TargetRef{
				{
					APIGroup:  rbacv1.GroupName,
					Kind:      "RoleBinding",
					Namespace: testOrg,
					Name:      testRoleName,
				},
			},
			allowed: false,
			errcode: http.StatusForbidden,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			invitation := userv1.Invitation{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-invitation",
				},
				Spec: userv1.InvitationSpec{
					TargetRefs: tc.targets,
				},
			}

			org := controlv1.OrganizationMembers{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "members",
					Namespace: testOrg,
				},
				Spec: controlv1.OrganizationMembersSpec{
					UserRefs: []controlv1.UserRef{{Name: allowedUser}},
				},
			}
			team := controlv1.Team{
				ObjectMeta: metav1.ObjectMeta{
					Name:      testTeam,
					Namespace: testOrg,
				},
				Spec: controlv1.TeamSpec{
					UserRefs: []controlv1.UserRef{{Name: allowedUser}},
				},
			}
			rb := rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: testOrg,
					Name:      testRoleName,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     rbacv1.UserKind,
						APIGroup: rbacv1.GroupName,
						Name:     allowedUser,
					},
				},
			}
			crb := rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: testRoleName,
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:     rbacv1.UserKind,
						APIGroup: rbacv1.GroupName,
						Name:     allowedUser,
					},
				},
			}

			iv := prepareInvitationValidatorTest(t, allowedUser, &invitation, &org, &team, &rb, &crb)

			invJson, err := json.Marshal(invitation)
			require.NoError(t, err)
			// Break admission request object JSON for
			// InvalidRequest testcase
			if tc.errcode == http.StatusBadRequest {
				invJson[10] = 'x'
			}
			invObj := runtime.RawExtension{
				Raw: invJson,
			}

			admissionRequest := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "e515f52d-7181-494d-a3d3-f0738856bd97",
					Kind: metav1.GroupVersionKind{
						Group:   "appuio.io",
						Version: "v1",
						Kind:    "Invitation",
					},
					Resource: metav1.GroupVersionResource{
						Group:    "user.appuio.io",
						Version:  "v1",
						Resource: "invitations",
					},
					Name:      "test-user",
					Operation: admissionv1.Update,
					UserInfo: authenticationv1.UserInfo{
						Username: tc.requestUser,
					},
					Object: invObj,
				},
			}

			resp := iv.Handle(context.Background(), admissionRequest)

			assert.Equal(t, tc.allowed, resp.Allowed)
			assert.Equal(t, tc.errcode, resp.Result.Code)
		})
	}
}

func prepareInvitationValidatorTest(t *testing.T, sarAllowedUser string, initObjs ...client.Object) *InvitationValidator {
	scheme := runtime.NewScheme()
	require.NoError(t, clientgoscheme.AddToScheme(scheme))
	require.NoError(t, orgv1.AddToScheme(scheme))
	require.NoError(t, controlv1.AddToScheme(scheme))
	require.NoError(t, userv1.AddToScheme(scheme))
	require.NoError(t, authorization.AddToScheme(scheme))

	decoder, err := admission.NewDecoder(scheme)
	require.NoError(t, err)

	// This would be auto-discovered by the real client
	drm := meta.NewDefaultRESTMapper(
		[]schema.GroupVersion{
			{Group: "appuio.io", Version: "v1"},
			{Group: "rbac.authorization.k8s.io", Version: "v1"},
		},
	)
	drm.AddSpecific(schema.GroupVersionKind{
		Group:   "appuio.io",
		Version: "v1",
		Kind:    "OrganizationMembers",
	}, schema.GroupVersionResource{
		Group:    "appuio.io",
		Version:  "v1",
		Resource: "organizationmembers",
	}, schema.GroupVersionResource{
		Group:    "appuio.io",
		Version:  "v1",
		Resource: "organizationmembers",
	}, meta.RESTScopeNamespace)

	drm.Add(schema.GroupVersionKind{
		Group:   "appuio.io",
		Version: "v1",
		Kind:    "Team",
	}, meta.RESTScopeNamespace)
	drm.Add(schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "ClusterRoleBinding",
	}, meta.RESTScopeRoot)
	drm.Add(schema.GroupVersionKind{
		Group:   "rbac.authorization.k8s.io",
		Version: "v1",
		Kind:    "RoleBinding",
	}, meta.RESTScopeNamespace)

	tr := subjectAccessReviewResponder{
		ktesting.NewObjectTracker(scheme, clientgoscheme.Codecs.UniversalDecoder()),
		sarAllowedUser,
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		WithRESTMapper(drm).
		WithObjectTracker(tr).
		Build()

	iv := &InvitationValidator{}
	iv.InjectClient(client)
	iv.InjectDecoder(decoder)

	return iv
}

// subjectAccessReviewResponder is a wrapper for testing.ObjectTracker that responds to SubjectAccessReview create requests
// and allows or denies the request based on the allowedUser name.
type subjectAccessReviewResponder struct {
	ktesting.ObjectTracker

	allowedUser string
}

func (r subjectAccessReviewResponder) Create(gvr schema.GroupVersionResource, obj runtime.Object, ns string) error {
	if sar, ok := obj.(*authorization.SubjectAccessReview); ok {
		sar.Status.Allowed = sar.Spec.User == r.allowedUser
		return nil
	}
	return r.ObjectTracker.Create(gvr, obj, ns)
}
