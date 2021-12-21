package main

import (
	"log"
	"os"
	"runtime"

	orgv1 "github.com/appuio/control-api/apis/organization/v1"
	orgStore "github.com/appuio/control-api/apiserver/organization"

	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// Strings are populated by Goreleaser
var (
	version = "snapshot"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	logger := newLogger("control-api", true)

	logger.WithValues(
		"version", version,
		"date", date,
		"commit", commit,
		"go_os", runtime.GOOS,
		"go_arch", runtime.GOARCH,
		"go_version", runtime.Version(),
		"uid", os.Getuid(),
		"gid", os.Getgid(),
	).Info("Starting control-api…")
	err := builder.APIServer.
		WithResourceAndHandler(&orgv1.Organization{}, orgStore.New()).
		WithLocalDebugExtension().
		WithoutEtcd().
		ExposeLoopbackAuthorizer().
		ExposeLoopbackClientConfig().
		ExposeLoopbackMasterClientConfig().
		Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func newLogger(name string, debug bool) logr.Logger {
	level := zapcore.InfoLevel
	if debug {
		level = zapcore.DebugLevel
	}
	logger := zap.New(zap.UseDevMode(true), zap.Level(level))
	return logger.WithName(name)
}
