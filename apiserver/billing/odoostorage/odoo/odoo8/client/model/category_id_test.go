package model_test

import (
	"encoding/json"
	"testing"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/client/model"
	"github.com/stretchr/testify/require"
)

func TestCategoryIDs_MarshalJSON(t *testing.T) {
	m, err := json.Marshal(model.CategoryIDs{1, 2, 3})

	require.NoError(t, err)
	require.Equal(t, `[[6,false,[1,2,3]]]`, string(m))
}
