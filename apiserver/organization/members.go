package organization

import (
	"context"

	controlv1 "github.com/appuio/control-api/apis/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups="appuio.io",resources=organizationmembers,verbs=get;list;watch;create;delete;patch;update;edit

// memberProvider is an abstraction for interacting with the OrganizationMembers Object
//
//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=./mock/$GOFILE
type memberProvider interface {
	CreateMembers(ctx context.Context, members *controlv1.OrganizationMembers) error
}

type kubeMemberProvider struct {
	Client client.Client

	usernamePrefix string
}

func (k kubeMemberProvider) CreateMembers(ctx context.Context, members *controlv1.OrganizationMembers) error {
	return k.Client.Create(ctx, members)
}
