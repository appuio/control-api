package controllers_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
)

func prepareTest(t *testing.T, initObjs ...client.Object) client.WithWatch {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(orgv1.AddToScheme(scheme))
	utilruntime.Must(controlv1.AddToScheme(scheme))
	utilruntime.Must(billingv1.AddToScheme(scheme))
	utilruntime.Must(userv1.AddToScheme(scheme))

	return &fakeSSA{
		fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(initObjs...).
			Build(),
	}
}

// fakeSSA is a fake client that approximates SSA.
// See https://github.com/kubernetes-sigs/controller-runtime/issues/2341#issuecomment-1689885336
// TODO Migrate to builder.WithInterceptorFuncs once controller-runtime is updated to the latest version.
// See https://github.com/kubernetes/kubernetes/issues/115598 for the upstream issue tracking fake SSA support.
type fakeSSA struct {
	client.WithWatch
}

// Patch approximates SSA by creating objects that don't exist yet.
func (f *fakeSSA) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	// Apply patches are supposed to upsert, but fake client fails if the object doesn't exist,
	// if an apply patch occurs for an object that doesn't yet exist, create it.
	if patch.Type() != types.ApplyPatchType {
		return f.WithWatch.Patch(ctx, obj, patch, opts...)
	}
	check, ok := obj.DeepCopyObject().(client.Object)
	if !ok {
		return errors.New("could not check for object in fake client")
	}
	if err := f.WithWatch.Get(ctx, client.ObjectKeyFromObject(obj), check); apierrors.IsNotFound(err) {
		if err := f.WithWatch.Create(ctx, check); err != nil {
			return fmt.Errorf("could not inject object creation for fake: %w", err)
		}
	}
	return f.WithWatch.Patch(ctx, obj, patch, opts...)
}

func requestFor(obj client.Object) ctrl.Request {
	return ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: obj.GetName(),
		},
	}
}
