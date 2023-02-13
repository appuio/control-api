package secretstorage

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
)

// NewStatusSubResourceRegisterer returns a helper type to register a status subresource for a resource.
//
//	builder.APIServer.
//		WithResourceAndHandler(&Resource{}, storage).
//		WithResourceAndHandler(StatusSubResourceRegisterer{&Resource{}}, storage).
func NewStatusSubResourceRegisterer(o resource.ObjectWithStatusSubResource) resource.ObjectWithStatusSubResource {
	return statusSubResourceRegisterer{o}
}

type statusSubResourceRegisterer struct {
	resource.ObjectWithStatusSubResource
}

func (o statusSubResourceRegisterer) GetGroupVersionResource() schema.GroupVersionResource {
	gvr := o.ObjectWithStatusSubResource.GetGroupVersionResource()
	gvr.Resource = fmt.Sprintf("%s/status", gvr.Resource)
	return gvr
}
