package config

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

type Config struct {
	Env        string     `mapstructure:"env"`
	API        API        `mapstructure:"api"`
	Database   Database   `mapstructure:"database"`
	Kafka      Kafka      `mapstructure:"kafka"`
	Test       Test       `mapstructure:"test"`
	Services   Services   `mapstructure:"services"`
	JWT        JWT        `mapstructure:"jwt"`
	Prometheus Prometheus `mapstructure:"prometheus"`
}

type API struct {
	BaseURL      string        `mapstructure:"base_url"`
	Timeout      time.Duration `mapstructure:"timeout"`
	ClientID     string        `mapstructure:"client_id"`
	ClientSecret string        `mapstructure:"client_secret"`
}

type Database struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type Kafka struct {
	Brokers                   string        `mapstructure:"brokers"`
	GroupID                   string        `mapstructure:"group_id"`
	PollInterval              time.Duration `mapstructure:"poll_interval"`
	Topics                    KafkaTopics   `mapstructure:"topics"`
	Acks                      string        `mapstructure:"acks"`
	Retries                   int           `mapstructure:"retries"`
	RetryBackoffMs            int           `mapstructure:"retry_backoff_ms"`
	DeliveryTimeoutMs         int           `mapstructure:"delivery_timeout_ms"`
	QueueBufferingMaxMessages int           `mapstructure:"queue_buffering_max_messages"`
	QueueBufferingMaxMs       int           `mapstructure:"queue_buffering_max_ms"`
	CompressionType           string        `mapstructure:"compression_type"`
	BatchSize                 int           `mapstructure:"batch_size"`
	LingerMs                  int           `mapstructure:"linger_ms"`
}

type KafkaTopics struct {
	UserEvents  string `mapstructure:"user_events"`
	PostEvents  string `mapstructure:"post_events"`
	ErrorEvents string `mapstructure:"error_events"`
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

type Prometheus struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

func MustLoad() *Config {
	viper.SetConfigName("test-config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	viper.SetDefault("env", "test")

	viper.SetDefault("api.base_url", "http://localhost:8080/api")
	viper.SetDefault("api.timeout", "10s")
	viper.SetDefault("api.client_id", "e2e-test-client")
	viper.SetDefault("api.client_secret", "e2e-test-secret")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.db_name", "pinstack_test")
	viper.SetDefault("database.ssl_mode", "disable")

	viper.SetDefault("kafka.brokers", "localhost:9092")
	viper.SetDefault("kafka.group_id", "e2e-test-group")
	viper.SetDefault("kafka.poll_interval", "100ms")
	viper.SetDefault("kafka.topics.user_events", "user-events")
	viper.SetDefault("kafka.topics.post_events", "post-events")
	viper.SetDefault("kafka.topics.error_events", "error-events")
	viper.SetDefault("kafka.acks", "all")
	viper.SetDefault("kafka.retries", 3)
	viper.SetDefault("kafka.retry_backoff_ms", 500)
	viper.SetDefault("kafka.delivery_timeout_ms", 5000)
	viper.SetDefault("kafka.queue_buffering_max_messages", 100000)
	viper.SetDefault("kafka.queue_buffering_max_ms", 5)
	viper.SetDefault("kafka.compression_type", "snappy")
	viper.SetDefault("kafka.batch_size", 16384)
	viper.SetDefault("kafka.linger_ms", 5)

	viper.SetDefault("test.concurrent", 5)
	viper.SetDefault("test.requests_per_test", 100)
	viper.SetDefault("test.test_timeout", "2m")
	viper.SetDefault("test.cleanup", true)
	viper.SetDefault("test.log_level", "info")

	viper.SetDefault("services.user_service.address", "localhost")
	viper.SetDefault("services.user_service.port", 50051)
	viper.SetDefault("services.auth_service.address", "localhost")
	viper.SetDefault("services.auth_service.port", 50052)
	viper.SetDefault("services.post_service.address", "localhost")
	viper.SetDefault("services.post_service.port", 50053)
	viper.SetDefault("services.relation_service.address", "localhost")
	viper.SetDefault("services.relation_service.port", 50054)
	viper.SetDefault("services.notification_service.address", "localhost")
	viper.SetDefault("services.notification_service.port", 50055)

	viper.SetDefault("jwt.secret", "my-secret")
	viper.SetDefault("jwt.access_expires_at", "1m")
	viper.SetDefault("jwt.refresh_expires_at", "5m")

	viper.SetDefault("prometheus.address", "0.0.0.0")
	viper.SetDefault("prometheus.port", 9106)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %s", err)
		os.Exit(1)
	}

	apiTimeout, err := time.ParseDuration(viper.GetString("api.timeout"))
	if err != nil {
		log.Printf("Error reading api.timeout: %s", err)
		apiTimeout = 10 * time.Second
	}

	pollInterval, err := time.ParseDuration(viper.GetString("kafka.poll_interval"))
	if err != nil {
		log.Printf("Error reading kafka.poll_interval: %s", err)
		pollInterval = 100 * time.Millisecond
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
		Database: Database{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetInt("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			DBName:   viper.GetString("database.db_name"),
			SSLMode:  viper.GetString("database.ssl_mode"),
		},
		Kafka: Kafka{
			Brokers:      viper.GetString("kafka.brokers"),
			GroupID:      viper.GetString("kafka.group_id"),
			PollInterval: pollInterval,
			Topics: KafkaTopics{
				UserEvents:  viper.GetString("kafka.topics.user_events"),
				PostEvents:  viper.GetString("kafka.topics.post_events"),
				ErrorEvents: viper.GetString("kafka.topics.error_events"),
			},
			Acks:                      viper.GetString("kafka.acks"),
			Retries:                   viper.GetInt("kafka.retries"),
			RetryBackoffMs:            viper.GetInt("kafka.retry_backoff_ms"),
			DeliveryTimeoutMs:         viper.GetInt("kafka.delivery_timeout_ms"),
			QueueBufferingMaxMessages: viper.GetInt("kafka.queue_buffering_max_messages"),
			QueueBufferingMaxMs:       viper.GetInt("kafka.queue_buffering_max_ms"),
			CompressionType:           viper.GetString("kafka.compression_type"),
			BatchSize:                 viper.GetInt("kafka.batch_size"),
			LingerMs:                  viper.GetInt("kafka.linger_ms"),
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
		Prometheus: Prometheus{
			Address: viper.GetString("prometheus.address"),
			Port:    viper.GetInt("prometheus.port"),
		},
	}

	return config
}
