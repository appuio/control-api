package organization

import (
	"context"
	"fmt"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ rest.Updater = &organizationStorage{}
var _ rest.CreaterUpdater = &organizationStorage{}

func (s *organizationStorage) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {

	oldObj, err := s.Get(ctx, name, nil)
	if err != nil {
		return nil, false, err
	}
	oldOrg, ok := oldObj.(*orgv1.Organization)
	if !ok {
		return nil, false, fmt.Errorf("old object is not an organization")
	}

	newObj, err := objInfo.UpdatedObject(ctx, oldObj)
	if err != nil {
		return nil, false, err
	}
	newOrg, ok := newObj.(*orgv1.Organization)
	if !ok {
		return nil, false, fmt.Errorf("new object is not an organization")
	}

	if updateValidation != nil {
		err = updateValidation(ctx, newOrg, oldObj)
		if err != nil {
			return nil, false, err
		}
	}

	if err := s.billingEntityValidator(ctx, newOrg, oldOrg); err != nil {
		return nil, false, fmt.Errorf("failed to validate billing entity reference: %w", err)
	}

	return newOrg, false, convertNamespaceError(s.namepaces.UpdateNamespace(ctx, newOrg.ToNamespace(), options))
}
