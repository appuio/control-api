package main

import (
	"context"
	"flag"
	"os"
	"text/template"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/Masterminds/sprig/v3"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"

	billingv1 "github.com/appuio/control-api/apis/billing/v1"
	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	userv1 "github.com/appuio/control-api/apis/user/v1"
	controlv1 "github.com/appuio/control-api/apis/v1"
	"github.com/appuio/control-api/controllers/saleorder"
	"github.com/appuio/control-api/mailsenders"

	"github.com/appuio/control-api/controllers"
	"github.com/appuio/control-api/webhooks"
	//+kubebuilder:scaffold:imports
)

const (
	defaultInvitationEmailTemplate = `Hello developer of great software, Kubernetes engineer or fellow human,

A user of APPUiO Cloud has invited you to join them. Follow https://portal.dev/invitations/{{.Object.ObjectMeta.Name}}?token={{.Object.Status.Token}} to accept this invitation.

APPUiO Cloud is a shared Kubernetes offering based on OpenShift provided by https://vshn.ch.

Unsure what to do next? Accept this invitation using the link above, login to one of the zones listed at https://portal.appuio.cloud/zones, deploy your application. A getting started guide on how to do so, is available at https://docs.appuio.cloud/user/tutorials/getting-started.html. To learn more about APPUiO Cloud in general, please visit https://appuio.cloud.

If you have any problems or questions, please email us at support@appuio.ch.

All the best
Your APPUiO Cloud Team`
	defaultBillingEntityEmailTemplate = `Good time of day!

A user of APPUiO Cloud has updated billing entity {{.Object.ObjectMeta.Name}} ({{.Object.Spec.Name}}).

See https://erp.vshn.net/web#id={{ trimPrefix "be-" .Object.ObjectMeta.Name }}&view_type=form&model=res.partner&menu_id=74&action=60 for details.

All the best
Your APPUiO Cloud Team`
)

