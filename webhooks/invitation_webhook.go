package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/multierr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	"github.com/appuio/control-api/controllers/targetref"
)

// +kubebuilder:webhook:path=/validate-user-appuio-io-v1-invitation,mutating=false,failurePolicy=fail,groups="user.appuio.io",resources=invitations,verbs=create;update,versions=v1,name=validate-invitations.user.appuio.io,admissionReviewVersions=v1,sideEffects=None

// InvitationValidator holds context for the validating admission webhook for users.appuio.io
type InvitationValidator struct {
	client  client.Client
	decoder *admission.Decoder
}

// Handle handles the users.appuio.io admission requests
func (v *InvitationValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := log.FromContext(ctx).WithName("validate-invitations.user.appuio.io")

	inv := &userv1.Invitation{}
	if err := v.decoder.Decode(req, inv); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.V(4).WithValues("invitation", inv).Info("Validating")

	username := req.UserInfo.Username
	authErrors := make([]error, 0, len(inv.Spec.TargetRefs))
	for _, target := range inv.Spec.TargetRefs {
		authErrors = append(authErrors, authorizeTarget(ctx, v.client, username, target))
	}
	if err := multierr.Combine(authErrors...); err != nil {
		return admission.Denied(fmt.Sprintf("user %q is not allowed to invite to the targets: %s", username, err))
	}

	return admission.Allowed("target refs are valid")
}

// InjectDecoder injects a Admission request decoder into the InvitationValidator
func (v *InvitationValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
}

// InjectClient injects a Kubernetes client into the InvitationValidator
func (v *InvitationValidator) InjectClient(c client.Client) error {
	v.client = c
	return nil
}

func authorizeTarget(ctx context.Context, c client.Client, user string, target userv1.TargetRef) error {
	o, err := targetref.GetTarget(ctx, c, target)
	if err != nil {
		return err
	}

	a, err := targetref.NewUserAccessor(o)
	if err != nil {
		return err
	}

	if a.HasUser(user) {
		return nil
	}

	return fmt.Errorf("target %q.%q/%q in namespace %q is not allowed", target.APIGroup, target.Kind, target.Name, target.Namespace)
}
