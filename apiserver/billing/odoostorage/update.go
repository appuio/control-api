package odoostorage

import (
	"context"
	"fmt"
	"reflect"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
)

func (s *billingEntityStorage) Update(ctx context.Context, name string, objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc, updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool, options *metav1.UpdateOptions) (runtime.Object, bool, error) {

	oldBE, err := s.storage.Get(ctx, name)
	if err != nil {
		return nil, false, fmt.Errorf("failed to get old object: %w", err)
	}

	newObj, err := objInfo.UpdatedObject(ctx, oldBE)
	if err != nil {
		// returns a 404 error if the there is no UID present in the object
		return nil, false, fmt.Errorf("failed to calculate new object: %w", err)
	}

	newBE, ok := newObj.(*billingv1.BillingEntity)
	if !ok {
		return nil, false, fmt.Errorf("new object is not an billingentity")
	}

	if updateValidation != nil {
		err = updateValidation(ctx, newBE, oldBE)
		if err != nil {
			return nil, false, fmt.Errorf("failed to validate new object: %w", err)
		}
	}

	if !reflect.DeepEqual(newBE.Spec, oldBE.Spec) {
		apimeta.SetStatusCondition(&newBE.Status.Conditions, metav1.Condition{
			Status: metav1.ConditionFalse,
			Type:   billingv1.ConditionEmailSent,
			Reason: billingv1.ConditionReasonUpdated,
		})
	}

	return newBE, false, s.storage.Update(ctx, newBE)
}
