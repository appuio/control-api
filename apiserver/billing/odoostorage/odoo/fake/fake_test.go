package fake_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/fake"
)

func TestFakeStorageE2E(t *testing.T) {
	ctx := context.Background()
	s := fake.NewFakeOdooStorage(false)

	_, err := s.Get(ctx, "be-2345")
	require.ErrorIs(t, err, odoo.ErrNotFound)

	err = s.Update(ctx, &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: "be-2345",
		},
	})
	require.ErrorIs(t, err, odoo.ErrNotFound)

	err = s.Create(ctx, &billingv1.BillingEntity{
		Spec: billingv1.BillingEntitySpec{
			Name: "Test",
		},
	})
	require.NoError(t, err)

	be, err := s.Get(ctx, "be-2345")
	require.NoError(t, err)
	require.Equal(t, "Test", be.Spec.Name)

	err = s.Update(ctx, &billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: "be-2345",
		},
		Spec: billingv1.BillingEntitySpec{
			Name: "Another Test",
		},
	})
	require.NoError(t, err)

	bes, err := s.List(ctx)
	require.NoError(t, err)
	require.Len(t, bes, 1)
	require.Equal(t, "Another Test", bes[0].Spec.Name)
}

func TestFakeStorage_List(t *testing.T) {
	ctx := context.Background()
	s := fake.NewFakeOdooStorage(false)

	_, err := s.List(ctx)
	require.NoError(t, err)

	require.NoError(t, s.Create(ctx, &billingv1.BillingEntity{}))
	require.NoError(t, s.Create(ctx, &billingv1.BillingEntity{}))
	require.NoError(t, s.Create(ctx, &billingv1.BillingEntity{}))

	l, err := s.List(ctx)
	require.NoError(t, err)

	names := make([]string, len(l))
	for i, be := range l {
		names[i] = be.Name
	}
	require.Equal(t, []string{"be-2345", "be-2347", "be-2349"}, names)
}
