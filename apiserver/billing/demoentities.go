package billingentity

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
)

var demoentities = []client.Object{
	&billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: "be-1245",
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   "Demo Entity 1",
			Emails: []string{"demo1@example.com"},
		},
	},
	&billingv1.BillingEntity{
		ObjectMeta: metav1.ObjectMeta{
			Name: "be-1233",
		},
		Spec: billingv1.BillingEntitySpec{
			Name:   "Demo Entity 2",
			Emails: []string{"demo2@example.com"},
		},
	},
}