// ControllerCommand creates a new command allowing to start the controller
func ControllerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "controller",
	}

	zapfs := flag.NewFlagSet("zap", flag.ExitOnError)
	opts := zap.Options{}
	oc := saleorder.Odoo16Credentials{}
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

	invEmailBackend := cmd.Flags().String("email-backend", "stdout", "Backend to use for sending invitation mails (one of stdout, mailgun)")
	invEmailSender := cmd.Flags().String("email-sender", "noreply@appuio.cloud", "Sender address for invitation mails")
	invEmailSubject := cmd.Flags().String("email-subject", "You have been invited to APPUiO Cloud", "Subject for invitation mails")
	emailBodyTemplate := cmd.Flags().String("email-body-template", defaultInvitationEmailTemplate, "Body for invitation mails")
	invEmailBaseRetryDelay := cmd.Flags().Duration("email-base-retry-interval", 15*time.Second, "Retry interval for sending e-mail messages. There is also an exponential back-off applied by the controller.")

	invEmailMailgunToken := cmd.Flags().String("mailgun-token", "CHANGEME", "Token used to access Mailgun API")
	invEmailMailgunDomain := cmd.Flags().String("mailgun-domain", "example.com", "Mailgun Domain to use")
	invEmailMailgunUrl := cmd.Flags().String("mailgun-url", "https://api.eu.mailgun.net/v3", "API base URL for your Mailgun account")
	invEmailMailgunTestMode := cmd.Flags().Bool("mailgun-test-mode", false, "If set, do not actually send e-mails")

	billingEntityEmailBodyTemplate := cmd.Flags().String("billingentity-email-body-template", defaultBillingEntityEmailTemplate, "Body for billing entity modification update mails")
	billingEntityEmailRecipient := cmd.Flags().String("billingentity-email-recipient", "", "Recipient e-mail address for billing entity modification update mails")
	billingEntityEmailSubject := cmd.Flags().String("billingentity-email-subject", "An APPUiO Billing Entity has been updated", "Subject for billing entity modification update mails")
	billingEntityEmailCronInterval := cmd.Flags().String("billingentity-email-cron-interval", "@every 1h", "Cron interval for how frequently billing entity update e-mails are sent")
	billingEntityRBACCronInterval := cmd.Flags().String("billingentity-rbac-cron-interval", "@every 3m", "Cron interval for how frequently billing entity rbac is reconciled")

	saleOrderCompatMode := cmd.Flags().Bool("sale-order-compatibility-mode", false, "Whether to enable compatibility mode for Sales Orders. If enabled, odoo8 billing entity IDs are used to create sales orders in odoo16.")
	saleOrderStorage := cmd.Flags().String("sale-order-storage", "none", "Type of sale order storage to use. Valid values are `none` and `odoo16`")
	saleOrderClientReference := cmd.Flags().String("sale-order-client-reference", "APPUiO Cloud", "Default client reference to add to newly created sales orders.")
	saleOrderInternalNote := cmd.Flags().String("sale-order-internal-note", "auto-generated by APPUiO Cloud Control API", "Default internal note to add to newly created sales orders.")
	cmd.Flags().StringVar(&oc.URL, "sale-order-odoo16-url", "http://localhost:8069", "URL of the Odoo instance to use for sale orders")
	cmd.Flags().StringVar(&oc.Database, "sale-order-odoo16-db", "odooDB", "Database of the Odoo instance to use for sale orders")
	cmd.Flags().StringVar(&oc.Admin, "sale-order-odoo16-account", "Admin", "Odoo Account name to use for sale orders")
	cmd.Flags().StringVar(&oc.Password, "sale-order-odoo16-password", "superSecret1238", "Odoo Account password to use for sale orders")

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

		bt, err := template.New("emailBody").Funcs(sprig.FuncMap()).Parse(*emailBodyTemplate)
		if err != nil {
			setupLog.Error(err, "Failed to parse email body template for invitations")
			os.Exit(1)
		}
		bet, err := template.New("emailBody").Funcs(sprig.FuncMap()).Parse(*billingEntityEmailBodyTemplate)
		if err != nil {
			setupLog.Error(err, "Failed to parse email body template for billing entity e-mails")
			os.Exit(1)
		}
		invitationBodyRenderer := &mailsenders.Renderer{Template: bt}
		billingEntityBodyRenderer := &mailsenders.Renderer{Template: bet}

		var invMailSender mailsenders.MailSender
		var beMailSender mailsenders.MailSender
		if *invEmailBackend == "mailgun" {
			b := mailsenders.NewMailgunSender(
				*invEmailMailgunDomain,
				*invEmailMailgunToken,
				*invEmailMailgunUrl,
				*invEmailSender,
				invitationBodyRenderer,
				*invEmailSubject,
				*invEmailMailgunTestMode,
			)
			invMailSender = &b
			if *billingEntityEmailRecipient != "" {
				be := mailsenders.NewMailgunSender(
					*invEmailMailgunDomain,
					*invEmailMailgunToken,
					*invEmailMailgunUrl,
					*invEmailSender,
					billingEntityBodyRenderer,
					*billingEntityEmailSubject,
					*invEmailMailgunTestMode,
				)
				beMailSender = &be
			} else {
				// fall back to stdout if no recipient e-mail is given
				beMailSender = &mailsenders.StdoutSender{
					Subject: *billingEntityEmailSubject,
					Body:    billingEntityBodyRenderer,
				}
			}
			invMailSender = &b
		} else {
			invMailSender = &mailsenders.StdoutSender{
				Subject: *invEmailSubject,
				Body:    invitationBodyRenderer,
			}
			beMailSender = &mailsenders.StdoutSender{
				Subject: *billingEntityEmailSubject,
				Body:    billingEntityBodyRenderer,
			}
		}

		mgr, err := setupManager(
			*usernamePrefix,
			*rolePrefix,
			*memberRoles,
			*beRefreshInterval,
			*beRefreshJitter,
			*invTokenValidFor,
			*redeemedInvitationTTL,
			*invEmailBaseRetryDelay,
			invMailSender,
			*saleOrderStorage,
			*saleOrderClientReference,
			*saleOrderInternalNote,
			*saleOrderCompatMode,
			oc,
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

		setupLog.Info("setting up email cron")
		emailCron, err := setupEmailCron(
			ctx,
			*billingEntityEmailCronInterval,
			mgr,
			beMailSender,
			*billingEntityEmailRecipient,
		)
		if err != nil {
			setupLog.Error(err, "unable to setup email cron")
			os.Exit(1)
		}
		emailCron.Start()

		setupLog.Info("setting up billing entity rbac cron")
		beRBACCron, err := setupBillingEntityRBACCron(
			ctx,
			*billingEntityRBACCronInterval,
			mgr,
		)
		if err != nil {
			setupLog.Error(err, "unable to setup rbac cron")
			os.Exit(1)
		}
		beRBACCron.Start()

		setupLog.Info("starting manager")
		if err := mgr.Start(ctx); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
		setupLog.Info("Stopping...")
		ecs := emailCron.Stop()
		becs := beRBACCron.Stop()
		<-ecs.Done()
		<-becs.Done()
	}

	return cmd
}

