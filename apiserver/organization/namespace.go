package organization

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// namespaceProvider is an abstraction for interacting with the Kubernetes API
//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=./mock/$GOFILE
type namespaceProvider interface {
	GetNamespace(ctx context.Context, name string, options *metav1.GetOptions) (*corev1.Namespace, error)
	DeleteNamespace(ctx context.Context, name string, options *metav1.DeleteOptions) (*corev1.Namespace, error)
	CreateNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.CreateOptions) error
	UpdateNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.UpdateOptions) error
	ListNamespaces(ctx context.Context, options *metainternalversion.ListOptions) (*corev1.NamespaceList, error)
	WatchNamespaces(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error)
}

type kubeNamespaceProvider struct {
	Client client.WithWatch
}

func (p *kubeNamespaceProvider) GetNamespace(ctx context.Context, name string, options *metav1.GetOptions) (*corev1.Namespace, error) {
	ns := corev1.Namespace{}
	err := p.Client.Get(ctx, types.NamespacedName{Name: name}, &ns)
	return &ns, err
}

func (p *kubeNamespaceProvider) DeleteNamespace(ctx context.Context, name string, options *metav1.DeleteOptions) (*corev1.Namespace, error) {
	ns := corev1.Namespace{}
	ns.Name = name
	err := p.Client.Delete(ctx, &ns, &client.DeleteOptions{
		Raw: options,
	})
	return &ns, err
}

func (p *kubeNamespaceProvider) CreateNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.CreateOptions) error {
	return p.Client.Create(ctx, ns, &client.CreateOptions{
		Raw: options,
	})
}

func (p *kubeNamespaceProvider) UpdateNamespace(ctx context.Context, ns *corev1.Namespace, options *metav1.UpdateOptions) error {
	return p.Client.Update(ctx, ns, &client.UpdateOptions{
		Raw: options,
	})
}

func (p *kubeNamespaceProvider) ListNamespaces(ctx context.Context, options *metainternalversion.ListOptions) (*corev1.NamespaceList, error) {
	nl := corev1.NamespaceList{}
	err := p.Client.List(ctx, &nl, &client.ListOptions{
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

func (p *kubeNamespaceProvider) WatchNamespaces(ctx context.Context, options *metainternalversion.ListOptions) (watch.Interface, error) {
	nl := corev1.NamespaceList{}
	return p.Client.Watch(ctx, &nl, &client.ListOptions{
		LabelSelector: options.LabelSelector,
		FieldSelector: options.FieldSelector,
		Limit:         options.Limit,
		Continue:      options.Continue,
	})
}
