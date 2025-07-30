package gateway_posts

import (
	"flag"
	"github.com/Soloda1/pinstack-system-tests/config"
	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/logger"
	"os"
	"testing"
)

var (
	cfg         *config.Config
	log         *logger.Logger
	apiClient   *client.Client
	postsClient *client.PostClient
)

// TestMain запускается перед любым тестом в пакете и настраивает тестовое окружение
func TestMain(m *testing.M) {
	flag.Parse()

	cfg = config.MustLoad("../../../../config")
	log = logger.New(cfg.Env)
	log.Info("Starting auth gateway tests", "env", cfg.Env)

	apiClient = client.NewClient(cfg, log)
	postsClient = client.NewPostClient(apiClient)

	log.Info("Setup completed, starting tests")
	code := m.Run()

	log.Info("Tests finished")

	os.Exit(code)
}
