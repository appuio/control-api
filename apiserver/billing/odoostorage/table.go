package odoostorage

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
)

// ConvertToTable translates the given object to a table for kubectl printing
func (s *billingEntityStorage) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	var table metav1.Table

	bes := []billingv1.BillingEntity{}
	if meta.IsListType(obj) {
		beList, ok := obj.(*billingv1.BillingEntityList)
		if !ok {
			return nil, fmt.Errorf("not an billing entity: %#v", obj)
		}
		bes = beList.Items
	} else {
		be, ok := obj.(*billingv1.BillingEntity)
		if !ok {
			return nil, fmt.Errorf("not an billing entity: %#v", obj)
		}
		bes = append(bes, *be)
	}

	for _, be := range bes {
		table.Rows = append(table.Rows, beToTableRow(&be))
	}

	if opt, ok := tableOptions.(*metav1.TableOptions); !ok || !opt.NoHeaders {
		desc := metav1.ObjectMeta{}.SwaggerDoc()
		table.ColumnDefinitions = []metav1.TableColumnDefinition{
			{Name: "Name", Type: "string", Format: "name", Description: desc["name"]},
			{Name: "Display Name", Type: "string", Description: "Name of the billing entity"},
			// {Name: "Age", Type: "date", Description: desc["creationTimestamp"]},
		}
	}
	return &table, nil
}

func beToTableRow(be *billingv1.BillingEntity) metav1.TableRow {
	return metav1.TableRow{
		Cells: []any{
			be.GetName(),
			be.Spec.Name,
			// duration.HumanDuration(time.Since(be.GetCreationTimestamp().Time))
		},
		Object: runtime.RawExtension{Object: be},
	}
}
