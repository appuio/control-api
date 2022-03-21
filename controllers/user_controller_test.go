package controllers_test

import (
	"context"
	"errors"
	"testing"

	controlv1 "github.com/appuio/control-api/apis/v1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	. "github.com/appuio/control-api/controllers"
)

func Test_UserController_Reconcile_Success(t *testing.T) {
	ctx := context.Background()

	subject := controlv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject-a",
		},
		Spec: controlv1.UserSpec{
			Preferences: controlv1.UserPreferences{
				DefaultOrganizationRef: "org-a",
			},
		},
	}

	keycloakUser := keycloak.User{
		ID:                     "subject-id",
		Username:               subject.Name,
		Email:                  "subject@email.com",
		FirstName:              "Subject",
		LastName:               "A",
		DefaultOrganizationRef: subject.Spec.Preferences.DefaultOrganizationRef,
	}

	c, keyMock, _ := prepareTest(t, &subject)
	keyMock.EXPECT().
		PutUser(gomock.Any(), keycloak.User{
			Username:               subject.Name,
			DefaultOrganizationRef: subject.Spec.Preferences.DefaultOrganizationRef,
		}).
		Return(keycloakUser, nil).
		Times(1)

	_, err := (&UserReconciler{
		Client:   c,
		Scheme:   &runtime.Scheme{},
		Keycloak: keyMock,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.NoError(t, err)

	reconciledUser := controlv1.User{}
	require.NoError(t, c.Get(ctx, types.NamespacedName{Name: subject.Name}, &reconciledUser))

	require.Equal(t, keycloakUser.ID, reconciledUser.Status.ID)
	require.Equal(t, keycloakUser.Username, reconciledUser.Status.Username)
	require.Equal(t, keycloakUser.Email, reconciledUser.Status.Email)
	require.Equal(t, keycloakUser.DisplayName(), reconciledUser.Status.DisplayName)
	require.Equal(t, keycloakUser.DefaultOrganizationRef, reconciledUser.Status.DefaultOrganizationRef)
}

func Test_UserController_Reconcile_Failure(t *testing.T) {
	ctx := context.Background()

	subject := controlv1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subject-a",
		},
	}

	c, keyMock, erMock := prepareTest(t, &subject)
	keyMock.EXPECT().
		PutUser(gomock.Any(), keycloak.User{
			Username:               subject.Name,
			DefaultOrganizationRef: subject.Spec.Preferences.DefaultOrganizationRef,
		}).
		Return(keycloak.User{}, errors.New("unknown errors")).
		Times(1)
	erMock.EXPECT().
		Event(gomock.Any(), "Warning", "UpdateFailed", gomock.Any()).
		Times(1)

	_, err := (&UserReconciler{
		Client:   c,
		Scheme:   &runtime.Scheme{},
		Keycloak: keyMock,
		Recorder: erMock,
	}).Reconcile(ctx, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name: subject.Name,
		},
	})
	require.Error(t, err)
}
