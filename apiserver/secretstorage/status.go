package secretstorage

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/appuio/control-api/apiserver/secretstorage/status"
)

// NewStatusSubResourceRegisterer returns a helper type to register a status subresource for a resource.
//
//	builder.APIServer.
//		WithResourceAndHandler(&Resource{}, storage).
//		WithResourceAndHandler(StatusSubResourceRegisterer{&Resource{}}, storage).
func NewStatusSubResourceRegisterer(o status.ObjectWithStatusSubResource) status.ObjectWithStatusSubResource {
	return statusSubResourceRegisterer{o}
}

type statusSubResourceRegisterer struct {
	status.ObjectWithStatusSubResource
}

func (o statusSubResourceRegisterer) GetGroupVersionResource() schema.GroupVersionResource {
	gvr := o.ObjectWithStatusSubResource.GetGroupVersionResource()
	gvr.Resource = fmt.Sprintf("%s/status", gvr.Resource)
	return gvr
}
