package organization

import (
	"context"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/apiserver-runtime/pkg/util/loopback"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type loopbackNamespaceProvider struct {
	initOnce sync.Once
	client   client.Client
}

func (p *loopbackNamespaceProvider) init() error {
	var err error
	p.initOnce.Do(func() {
		if p.client == nil {
			p.client, err = client.New(loopback.GetLoopbackMasterClientConfig(), client.Options{})
		}
	})
	return err
}

func (p *loopbackNamespaceProvider) getNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}
	ns := corev1.Namespace{}
	err = p.client.Get(ctx, types.NamespacedName{Name: name}, &ns)
	return &ns, err
}

func (p *loopbackNamespaceProvider) createNamespace(ctx context.Context, ns *corev1.Namespace) error {
	err := p.init()
	if err != nil {
		return err
	}
	return p.client.Create(ctx, ns)
}

func (p *loopbackNamespaceProvider) listNamespaces(ctx context.Context) (*corev1.NamespaceList, error) {
	err := p.init()
	if err != nil {
		return nil, err
	}

	nl := corev1.NamespaceList{}
	err = p.client.List(ctx, &nl, client.MatchingLabels{typeKey: "organization"})
	if err != nil {
		return nil, err
	}
	return &nl, nil
}
