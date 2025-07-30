package gateway_notification

import (
	"flag"
	"github.com/Soloda1/pinstack-system-tests/config"
	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/logger"
	"os"
	"testing"
)

var (
	cfg                *config.Config
	log                *logger.Logger
	apiClient          *client.Client
	notificationClient *client.NotificationClient
)

// TestMain runs before any test in the package and sets up the test environment
func TestMain(m *testing.M) {
	flag.Parse()

	cfg = config.MustLoad("../../../../config")
	log = logger.New(cfg.Env)
	log.Info("Starting notification gateway tests", "env", cfg.Env)

	apiClient = client.NewClient(cfg, log)
	notificationClient = client.NewNotificationClient(apiClient)

	log.Info("Setup completed, starting tests")
	code := m.Run()

	log.Info("Tests finished")

	os.Exit(code)
}
