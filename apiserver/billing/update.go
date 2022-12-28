package billingentity

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
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

	oldBE, err := s.storage.Get(ctx, name)
	if err != nil {
		return nil, false, err
	}

	newObj, err := objInfo.UpdatedObject(ctx, oldBE)
	if err != nil {
		return nil, false, err
	}

	newBE, ok := newObj.(*billingv1.BillingEntity)
	if !ok {
		return nil, false, fmt.Errorf("new object is not an billingentity")
	}

	if updateValidation != nil {
		err = updateValidation(ctx, newBE, oldBE)
		if err != nil {
			return nil, false, err
		}
	}

	return newBE, false, s.storage.Update(ctx, newBE)
}
