package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

type Config struct {
	Env      string   `mapstructure:"env"`
	API      API      `mapstructure:"api"`
	Test     Test     `mapstructure:"test"`
	Services Services `mapstructure:"services"`
	JWT      JWT      `mapstructure:"jwt"`
}

type API struct {
	BaseURL      string        `mapstructure:"base_url"`
	Timeout      time.Duration `mapstructure:"timeout"`
	ClientID     string        `mapstructure:"client_id"`
	ClientSecret string        `mapstructure:"client_secret"`
}

type Test struct {
	Concurrent      int           `mapstructure:"concurrent"`
	RequestsPerTest int           `mapstructure:"requests_per_test"`
	TestTimeout     time.Duration `mapstructure:"test_timeout"`
	Cleanup         bool          `mapstructure:"cleanup"`
	LogLevel        string        `mapstructure:"log_level"`
}

type Services struct {
	UserService         ServiceConfig `mapstructure:"user_service"`
	AuthService         ServiceConfig `mapstructure:"auth_service"`
	PostService         ServiceConfig `mapstructure:"post_service"`
	RelationService     ServiceConfig `mapstructure:"relation_service"`
	NotificationService ServiceConfig `mapstructure:"notification_service"`
}

type ServiceConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

type JWT struct {
	Secret           string        `mapstructure:"secret"`
	AccessExpiresAt  time.Duration `mapstructure:"access_expires_at"`
	RefreshExpiresAt time.Duration `mapstructure:"refresh_expires_at"`
}

func MustLoad(configPath string) *Config {
	viper.SetConfigName("test-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	viper.SetDefault("env", "test")

	viper.SetDefault("api.base_url", "http://localhost:42080/api")
	viper.SetDefault("api.timeout", "10s")
	viper.SetDefault("api.client_id", "e2e-test-client")
	viper.SetDefault("api.client_secret", "e2e-test-secret")

	viper.SetDefault("test.concurrent", 5)
	viper.SetDefault("test.requests_per_test", 100)
	viper.SetDefault("test.test_timeout", "2m")
	viper.SetDefault("test.cleanup", true)
	viper.SetDefault("test.log_level", "info")

	viper.SetDefault("services.user_service.address", "localhost")
	viper.SetDefault("services.user_service.port", 42051)
	viper.SetDefault("services.auth_service.address", "localhost")
	viper.SetDefault("services.auth_service.port", 42052)
	viper.SetDefault("services.post_service.address", "localhost")
	viper.SetDefault("services.post_service.port", 42053)
	viper.SetDefault("services.relation_service.address", "localhost")
	viper.SetDefault("services.relation_service.port", 42054)
	viper.SetDefault("services.notification_service.address", "localhost")
	viper.SetDefault("services.notification_service.port", 42055)

	viper.SetDefault("jwt.secret", "my-secret")
	viper.SetDefault("jwt.access_expires_at", "1m")
	viper.SetDefault("jwt.refresh_expires_at", "30m")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %s", err)
		os.Exit(1)
	}

	apiTimeout, err := time.ParseDuration(viper.GetString("api.timeout"))
	if err != nil {
		log.Printf("Error reading api.timeout: %s", err)
		apiTimeout = 10 * time.Second
	}

	testTimeout, err := time.ParseDuration(viper.GetString("test.test_timeout"))
	if err != nil {
		log.Printf("Error reading  test.test_timeout: %s", err)
		testTimeout = 2 * time.Minute
	}

	accessExpiresAt, err := time.ParseDuration(viper.GetString("jwt.access_expires_at"))
	if err != nil {
		log.Printf("Error reading jwt.access_expires_at: %s", err)
		accessExpiresAt = 1 * time.Minute
	}

	refreshExpiresAt, err := time.ParseDuration(viper.GetString("jwt.refresh_expires_at"))
	if err != nil {
		log.Printf("Error reading  jwt.refresh_expires_at: %s", err)
		refreshExpiresAt = 5 * time.Minute
	}

	config := &Config{
		Env: viper.GetString("env"),
		API: API{
			BaseURL:      viper.GetString("api.base_url"),
			Timeout:      apiTimeout,
			ClientID:     viper.GetString("api.client_id"),
			ClientSecret: viper.GetString("api.client_secret"),
		},
		Test: Test{
			Concurrent:      viper.GetInt("test.concurrent"),
			RequestsPerTest: viper.GetInt("test.requests_per_test"),
			TestTimeout:     testTimeout,
			Cleanup:         viper.GetBool("test.cleanup"),
			LogLevel:        viper.GetString("test.log_level"),
		},
		Services: Services{
			UserService: ServiceConfig{
				Address: viper.GetString("services.user_service.address"),
				Port:    viper.GetInt("services.user_service.port"),
			},
			AuthService: ServiceConfig{
				Address: viper.GetString("services.auth_service.address"),
				Port:    viper.GetInt("services.auth_service.port"),
			},
			PostService: ServiceConfig{
				Address: viper.GetString("services.post_service.address"),
				Port:    viper.GetInt("services.post_service.port"),
			},
			RelationService: ServiceConfig{
				Address: viper.GetString("services.relation_service.address"),
				Port:    viper.GetInt("services.relation_service.port"),
			},
			NotificationService: ServiceConfig{
				Address: viper.GetString("services.notification_service.address"),
				Port:    viper.GetInt("services.notification_service.port"),
			},
		},
		JWT: JWT{
			Secret:           viper.GetString("jwt.secret"),
			AccessExpiresAt:  accessExpiresAt,
			RefreshExpiresAt: refreshExpiresAt,
		},
	}

	return config
}
