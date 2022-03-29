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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
)

func TestUserValidator_Handle(t *testing.T) {
	ctx := context.Background()
	tests := map[string]struct {
		orgref  string
		org     string
		orgmemb []string
		allowed bool
		errcode int32
	}{
		"UserIsMember allowed": {
			orgref:  "test-org",
			org:     "test-org",
			orgmemb: []string{"test-user", "test-user-2"},
			allowed: true,
			errcode: http.StatusOK,
		},
		"UserIsNotMember denied": {
			orgref:  "test-org",
			org:     "test-org",
			orgmemb: []string{"test-user-2", "test-user-3"},
			allowed: false,
			errcode: http.StatusForbidden,
		},
		"OrgDoesNotExist denied": {
			orgref:  "test-org-2",
			org:     "test-org",
			orgmemb: []string{"test-user"},
			allowed: false,
			errcode: http.StatusForbidden,
		},
		"InvalidRequest denied": {
			orgref:  "",
			org:     "",
			orgmemb: []string{"test-user"},
			allowed: false,
			errcode: http.StatusBadRequest,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			user := controlv1.User{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-user",
				},
				Spec: controlv1.UserSpec{
					Preferences: controlv1.UserPreferences{
						DefaultOrganizationRef: tc.orgref,
					},
				},
			}

			userRefs := []controlv1.UserRef{}
			for _, uname := range tc.orgmemb {
				userRefs = append(userRefs, controlv1.UserRef{Name: uname})
			}
			orgmemb := controlv1.OrganizationMembers{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "members",
					Namespace: tc.org,
				},
				Spec: controlv1.OrganizationMembersSpec{
					UserRefs: userRefs,
				},
			}

			uv := prepareTest(t, &user, &orgmemb)

			userJson, err := json.Marshal(user)
			require.NoError(t, err)
			// Break admission request object JSON for
			// InvalidRequest testcase
			if tc.errcode == 400 {
				userJson[10] = 'x'
			}
			userObj := runtime.RawExtension{
				Raw: userJson,
			}

			admissionRequest := admission.Request{
				AdmissionRequest: admissionv1.AdmissionRequest{
					UID: "e515f52d-7181-494d-a3d3-f0738856bd97",
					Kind: metav1.GroupVersionKind{
						Group:   "appuio.io",
						Version: "v1",
						Kind:    "User",
					},
					Resource: metav1.GroupVersionResource{
						Group:    "appuio.io",
						Version:  "v1",
						Resource: "users",
					},
					Name:      "test-user",
					Operation: admissionv1.Update,
					UserInfo: authenticationv1.UserInfo{
						Username: "kubernetes-admin",
						Groups: []string{
							"system:masters",
							"system:authenticated",
						},
					},
					Object: userObj,
				},
			}

			resp := uv.Handle(ctx, admissionRequest)

			assert.Equal(t, tc.allowed, resp.Allowed)
			assert.Equal(t, tc.errcode, resp.Result.Code)
		})
	}
}

func prepareTest(t *testing.T, initObjs ...client.Object) *UserValidator {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(orgv1.AddToScheme(scheme))
	utilruntime.Must(controlv1.AddToScheme(scheme))

	decoder, err := admission.NewDecoder(scheme)
	require.NoError(t, err)

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		Build()

	uv := &UserValidator{}
	uv.InjectClient(client)
	uv.InjectDecoder(decoder)

	return uv
}
