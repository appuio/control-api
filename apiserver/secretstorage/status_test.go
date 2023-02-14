package secretstorage

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/appuio/control-api/apiserver/testresource"
)

func TestStatusSubResourceRegisterer(t *testing.T) {
	obj := &testresource.TestResourceWithStatus{}
	gvr := statusSubResourceRegisterer{obj}.GetGroupVersionResource()
	require.Equal(t, obj.GetGroupVersionResource().Resource+"/status", gvr.Resource)
}
