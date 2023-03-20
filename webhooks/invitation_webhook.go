package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/multierr"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	authorizationv1 "k8s.io/kubernetes/pkg/apis/authorization/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	userv1 "github.com/appuio/control-api/apis/user/v1"
	"github.com/appuio/control-api/controllers/targetref"
)

// +kubebuilder:webhook:path=/validate-user-appuio-io-v1-invitation,mutating=false,failurePolicy=fail,groups="user.appuio.io",resources=invitations,verbs=create;update,versions=v1,name=validate-invitations.user.appuio.io,admissionReviewVersions=v1,sideEffects=None

// +kubebuilder:rbac:groups=authorization.k8s.io,resources=subjectaccessreviews,verbs=create

// InvitationValidator holds context for the validating admission webhook for users.appuio.io
type InvitationValidator struct {
	client  client.Client
	decoder *admission.Decoder

	UsernamePrefix string
}

// Handle handles the users.appuio.io admission requests
func (v *InvitationValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := log.FromContext(ctx).WithName("validate-invitations.user.appuio.io")

	inv := &userv1.Invitation{}
	if err := v.decoder.Decode(req, inv); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.V(1).WithValues("invitation", inv).Info("Validating")

	authErrors := make([]error, 0, len(inv.Spec.TargetRefs))
	for _, target := range inv.Spec.TargetRefs {
		authErrors = append(authErrors, authorizeTarget(ctx, v.client, req.UserInfo, target))
	}
	if err := multierr.Combine(authErrors...); err != nil {
		return admission.Denied(fmt.Sprintf("user %q is not allowed to invite to the targets: %s", req.UserInfo.Username, err))
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

func authorizeTarget(ctx context.Context, c client.Client, user authenticationv1.UserInfo, target userv1.TargetRef) error {
	// Check if the target references a supported resource
	_, err := targetref.NewObjectFromRef(target)
	if err != nil {
		return err
	}

	return canEditTarget(ctx, c, user, target)
}

// canEditTarget checks if the user is allowed to edit the target.
// it does so by creating a SubjectAccessReview to `update` the resource and checking the response.
func canEditTarget(ctx context.Context, c client.Client, user authenticationv1.UserInfo, target userv1.TargetRef) error {
	const verb = "update"

	ra, err := mapTargetRefToResourceAttribute(c, target)
	if err != nil {
		return err
	}
	ra["verb"] = verb

	// I could not find a way to create a SubjectAccessReview object with the client.
	// `no kind "CreateOptions" is registered for the internal version of group "authorization.k8s.io" in scheme`
	// even after installing the authorization scheme.
	rawSAR := &unstructured.Unstructured{
		Object: map[string]any{
			"spec": map[string]any{
				"resourceAttributes": ra,

				"user":   user.Username,
				"groups": user.Groups,
				"uid":    user.UID,
			},
		}}
	rawSAR.SetGroupVersionKind(authorizationv1.SchemeGroupVersion.WithKind("SubjectAccessReview"))

	if err := c.Create(ctx, rawSAR); err != nil {
		return fmt.Errorf("failed to create SubjectAccessReview: %w", err)
	}

	allowed, _, err := unstructured.NestedBool(rawSAR.Object, "status", "allowed")
	if err != nil {
		return fmt.Errorf("failed to get SubjectAccessReview status.allowed: %w", err)
	}

	if !allowed {
		return fmt.Errorf("%q on target %q.%q/%q in namespace %q is not allowed", verb, target.APIGroup, target.Kind, target.Name, target.Namespace)
	}

	return nil
}

func mapTargetRefToResourceAttribute(c client.Client, target userv1.TargetRef) (map[string]any, error) {
	rm, err := c.RESTMapper().RESTMapping(schema.GroupKind{
		Group: target.APIGroup,
		Kind:  target.Kind,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get REST mapping for %q.%q: %w", target.APIGroup, target.Kind, err)
	}

	return map[string]any{
		"group":    target.APIGroup,
		"version":  rm.Resource.Version,
		"resource": rm.Resource.Resource,

		"namespace": target.Namespace,
		"name":      target.Name,
	}, nil
}
