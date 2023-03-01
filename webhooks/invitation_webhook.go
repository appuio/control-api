package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/multierr"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
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
	switch {
	case target.APIGroup == "appuio.io" && target.Kind == "OrganizationMembers":
		om := controlv1.OrganizationMembers{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &om); err != nil {
			return err
		}
		if isInSlice(om.Spec.UserRefs, controlv1.UserRef{Name: user}) {
			return nil
		}
	case target.APIGroup == "appuio.io" && target.Kind == "Team":
		te := controlv1.Team{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &te); err != nil {
			return err
		}
		if isInSlice(te.Spec.UserRefs, controlv1.UserRef{Name: user}) {
			return nil
		}
	case target.APIGroup == rbacv1.GroupName && target.Kind == "ClusterRoleBinding":
		crb := rbacv1.ClusterRoleBinding{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name}, &crb); err != nil {
			return err
		}
		if isInSlice(crb.Subjects, newSubject(user)) {
			return nil
		}
	case target.APIGroup == rbacv1.GroupName && target.Kind == "RoleBinding":
		rb := rbacv1.RoleBinding{}
		if err := c.Get(ctx, client.ObjectKey{Name: target.Name, Namespace: target.Namespace}, &rb); err != nil {
			return err
		}
		if isInSlice(rb.Subjects, newSubject(user)) {
			return nil
		}
	}

	return fmt.Errorf("target %q.%q/%q in namespace %q is not allowed", target.APIGroup, target.Kind, target.Name, target.Namespace)
}

func isInSlice[T comparable](s []T, e T) (found bool) {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func newSubject(user string) rbacv1.Subject {
	return rbacv1.Subject{
		Kind:     rbacv1.UserKind,
		APIGroup: rbacv1.GroupName,
		Name:     user,
	}
}
