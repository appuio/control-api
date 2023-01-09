package odoostorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
)

func TestBillingEntityStorage_ConvertToTable(t *testing.T) {
	tests := map[string]struct {
		obj          runtime.Object
		tableOptions runtime.Object
		fail         bool
		nrRows       int
	}{
		"GivenEmptyBe_ThenSingleRow": {
			obj:    &billingv1.BillingEntity{},
			nrRows: 1,
		},
		"GivenBe_ThenSingleRow": {
			obj: &billingv1.BillingEntity{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: billingv1.BillingEntitySpec{
					Name: "bar",
				},
			},
			nrRows: 1,
		},
		"GivenBeList_ThenMultipleRow": {
			obj: &billingv1.BillingEntityList{
				Items: []billingv1.BillingEntity{
					{},
					{},
					{},
				},
			},
			nrRows: 3,
		},
		"GivenNil_ThenFail": {
			obj:  nil,
			fail: true,
		},
		"GivenNonBe_ThenFail": {
			obj:  &corev1.Pod{},
			fail: true,
		},
		"GivenNonBeList_ThenFail": {
			obj:  &corev1.PodList{},
			fail: true,
		},
	}
	beStore := &billingEntityStorage{}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			table, err := beStore.ConvertToTable(context.TODO(), tc.obj, tc.tableOptions)
			if tc.fail {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, table.Rows, tc.nrRows)
		})
	}
}
