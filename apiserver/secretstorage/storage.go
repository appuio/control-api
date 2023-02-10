// Package secretstorage implements a storage backend for resources implementing apiserver-runtime's resource.Object interface.
// The storage backend stores the object in a kubernetes secret.
// The secret is named after the object and the object is stored in the secret's data field.
// Warning: Not all features of the storage backend are implemented.
// Missing features:
// - Field selectors
// - Label selectors
// - you tell me
// UID, CreationTimestamp and ResourceVersion are taken from the secret's metadata.
// UID are namespaced UUIDs, generated from the object's UID and a fixed random UUID as the namespace.
package secretstorage

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/server/storage"
	"sigs.k8s.io/apiserver-runtime/pkg/builder/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	secretObjectKey = "object"
)

var uuidNamespace = uuid.MustParse("5FC62D70-B17C-44A2-9BBB-6B3DD71C4A2E")

type secretStorage struct {
	// object is the type of object this storage is for
	object resource.Object
	// client is the client used to interact with the Kubernetes API.
	client client.WithWatch
	// codec is the codec used to encode and decode the object.
	codec runtime.Codec
	// namespace is the namespace to store the secrets in.
	namespace string
}

// NewStorage creates a new storage for the given object.
func NewStorage(object resource.Object, cc client.WithWatch, backingNS string) (rest.StandardStorage, error) {
	// Supporting namespaced objects would need some way to create a unique hash out of the namespace and name.
	// k8s.io/apiserver/pkg/endpoints/request.NamespaceFrom(ctx)
	if object.NamespaceScoped() {
		return nil, fmt.Errorf("namespace scoped objects are not yet supported")
	}

	scheme := cc.Scheme()

	vs := scheme.PrioritizedVersionsForGroup(object.GetGroupVersionResource().Group)
	if len(vs) == 0 {
		return nil, fmt.Errorf("no versions registered for group %q", object.GetGroupVersionResource().Group)
	}

	codec, _, err := storage.NewStorageCodec(storage.StorageCodecConfig{
		StorageMediaType:  runtime.ContentTypeJSON,
		StorageSerializer: serializer.NewCodecFactory(scheme),
		StorageVersion:    vs[0],
		MemoryVersion:     vs[0],
	})
	if err != nil {
		return nil, err
	}

	return &secretStorage{
		object:    object,
		client:    cc,
		codec:     codec,
		namespace: backingNS,
	}, nil
}

// New implements rest.Storage
func (s *secretStorage) New() runtime.Object {
	return s.object.New()
}

func (s *secretStorage) NewList() runtime.Object {
	return s.object.NewList()
}

// Destroy implements rest.Storage
func (s *secretStorage) Destroy() {}

// NamespaceScoped implements rest.Scoper
func (s *secretStorage) NamespaceScoped() bool {
	return s.object.NamespaceScoped()
}

func (s *secretStorage) Create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, opts *metav1.CreateOptions) (runtime.Object, error) {
	return s.create(ctx, obj, createValidation, opts)
}

func (s *secretStorage) ConvertToTable(ctx context.Context, obj runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return rest.NewDefaultTableConvertor(s.object.GetGroupVersionResource().GroupResource()).ConvertToTable(ctx, obj, tableOptions)
}

func (s *secretStorage) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	obj, err := s.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, false, err
	}

	if deleteValidation != nil {
		if err := deleteValidation(ctx, obj); err != nil {
			return nil, false, fmt.Errorf("failed to validate object: %w", err)
		}
	}

	return obj, true, s.client.Delete(ctx, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: s.getBackingNamespace(),
		},
	}, &client.DeleteOptions{
		GracePeriodSeconds: options.GracePeriodSeconds,
		Preconditions:      options.Preconditions,
		PropagationPolicy:  options.PropagationPolicy,
		DryRun:             options.DryRun,
	})
}

func (s *secretStorage) DeleteCollection(ctx context.Context, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions, listOptions *metainternalversion.ListOptions) (runtime.Object, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *secretStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	rs := corev1.Secret{}

	if err := s.client.Get(ctx, client.ObjectKey{Name: name, Namespace: s.getBackingNamespace()}, &rs); err != nil {
		// Wrapping the not found error breaks kubectl apply (404)
		return nil, err
	}

	return s.objectFromBackingSecret(&rs)
}

func (s *secretStorage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	rsl := &corev1.SecretList{}
	if err := s.client.List(ctx, rsl, &client.ListOptions{
		// TODO(swi): Should we add labels to the secret so we can filter on them?
		// Those would need to be hashed to not introduce collisions with controllers tracking state through labels like argocd does.
		// LabelSelector:         nil,
		// FieldSelector:         nil,
		Namespace: s.getBackingNamespace(),
		Limit:     options.Limit,
		Continue:  options.Continue,
	}); err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	objList := s.object.NewList()
	l, err := apimeta.ExtractList(objList)
	if err != nil {
		return nil, err
	}

	for _, rs := range rsl.Items {
		obj, err := s.objectFromBackingSecret(&rs)
		if err != nil {
			return nil, fmt.Errorf("failed to decode object from secret: %w", err)
		}
		l = append(l, obj)
	}

	if err := apimeta.SetList(objList, l); err != nil {
		return nil, err
	}

	return objList, nil
}

