package billingentity

import (
	"context"
	"fmt"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"
)

var _ rest.Updater = &billingEntityStorage{}
var _ rest.CreaterUpdater = &billingEntityStorage{}

func (s *billingEntityStorage) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {

	err := s.authorizer.AuthorizeContext(ctx)
	if err != nil {
		return nil, false, err
	}

	newOrg := &billingv1.BillingEntity{}

	oldOrg, err := s.Get(ctx, name, nil)
	if err != nil {

		return nil, false, err
	}

	newObj, err := objInfo.UpdatedObject(ctx, oldOrg)
	if err != nil {
		return nil, false, err
	}

	newOrg, ok := newObj.(*billingv1.BillingEntity)
	if !ok {
		return nil, false, fmt.Errorf("new object is not an billingentity")
	}

	if updateValidation != nil {
		err = updateValidation(ctx, newOrg, oldOrg)
		if err != nil {
			return nil, false, err
		}
	}

	return newOrg, false, apierrors.NewMethodNotSupported((&billingv1.BillingEntity{}).GetGroupVersionResource().GroupResource(), "update")
}
