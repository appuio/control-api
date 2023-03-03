package main

import (
	"flag"
	"os"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/spf13/cobra"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"

	"github.com/appuio/control-api/controllers"
	"github.com/appuio/control-api/webhooks"
	//+kubebuilder:scaffold:imports
)

// ControllerCommand creates a new command allowing to start the controller
func ControllerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "controller",
	}

	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := zap.Options{}
	opts.BindFlags(zapfs)
	cmd.Flags().AddGoFlagSet(zapfs)

	metricsAddr := cmd.Flags().String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	enableLeaderElection := cmd.Flags().Bool("leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	probeAddr := cmd.Flags().String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	usernamePrefix := cmd.Flags().String("username-prefix", "", "Prefix prepended to username claims. Usually the same as \"--oidc-username-prefix\" of the Kubernetes API server")
	rolePrefix := cmd.Flags().String("role-prefix", "control-api:user:", "Prefix prepended to generated cluster roles and bindings to prevent name collisions.")
	memberRoles := cmd.Flags().StringSlice("member-roles", []string{}, "ClusterRoles to assign to every organization member for its namespace")
	webhookCertDir := cmd.Flags().String("webhook-cert-dir", "", "Directory holding TLS certificate and key for the webhook server. If left empty, {TempDir}/k8s-webhook-server/serving-certs is used")
	webhookPort := cmd.Flags().Int("webhook-port", 9443, "The port on which the admission webhooks are served")

	beRefreshInterval := cmd.Flags().Duration("billing-entity-refresh-interval", 5*time.Minute, "The interval at which the billing entity cache is refreshed")
	beRefreshJitter := cmd.Flags().Duration("billing-entity-refresh-jitter", time.Minute, "The jitter added to the interval at which the billing entity cache is refreshed")

	invTokenValidFor := cmd.Flags().Duration("invitation-valid-for", 30*24*time.Hour, "The duration an invitation token is valid for")
	redeemedInvitationTTL := cmd.Flags().Duration("redeemed-invitation-ttl", 30*24*time.Hour, "The duration for which a redeemed invitation is kept before deleting it")

	cmd.Run = func(*cobra.Command, []string) {
		scheme := runtime.NewScheme()
		setupLog := ctrl.Log.WithName("setup")

		utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		utilruntime.Must(orgv1.AddToScheme(scheme))
		utilruntime.Must(controlv1.AddToScheme(scheme))
		utilruntime.Must(billingv1.AddToScheme(scheme))
		utilruntime.Must(userv1.AddToScheme(scheme))
		//+kubebuilder:scaffold:scheme

		ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
		ctx := ctrl.SetupSignalHandler()

		mgr, err := setupManager(
			*usernamePrefix,
			*rolePrefix,
			*memberRoles,
			*beRefreshInterval,
			*beRefreshJitter,
			*invTokenValidFor,
			*redeemedInvitationTTL,
			ctrl.Options{
				Scheme:                 scheme,
				MetricsBindAddress:     *metricsAddr,
				Port:                   *webhookPort,
				HealthProbeBindAddress: *probeAddr,
				LeaderElection:         *enableLeaderElection,
				LeaderElectionID:       "d9e2acbf.control-api.appuio.io",
				CertDir:                *webhookCertDir,
			})
		if err != nil {
			setupLog.Error(err, "unable to setup manager")
			os.Exit(1)
		}

		setupLog.Info("starting manager")
		if err := mgr.Start(ctx); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
	}

	return cmd
}

func setupManager(usernamePrefix, rolePrefix string, memberRoles []string, beRefreshInterval, beRefreshJitter, invTokenValidFor time.Duration, redeemedInvitationTTL time.Duration, opt ctrl.Options) (ctrl.Manager, error) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), opt)
	if err != nil {
		return nil, err
	}

	metrics.Registry.MustRegister(
		&controllers.OrgBillingRefLinkMetric{
			Client: mgr.GetClient(),
		})

	ur := &controllers.UserReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("user-controller"),

		UserPrefix: usernamePrefix,
		RolePrefix: rolePrefix,
	}
	if err = ur.SetupWithManager(mgr); err != nil {
		return nil, err
	}
	if len(memberRoles) > 0 {
		omr := &controllers.OrganizationMembersReconciler{
			Client:   mgr.GetClient(),
			Scheme:   mgr.GetScheme(),
			Recorder: mgr.GetEventRecorderFor("organization-members-controller"),

			UserPrefix:  usernamePrefix,
			MemberRoles: memberRoles,
		}
		if err = omr.SetupWithManager(mgr); err != nil {
			return nil, err
		}
	}
	obenc := &controllers.OrgBillingEntityNameCacheController{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("organization-billing-entity-name-cache-controller"),

		RefreshInterval: beRefreshInterval,
		RefreshJitter:   beRefreshJitter,
	}
	if err = obenc.SetupWithManager(mgr); err != nil {
		return nil, err
	}
	invtoc := &controllers.InvitationTokenReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("invitation-token-controller"),

		TokenValidFor: invTokenValidFor,
	}
	if err = invtoc.SetupWithManager(mgr); err != nil {
		return nil, err
	}
	invred := &controllers.InvitationRedeemReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("invitation-redeem-controller"),
	}
	if err = invred.SetupWithManager(mgr); err != nil {
		return nil, err
	}
	invclean := &controllers.InvitationCleanupReconciler{
		Client:                mgr.GetClient(),
		Scheme:                mgr.GetScheme(),
		Recorder:              mgr.GetEventRecorderFor("invitation-cleanup-controller"),
		RedeemedInvitationTTL: redeemedInvitationTTL,
	}
	if err = invclean.SetupWithManager(mgr); err != nil {
		return nil, err
	}

	mgr.GetWebhookServer().Register("/validate-appuio-io-v1-user", &webhook.Admission{
		Handler: &webhooks.UserValidator{},
	})
	mgr.GetWebhookServer().Register("/validate-user-appuio-io-v1-invitation", &webhook.Admission{
		Handler: &webhooks.InvitationValidator{},
	})

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return nil, err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return nil, err
	}
	return mgr, err
}