func (s *secretStorage) Watch(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	rsl := &corev1.SecretList{}
	w, err := s.client.Watch(ctx, rsl, &client.ListOptions{
		// TODO(swi): Should we add labels to the secret so we can filter on them?
		// Those would need to be hashed to not introduce collisions with controllers tracking state through labels like argocd does.
		// LabelSelector:         nil,
		// FieldSelector:         nil,
		Namespace: s.getBackingNamespace(),
		Limit:     options.Limit,
		Continue:  options.Continue,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return watch.Filter(w, func(in watch.Event) (out watch.Event, keep bool) {
		if in.Object == nil {
			// This should never happen, let downstream deal with it
			return in, true
		}
		rs, ok := in.Object.(*corev1.Secret)
		if !ok {
			// We received a non Secret object
			// This is most likely an error so we pass it on
			return in, true
		}

		obj, err := s.objectFromBackingSecret(rs)
		if err != nil {
			return watch.Event{
				Type:   watch.Error,
				Object: &metav1.Status{Message: fmt.Sprintf("failed to decode object: %v", err)},
			}, true
		}
		in.Object = obj

		return in, true
	}), nil
}

func (s *secretStorage) Update(
	ctx context.Context, name string,
	objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc,
	updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *metav1.UpdateOptions,
) (runtime.Object, bool, error) {
	var isCreate bool
	rs := &corev1.Secret{}
	err := s.client.Get(ctx, types.NamespacedName{Name: name, Namespace: s.getBackingNamespace()}, rs)
	if err != nil {
		if !forceAllowCreate {
			return nil, false, fmt.Errorf("failed to get old object: %w", err)
		}
		isCreate = true
	}

	oldObj, err := s.objectFromBackingSecret(rs)
	if err != nil {
		return nil, false, fmt.Errorf("failed to decode object: %w", err)
	}

	newObj, err := objInfo.UpdatedObject(ctx, oldObj)
	if err != nil {
		return nil, false, fmt.Errorf("failed to calculate new object: %w", err)
	}

	if isCreate {
		if createValidation != nil {
			if err := createValidation(ctx, newObj); err != nil {
				return nil, false, err
			}
		}
		newObj, err = s.create(ctx, newObj, createValidation, &metav1.CreateOptions{DryRun: options.DryRun})
		return newObj, true, err
	}

	if updateValidation != nil {
		if err := updateValidation(ctx, newObj, oldObj); err != nil {
			return nil, false, fmt.Errorf("failed to validate new object: %w", err)
		}
	}

	newObjRaw := &bytes.Buffer{}
	err = s.codec.Encode(newObj, newObjRaw)
	if err != nil {
		return nil, false, fmt.Errorf("failed to encode new object: %w", err)
	}

	p, err := objectPatch(newObjRaw.Bytes())
	if err != nil {
		return nil, false, fmt.Errorf("failed to create patch: %w", err)
	}

	err = s.client.Patch(ctx, rs, p, &client.PatchOptions{DryRun: options.DryRun})
	if err != nil {
		return newObj, false, fmt.Errorf("failed to update backing secret: %w", err)
	}

	return newObj, false, nil
}

func (s *secretStorage) create(ctx context.Context, obj runtime.Object, createValidation rest.ValidateObjectFunc, opts *metav1.CreateOptions) (runtime.Object, error) {
	ac, err := apimeta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to access object metadata: %w", err)
	}

	if createValidation != nil {
		if err := createValidation(ctx, obj); err != nil {
			return nil, fmt.Errorf("failed to validate object: %w", err)
		}
	}

	raw := &bytes.Buffer{}
	if err := s.codec.Encode(obj, raw); err != nil {
		return nil, fmt.Errorf("failed to encode object: %w", err)
	}
	rs := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ac.GetName(),
			Namespace: s.getBackingNamespace(),
		},
		Data: map[string][]byte{
			secretObjectKey: raw.Bytes(),
		},
	}

	if err := s.client.Create(ctx, &rs, &client.CreateOptions{
		DryRun: opts.DryRun,
	}); err != nil {
		return nil, fmt.Errorf("failed to create secret: %w", err)
	}

	return obj, nil
}

func (s *secretStorage) getBackingNamespace() string {
	if s.namespace != "" {
		return s.namespace
	}
	return "default"
}

func (s *secretStorage) objectFromBackingSecret(rs *corev1.Secret) (runtime.Object, error) {
	obj := s.object.New()
	if _, _, err := s.codec.Decode(rs.Data[secretObjectKey], nil, obj); err != nil {
		return nil, fmt.Errorf("failed to decode object: %w", err)
	}

	ac, err := apimeta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to access object metadata: %w", err)
	}

	// Use the backing secret's creation timestamp and resource version
	// Resource version does not need to be globally unique but needs to change on every update
	ac.SetCreationTimestamp(rs.CreationTimestamp)
	ac.SetResourceVersion(rs.ResourceVersion)
	// Use the backing secret's UID but namespace it to avoid collisions
	// UID should be globally unique and not change on updates (note that OCP sometimes reuses UIDs eg. Project/Namespace)
	if rs.UID != "" {
		// creates a v5 UUID based on the backing secret's UID and a UUID namespace
		ac.SetUID(types.UID(uuid.NewSHA1(uuidNamespace, []byte(rs.UID)).String()))
	} else {
		ac.SetUID("")
	}

	return obj, nil
}

func objectPatch(serialized []byte) (client.Patch, error) {
	jp, err := json.Marshal(map[string]any{
		"data": map[string]any{
			secretObjectKey: base64.StdEncoding.EncodeToString(serialized),
		},
	})
	return client.RawPatch(types.StrategicMergePatchType, jp), err
}
