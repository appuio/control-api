package main

import (
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/spf13/cobra"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
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

	cmd.Run = func(*cobra.Command, []string) {
		scheme := runtime.NewScheme()
		setupLog := ctrl.Log.WithName("setup")

		utilruntime.Must(clientgoscheme.AddToScheme(scheme))
		utilruntime.Must(orgv1.AddToScheme(scheme))
		utilruntime.Must(controlv1.AddToScheme(scheme))
		//+kubebuilder:scaffold:scheme

		ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
		ctx := ctrl.SetupSignalHandler()

		mgr, err := setupManager(
			*usernamePrefix,
			*rolePrefix,
			*memberRoles,
			ctrl.Options{
				Scheme:                 scheme,
				MetricsBindAddress:     *metricsAddr,
				Port:                   9443,
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

func setupManager(usernamePrefix, rolePrefix string, memberRoles []string, opt ctrl.Options) (ctrl.Manager, error) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), opt)
	if err != nil {
		return nil, err
	}

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

	mgr.GetWebhookServer().Register("/validate-appuio-io-v1-user", &webhook.Admission{
		Handler: &webhooks.UserValidator{},
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
