package model_test

import (
	"testing"

	json "encoding/json"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/model"
	"github.com/stretchr/testify/require"
)

func Test_Nullable(t *testing.T) {
	type mt struct {
		NullableString model.Nullable[string] `json:"nullable_string"`
	}

	t.Run("null", func(t *testing.T) {
		var m mt
		err := json.Unmarshal([]byte(`{"nullable_string": null}`), &m)
		require.NoError(t, err)
		require.False(t, m.NullableString.Valid)
	})

	t.Run("false", func(t *testing.T) {
		var m mt
		err := json.Unmarshal([]byte(`{"nullable_string": false}`), &m)
		require.NoError(t, err)
		require.False(t, m.NullableString.Valid)

		raw, err := json.Marshal(m)
		require.NoError(t, err)
		require.JSONEq(t, `{"nullable_string": false}`, string(raw))
	})

	t.Run("string", func(t *testing.T) {
		var m mt
		err := json.Unmarshal([]byte(`{"nullable_string": "test"}`), &m)
		require.NoError(t, err)
		require.True(t, m.NullableString.Valid)
		require.Equal(t, "test", m.NullableString.Value)

		raw, err := json.Marshal(m)
		require.NoError(t, err)
		require.JSONEq(t, `{"nullable_string": "test"}`, string(raw))
	})
}
