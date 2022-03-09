package organization

import (
	"context"
	"errors"

	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/endpoints/request"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;delete;patch;update;edit

// Needed so that we are allowed to delegate the default clusterroles
// +kubebuilder:rbac:groups="rbac.appuio.io",resources=organizations,verbs=get;list;watch;create;delete;patch;update;edit
// +kubebuilder:rbac:groups="organization.appuio.io",resources=organizations,verbs=get;list;watch;create;delete;patch;update;edit
// +kubebuilder:rbac:groups="appuio.io",resources=teams,verbs=get;list;watch;create;delete;patch;update

//go:generate go run github.com/golang/mock/mockgen -source=$GOFILE -destination=./mock/$GOFILE
type roleBindingCreator interface {
	CreateRoleBindings(ctx context.Context, namespace string) error
}

type kubeRoleBindingCreator struct {
	Client client.Client

	ClusterRoles []string
}

func (g kubeRoleBindingCreator) CreateRoleBindings(ctx context.Context, namespace string) error {
	for _, cr := range g.ClusterRoles {
		rb, err := generateRoleBinding(ctx, namespace, cr)
		if err != nil {
			return err
		}
		err = g.Client.Create(ctx, rb)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			return err
		}
	}
	return nil
}

func generateRoleBinding(ctx context.Context, namespace, clusterRole string) (*rbacv1.RoleBinding, error) {
	user, ok := request.UserFrom(ctx)
	if !ok {
		return nil, errors.New("unkown user")
	}

	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterRole,
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:     rbacv1.UserKind,
				APIGroup: rbacv1.GroupName,
				Name:     user.GetName(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			APIGroup: rbacv1.GroupName,
			Name:     clusterRole,
		},
	}, nil
}
