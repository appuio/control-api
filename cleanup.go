package main

import (
	"os"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/appuio/control-api/apiserver/billing/odoostorage/odoo/odoo8"
)

// APICommand creates a new command allowing to start the API server
func CleanupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "cleanup",
	}

	odooUrl := cmd.Flags().String("billing-entity-odoo8-url", "http://localhost:8069", "URL of the Odoo instance to use for billing entities")
	debugTransport := cmd.Flags().Bool("billing-entity-odoo8-debug-transport", false, "Enable debug logging for the Odoo transport")
	minAge := cmd.Flags().Duration("billing-entity-odoo8-cleanup-after", time.Hour, "Clean up only records older than this")

	cmd.Run = func(cmd *cobra.Command, args []string) {
		ctx := ctrl.SetupSignalHandler()
		l := klog.FromContext(ctx)
		scrubber := odoo8.NewFailedRecordScrubber(
			*odooUrl,
			*debugTransport,
		)

		err := scrubber.CleanupIncompleteRecords(ctx, *minAge)
		if err != nil {
			l.Error(err, "Unable to clean up incomplete records")
			os.Exit(1)
		}
		l.Info("Cleanup complete!")
	}

	return cmd
}
