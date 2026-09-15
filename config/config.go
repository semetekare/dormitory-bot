package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server        ServerConfig   `mapstructure:"server"`
	Database      DatabaseConfig `mapstructure:"database"`
	Redis         RedisConfig    `mapstructure:"redis"`
	Logging       LoggingConfig  `mapstructure:"logging"`
	MAXBotToken   string         `mapstructure:"max_bot_token"`
	EncryptionKey string         `mapstructure:"encryption_key"`
	Timezone      string         `mapstructure:"timezone"`
	LogsDir       string         `mapstructure:"logs_dir"`
	RabbitMQ      RabbitMQConfig `mapstructure:"rabbitmq"`
	MockMode      bool           `mapstructure:"mock_mode"`
	JWTSecret     string         `mapstructure:"jwt_secret"`
	MiniAppURL    string         `mapstructure:"mini_app_url"`
	BotName       string         `mapstructure:"bot_name"`
}

type ServerConfig struct {
	Port         string        `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RabbitMQConfig struct {
	URL          string `mapstructure:"url"`
	ExchangeName string `mapstructure:"exchange_name"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

func Load() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "dormitory_user")
	viper.SetDefault("database.password", "dormitory_password")
	viper.SetDefault("database.dbname", "dormitory_db")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", "5m")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	viper.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	viper.SetDefault("rabbitmq.exchange_name", "dormitory.eis")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Server
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")

	// Database
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "DATABASE_SSLMODE")
	viper.BindEnv("database.max_open_conns", "DATABASE_MAX_OPEN_CONNS")
	viper.BindEnv("database.max_idle_conns", "DATABASE_MAX_IDLE_CONNS")
	viper.BindEnv("database.conn_max_lifetime", "DATABASE_CONN_MAX_LIFETIME")

	// Redis
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")

	// RabbitMQ
	viper.BindEnv("rabbitmq.url", "RABBITMQ_URL")
	viper.BindEnv("rabbitmq.exchange_name", "RABBITMQ_EXCHANGE_NAME")

	// Bot secrets
	viper.BindEnv("max_bot_token", "MAX_BOT_TOKEN")
	viper.BindEnv("encryption_key", "ENCRYPTION_KEY")
	viper.BindEnv("timezone", "TIMEZONE")
	viper.SetDefault("timezone", "Europe/Moscow")
	viper.SetDefault("logs_dir", "logs")

	// Mock mode
	viper.BindEnv("mock_mode", "MOCK_MODE")
	viper.SetDefault("mock_mode", false)

	// JWT
	viper.BindEnv("jwt_secret", "JWT_SECRET")
	viper.SetDefault("jwt_secret", "")

	// Mini App
	viper.BindEnv("mini_app_url", "MINI_APP_URL")
	viper.SetDefault("mini_app_url", "http://localhost:5173")

	// Bot name (для deep link: https://max.ru/<botName>?startapp=...)
	viper.BindEnv("bot_name", "BOT_NAME")
	viper.SetDefault("bot_name", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("unable to decode into struct: %w", err))
	}

	return &cfg
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

func (c *Config) LogDir() string {
	if c.LogsDir != "" {
		return c.LogsDir
	}
	return "logs"
}

// JWTSecretOrEncryptionKey возвращает JWT-секрет.
// Если JWT_SECRET не задан, падает на EncryptionKey (обратная совместимость).
func (c *Config) JWTSecretOrEncryptionKey() string {
	if c.JWTSecret != "" {
		return c.JWTSecret
	}
	return c.EncryptionKey
}