func setupManager(
	usernamePrefix,
	rolePrefix string,
	memberRoles []string,
	beRefreshInterval,
	beRefreshJitter,
	invTokenValidFor time.Duration,
	redeemedInvitationTTL time.Duration,
	invEmailBaseRetryDelay time.Duration,
	mailSender mailsenders.MailSender,
	saleOrderStorage string,
	saleOrderClientReference string,
	saleOrderInternalNote string,
	saleOrderCompatMode bool,
	odooCredentials saleorder.Odoo16Credentials,
	opt ctrl.Options,
) (ctrl.Manager, error) {
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), opt)
	if err != nil {
		return nil, err
	}

	metrics.Registry.MustRegister(
		&controllers.OrgBillingRefLinkMetric{
			Client: mgr.GetClient(),
		})
	metrics.Registry.MustRegister(
		&controllers.OrgInfoMetric{
			Client: mgr.GetClient(),
		})
	metrics.Registry.MustRegister(
		&controllers.EmailPendingMetric{
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
	dor := &controllers.DefaultOrganizationReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("organization-members-controller"),
	}
	if err = dor.SetupWithManager(mgr); err != nil {
		return nil, err
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

		UsernamePrefix: usernamePrefix,
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

	invmail := controllers.NewInvitationEmailReconciler(
		mgr.GetClient(),
		mgr.GetEventRecorderFor("invitation-email-controller"),
		mgr.GetScheme(),
		mailSender,
		invEmailBaseRetryDelay,
	)
	if err = invmail.SetupWithManager(mgr); err != nil {
		return nil, err
	}

	if saleOrderStorage == "odoo16" {
		storage, err := saleorder.NewOdoo16Storage(&odooCredentials, &saleorder.Odoo16Options{
			SaleOrderClientReferencePrefix: saleOrderClientReference,
			SaleOrderInternalNote:          saleOrderInternalNote,
			Odoo8CompatibilityMode:         saleOrderCompatMode,
		})
		if err != nil {
			return nil, err
		}
		saleorder := &controllers.SaleOrderReconciler{
			Client:           mgr.GetClient(),
			Scheme:           mgr.GetScheme(),
			Recorder:         mgr.GetEventRecorderFor("sale-order-controller"),
			SaleOrderStorage: storage,
		}
		if err = saleorder.SetupWithManager(mgr); err != nil {
			return nil, err
		}
	}

	metrics.Registry.MustRegister(invmail.GetMetrics())

	mgr.GetWebhookServer().Register("/validate-appuio-io-v1-user", &webhook.Admission{
		Handler: &webhooks.UserValidator{},
	})
	mgr.GetWebhookServer().Register("/validate-user-appuio-io-v1-invitation", &webhook.Admission{
		Handler: &webhooks.InvitationValidator{
			UsernamePrefix: usernamePrefix,
		},
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

func setupEmailCron(
	ctx context.Context,
	crontab string,
	mgr ctrl.Manager,
	beMailSender mailsenders.MailSender,
	beMailRecipient string,
) (*cron.Cron, error) {

	bemail := controllers.NewBillingEntityEmailCronJob(
		mgr.GetClient(),
		mgr.GetEventRecorderFor("invitation-email-controller"),
		beMailSender,
		beMailRecipient,
	)

	metrics.Registry.MustRegister(bemail.GetMetrics())
	syncLog := ctrl.Log.WithName("email_cron")

	c := cron.New()
	_, err := c.AddFunc(crontab, func() {
		err := bemail.Run(ctx)

		if err == nil {
			return
		}
		syncLog.Error(err, "Error during periodic job")

	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

func setupBillingEntityRBACCron(
	ctx context.Context,
	crontab string,
	mgr ctrl.Manager,
) (*cron.Cron, error) {

	rbac := &controllers.BillingEntityRBACCronJob{Client: mgr.GetClient()}
	syncLog := ctrl.Log.WithName("be_rbac_cron")
	c := cron.New()
	_, err := c.AddFunc(crontab, func() {
		err := rbac.Run(ctx)
		if err != nil {
			syncLog.Error(err, "Error during periodic job")
		}
	})
	return c, err
}
