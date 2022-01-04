package organization

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/apiserver-runtime/pkg/util/loopback"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// namespaceProvider is an abstraction for interacting with the Kubernetes API
//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE
type namespaceProvider interface {
	GetNamespace(ctx context.Context, name string, options *metav1.GetOptions) (*corev1.Namespace, error)
	DeleteNamespace(ctx context.Context, name string, options *metav1.DeleteOptions) (*corev1.Namespace, error)
	CreateNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.CreateOptions) error
	UpdateNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.UpdateOptions) error
	ListNamespaces(ctx context.Context, options *metainternalversion.ListOptions) (*corev1.NamespaceList, error)
	WatchNamespaces(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error)
}

type loopbackNamespaceProvider struct {
	initOnce sync.Once
	client   client.WithWatch
}

func (p *loopbackNamespaceProvider) init() error {
	// The LoopbackMasterClientConfig is initialized lazily by the runtime
	// We initialize the client once from which ever method is called first
	var err error
	p.initOnce.Do(func() {
		if p.client == nil {
			p.client, err = client.NewWithWatch(loopback.GetLoopbackMasterClientConfig(), client.Options{})
		}
	})
	return err
}

func (p *loopbackNamespaceProvider) GetNamespace(ctx context.Context, name string, options *metav1.GetOptions) (*corev1.Namespace, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}
	ns := corev1.Namespace{}
	err = p.client.Get(ctx, types.NamespacedName{Name: name}, &ns)
	return &ns, err
}

func (p *loopbackNamespaceProvider) DeleteNamespace(ctx context.Context, name string, options *metav1.DeleteOptions) (*corev1.Namespace, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}
	ns := corev1.Namespace{}
	ns.Name = name
	err = p.client.Delete(ctx, &ns, &client.DeleteOptions{
		Raw: options,
	})
	return &ns, err
}

func (p *loopbackNamespaceProvider) CreateNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.CreateOptions) error {
	err := p.init()
	if err != nil {
		return err
	}
	return p.client.Create(ctx, ns, &client.CreateOptions{
		Raw: options,
	})
}

func (p *loopbackNamespaceProvider) UpdateNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.UpdateOptions) error {
	err := p.init()
	if err != nil {
		return err
	}
	return p.client.Update(ctx, ns, &client.UpdateOptions{
		Raw: options,
	})
}

func (p *loopbackNamespaceProvider) ListNamespaces(ctx context.Context, options *metainternalversion.ListOptions) (*corev1.NamespaceList, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}
	nl := corev1.NamespaceList{}
	err = p.client.List(ctx, &nl, &client.ListOptions{
		LabelSelector: options.LabelSelector,
		FieldSelector: options.FieldSelector,
		Limit:         options.Limit,
		Continue:      options.Continue,
	})
	if err != nil {
		return nil, err
	}
	return &nl, nil
}

func (p *loopbackNamespaceProvider) WatchNamespaces(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}
	nl := corev1.NamespaceList{}
	return p.client.Watch(ctx, &nl, &client.ListOptions{
		LabelSelector: options.LabelSelector,
		FieldSelector: options.FieldSelector,
		Limit:         options.Limit,
		Continue:      options.Continue,
	})
}
