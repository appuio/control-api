package model_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/model"
)

func TestInvoiceLineTaxIDMarshal(t *testing.T) {
	subject := model.InvoiceLineTaxID{ID: 7}
	marshalled, err := json.Marshal(subject)
	require.NoError(t, err)

	require.JSONEq(t, string(marshalled), fmt.Sprintf(`[6,false,[%d]]`, subject.ID))

	var unmarshalled model.InvoiceLineTaxID
	require.NoError(t, json.Unmarshal(marshalled, &unmarshalled))
	require.Equal(t, subject.ID, unmarshalled.ID)
}

func TestInvoiceLineTaxIDUnmarshal(t *testing.T) {
	tests := []struct {
		raw  string
		errf require.ErrorAssertionFunc
	}{
		{`[6,false,[33]]`, require.NoError},
		{`[6.3,false,[33]]`, require.Error},
		{`""`, require.Error},
		{`[6,false,[33],7]`, require.Error},
		{`["6",false,[33]]`, require.Error},
		{`[6,"false",[33]]`, require.Error},
		{`[6,false,44]`, require.Error},
		{`[6,false,["33"]]`, require.Error},
		{`[6,false,[33,3]]`, require.Error},
		{`[6,false,[33.3]]`, require.Error},
	}

	for _, testCase := range tests {
		var unmarshalled model.InvoiceLineTaxID
		err := json.Unmarshal([]byte(testCase.raw), &unmarshalled)
		testCase.errf(t, err)
	}
}
