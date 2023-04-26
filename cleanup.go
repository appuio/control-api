package main

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8"
	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8/countries"
)

// APICommand creates a new command allowing to start the API server
func CleanupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "cleanup",
	}

	odooUrl := cmd.Flags().String("billing-entity-odoo8-url", "http://localhost:8069", "URL of the Odoo instance to use for billing entities")
	debugTransport := cmd.Flags().Bool("billing-entity-odoo8-debug-transport", false, "Enable debug logging for the Odoo transport")
	countryList := cmd.Flags().String("billing-entity-odoo8-country-list", "countries.yaml", "Path to the country list file in the format of [{name: \"Germany\", code: \"DE\", id: 81},...]")
	accountingContactDisplayName := cmd.Flags().String("billing-entity-odoo8-accounting-contact-display-name", "Accounting", "Display name of the accounting contact")
	languagePreference := cmd.Flags().String("billing-entity-odoo8-language-preference", "en_US", "Language preference of the Odoo record")
	paymentTerm := cmd.Flags().Int("billing-entity-odoo8-payment-term-id", 2, "Payment term ID of the Odoo record")
	minAge := cmd.Flags().Duration("billing-entity-odoo8-cleanup-after", time.Hour, "Clean up only records older than this")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		ctx := ctrl.SetupSignalHandler()
		l := klog.FromContext(ctx)
		countryIDs, err := countries.LoadCountryIDs(*countryList)
		if err != nil {
			l.Error(err, "Unable to load country list")
			os.Exit(1)
		}
		storage := odoo8.NewOdoo8Storage(
			*odooUrl,
			*debugTransport,
			odoo8.Config{
				AccountingContactDisplayName: *accountingContactDisplayName,
				LanguagePreference:           *languagePreference,
				PaymentTermID:                *paymentTerm,
				CountryIDs:                   countryIDs,
			},
		)

		err = storage.CleanupIncompleteRecords(ctx, *minAge)
		if err != nil {
			l.Error(err, "Unable to clean up incomplete records")
			os.Exit(1)
		}
		l.Info("Cleanup complete!")
	}

	return cmd
}
