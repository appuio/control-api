package organization

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestOrganizationStorage_ConvertToTable(t *testing.T) {
	tests := map[string]struct {
		obj          runtime.Object
		tableOptions runtime.Object
		fail         bool
		nrRows       int
	}{
		"GivenEmptyOrg_ThenSingleRow": {
			obj:    &orgv1.Organization{},
			nrRows: 1,
		},
		"GivenOrg_ThenSingleRow": {
			obj: &orgv1.Organization{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: orgv1.OrganizationSpec{
					DisplayName: "bar",
				},
			},
			nrRows: 1,
		},
		"GivenOrgList_ThenMultipleRow": {
			obj: &orgv1.OrganizationList{
				Items: []orgv1.Organization{
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
		"GivenNonOrg_ThenFail": {
			obj:  &corev1.Pod{},
			fail: true,
		},
		"GivenNonOrgList_ThenFail": {
			obj:  &corev1.PodList{},
			fail: true,
		},
	}
	orgStore := &organizationStorage{}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			table, err := orgStore.ConvertToTable(context.TODO(), tc.obj, tc.tableOptions)
			if tc.fail {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, table.Rows, tc.nrRows)
		})
	}
}
