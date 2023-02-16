package controllers_test

import (
	"testing"

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

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(initObjs...).
		Build()
}

func requestFor(obj client.Object) ctrl.Request {
	return ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: obj.GetName(),
		},
	}
}
