package organization

import (
	"context"
	"fmt"

	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
)

// +kubebuilder:rbac:groups="",resources=users;groups;serviceaccounts,verbs=impersonate

// impersonator can build a client that impersonates a user
type impersonator interface {
	Impersonate(u user.Info) (client.Client, error)
}

// impersonatorFromRestconf can build a client that impersonates a user
// from a rest.Config and client.Options
type impersonatorFromRestconf struct {
	config *restclient.Config
	opts   client.Options
}

var _ impersonator = impersonatorFromRestconf{}

// Impersonate returns a client that impersonates the given user
func (c impersonatorFromRestconf) Impersonate(u user.Info) (client.Client, error) {
	conf := restclient.CopyConfig(c.config)

	conf.Impersonate = restclient.ImpersonationConfig{
		UserName: u.GetName(),
		UID:      u.GetUID(),
		Groups:   u.GetGroups(),
		Extra:    u.GetExtra(),
	}
	return client.New(conf, c.opts)
}

// billingEntityValidator validates that the billing entity exists and the requesting user has access to it.
// it does so by impersonating the user and trying to get the billing entity.
func (s *organizationStorage) billingEntityValidator(ctx context.Context, org, oldOrg *orgv1.Organization) error {
	// check if changed
	if oldOrg != nil && oldOrg.Spec.BillingEntityRef == org.Spec.BillingEntityRef {
		return nil
	}
	// check if we allow empty billing entities
	if org.Spec.BillingEntityRef == "" && s.allowEmptyBillingEntity {
		return nil
	}

	user, ok := request.UserFrom(ctx)
	if !ok {
		return fmt.Errorf("no user in context")
	}

	var be billingv1.BillingEntity
	c, err := s.impersonator.Impersonate(user)
	if err != nil {
		return fmt.Errorf("failed to impersonate user: %w", err)
	}

	if err := c.Get(ctx, client.ObjectKey{Name: org.Spec.BillingEntityRef}, &be); err != nil {
		return err
	}

	return nil
}
