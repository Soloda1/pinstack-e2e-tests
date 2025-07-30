package gateway_relation

import (
	"flag"
	"github.com/Soloda1/pinstack-system-tests/config"
	"github.com/Soloda1/pinstack-system-tests/internal/client"
	"github.com/Soloda1/pinstack-system-tests/internal/logger"
	"os"
	"testing"
)

var (
	cfg            *config.Config
	log            *logger.Logger
	apiClient      *client.Client
	relationClient *client.RelationClient
)

// TestMain запускается перед любым тестом в пакете и настраивает тестовое окружение
func TestMain(m *testing.M) {
	flag.Parse()

	cfg = config.MustLoad()

	log = logger.New(cfg.Env)
	log.Info("Starting auth gateway tests", "env", cfg.Env)

	apiClient = client.NewClient(cfg, log)
	relationClient = client.NewRelationClient(apiClient)

	log.Info("Setup completed, starting tests")
	code := m.Run()

	log.Info("Tests finished")

	os.Exit(code)
}
