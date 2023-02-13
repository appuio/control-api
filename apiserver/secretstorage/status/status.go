package status

import "sigs.k8s.io/apiserver-runtime/pkg/builder/resource"

// ObjectWithStatusSubResource defines an interface for getting and setting the status sub-resource for a resource.
// It is a copy of the interface in the apiserver-runtime package, but with the SecretStorage prefix.
// The prefix is needed since the apiserver-runtime package fails if the storage is not a k8s.io/apiserver/pkg/registry/generic/registry.Store.
type ObjectWithStatusSubResource interface {
	resource.Object

	// SecretStorageGetStatus should return the status sub-resource for the resource.
	SecretStorageGetStatus() (statusSubResource StatusSubResource)
}

// StatusSubResource defines required methods for implementing a status subresource.
type StatusSubResource interface {
	resource.SubResource
	// SecretStorageCopyTo copies the content of the status subresource to a parent resource.
	SecretStorageCopyTo(parent ObjectWithStatusSubResource)
}
