package model_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/model"
)

func TestOdooCompositeIDMarshal(t *testing.T) {
	subject := model.OdooCompositeID{ID: 2, Name: "10 Days"}
	marshalled, err := json.Marshal(subject)
	require.NoError(t, err)

	require.JSONEq(t, string(marshalled), fmt.Sprintf(`[%d,%q]`, subject.ID, subject.Name))

	var unmarshalled model.OdooCompositeID
	require.NoError(t, json.Unmarshal(marshalled, &unmarshalled))
	require.Equal(t, subject.ID, unmarshalled.ID)
}

func TestOdooCompositeIDUnmarshal(t *testing.T) {
	tests := []struct {
		raw      string
		errf     require.ErrorAssertionFunc
		expected model.OdooCompositeID
	}{
		{
			// false is returned if the field is not set
			raw:      `false`,
			errf:     require.NoError,
			expected: model.OdooCompositeID{Valid: false},
		},
		{
			raw:      `[2,"10 Days"]`,
			errf:     require.NoError,
			expected: model.OdooCompositeID{Valid: true, ID: 2, Name: "10 Days"},
		},
		{
			raw:      `[3245,"Web Marketing Geniuses"]`,
			errf:     require.NoError,
			expected: model.OdooCompositeID{Valid: true, ID: 3245, Name: "Web Marketing Geniuses"},
		},
		{
			raw:  `[2.5,"test"]`,
			errf: require.Error,
		},
		{
			raw:  `""`,
			errf: require.Error,
		},
		{
			raw:  `[2,"10 Days",5]`,
			errf: require.Error,
		},
		{
			raw:  `["2","10 Days"]`,
			errf: require.Error,
		},
		{
			raw:  `[2,10]`,
			errf: require.Error,
		},
	}

	for _, testCase := range tests {
		var unmarshalled model.OdooCompositeID
		err := json.Unmarshal([]byte(testCase.raw), &unmarshalled)
		testCase.errf(t, err)
		require.Equal(t, testCase.expected, unmarshalled)
	}
}
