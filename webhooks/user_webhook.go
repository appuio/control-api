package webhooks

import (
	"context"
	"fmt"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	controlv1 "github.com/appuio/control-api/apis/v1"
	"github.com/appuio/control-api/pkg/sar"
)

// +kubebuilder:webhook:path=/validate-appuio-io-v1-user,mutating=false,failurePolicy=fail,groups="appuio.io",resources=users,verbs=create;update,versions=v1,name=validate-users.appuio.io,admissionReviewVersions=v1,sideEffects=None

// +kubebuilder:rbac:groups=appuio.io,resources=organizationmembers,verbs=get

// UserValidator holds context for the validating admission webhook for users.appuio.io
type UserValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// Handle handles the users.appuio.io admission requests
func (v *UserValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := log.FromContext(ctx).WithName("webhook.validate-users.appuio.io")

	// Allow the user to create or update itself.
	// A special permission, used for controllers, `create rbac.appuio.io users` can override this.
	if req.AdmissionRequest.Operation == admissionv1.Create && req.UserInfo.Username != req.Name {
		if err := sar.AuthorizeResource(ctx, v.client, req.UserInfo, sar.ResourceAttributes{
			Verb:     "create",
			Group:    "rbac.appuio.io",
			Resource: req.Resource.Resource,
			Version:  req.Resource.Version,
			Name:     req.Name,
		}); err != nil {
			log.Info("User not authorized to create other users", "request_user", req.AdmissionRequest.UserInfo, "user", req.Name, "error", err)
			return admission.Denied(fmt.Sprintf("user %q is not allowed to create or update %q", req.UserInfo.Username, req.Name))
		}
		log.Info("User authorized to create other users", "user", req.AdmissionRequest.UserInfo)
	}

	user := &controlv1.User{}
	if err := v.decoder.Decode(req, user); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.V(1).WithValues("user", user).Info("Validating")

	orgref := user.Spec.Preferences.DefaultOrganizationRef
	if orgref == "" {
		// No default org is a valid config
		return admission.Allowed("user does not have a default organization")
	}
	orgMembKey := types.NamespacedName{
		Name:      "members",
		Namespace: orgref,
	}
	orgmemb := &controlv1.OrganizationMembers{}
	if err := v.client.Get(ctx, orgMembKey, orgmemb); err != nil {
		return admission.Denied(fmt.Sprintf("Unable to load members for organization %s", orgref))
	}

	log.V(1).WithValues("orgref", orgref, "orgmemb", orgmemb).Info("organizationmembers of requested default organization")

	for _, orguser := range orgmemb.Spec.UserRefs {
		if user.Name == orguser.Name {
			return admission.Allowed("user is member of requested default organization")
		}
	}

	return admission.Denied(fmt.Sprintf("User %s isn't member of organization %s", user.Name, orgref))
}

// InjectDecoder injects a Admission request decoder into the UserValidator
func (v *UserValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

// InjectClient injects a Kubernetes client into the UserValidator
func (v *UserValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}
