package organization

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/authorization/authorizer"
	"sigs.k8s.io/apiserver-runtime/pkg/util/loopback"
)

type loopbackNamespaceAuthorizer struct{}

func (auth loopbackNamespaceAuthorizer) Authorize(ctx context.Context, attr authorizer.Attributes) (authorized authorizer.Decision, reason string, err error) {
	groupVersionResource := corev1.SchemeGroupVersion.WithResource("namespaces")

	return loopback.GetAuthorizer().Authorize(ctx, authorizer.AttributesRecord{
		User:            attr.GetUser(),
		Verb:            attr.GetVerb(),
		Name:            orgNameToNamespaceName(attr.GetName()),
		Namespace:       attr.GetNamespace(),
		APIGroup:        groupVersionResource.Group,
		APIVersion:      groupVersionResource.Version,
		Resource:        groupVersionResource.Resource,
		Subresource:     attr.GetSubresource(),
		ResourceRequest: attr.IsResourceRequest(),
		Path:            attr.GetPath(),
	})
}
