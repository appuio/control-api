package model_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/model"
)

func TestOdooCompositeIDMarshal(t *testing.T) {
	subject := model.OdooCompositeID{Valid: true, ID: 2, Name: "10 Days"}
	marshalled, err := json.Marshal(subject)
	require.NoError(t, err)
	require.JSONEq(t, fmt.Sprintf(`%d`, subject.ID), string(marshalled))

	subject = model.OdooCompositeID{Valid: false}
	marshalled, err = json.Marshal(subject)
	require.NoError(t, err)
	require.JSONEq(t, "false", string(marshalled))
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
