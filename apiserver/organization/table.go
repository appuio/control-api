package organization

import (
	"context"
	"fmt"
	"time"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/duration"
)

func (s *organizationStorage) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	var table metav1.Table

	orgs := []orgv1.Organization{}
	if meta.IsListType(obj) {
		orgList, ok := obj.(*orgv1.OrganizationList)
		if !ok {
			return nil, fmt.Errorf("not an organization: %#v", obj)
		}
		orgs = orgList.Items
	} else {
		org, ok := obj.(*orgv1.Organization)
		if !ok {
			return nil, fmt.Errorf("not an organization: %#v", obj)
		}
		orgs = append(orgs, *org)
	}

	for _, org := range orgs {
		table.Rows = append(table.Rows, orgToTableRow(&org))
	}

	if opt, ok := tableOptions.(*metav1.TableOptions); !ok || !opt.NoHeaders {
		desc := metav1.ObjectMeta{}.SwaggerDoc()
		table.ColumnDefinitions = []metav1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name", Description: desc["name"]},
			{Name: "Display Name", Type: "string", Description: "Name of the organization"},
			{Name: "Age", Type: "date", Description: desc["creationTimestamp"]},
		}
	}
	return &table, nil
}

func orgToTableRow(org *orgv1.Organization) metav1.TableRow {
	return metav1.TableRow{
		Cells:  []interface{}{org.GetName(), org.Spec.DisplayName, duration.HumanDuration(time.Since(org.GetCreationTimestamp().Time))},
		Object: runtime.RawExtension{Object: org},
	}

}
