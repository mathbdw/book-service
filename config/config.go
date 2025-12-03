package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/caarlos0/env/v11"
)

// Build information -ldflags .
const (
	version    string = "dev"
	commitHash string = "-"
)

// Database - contains all parameters database connection.
type Database struct {
	Host            string        `yaml:"host" env:"PG_HOST,required"`
	Port            uint16        `yaml:"port" env:"PG_PORT,required"`
	User            string        `yaml:"user" env:"PG_USER,required"`
	Password        string        `yaml:"password" env:"PG_PASSWORD,required"`
	Migrations      string        `yaml:"migrations"`
	Name            string        `yaml:"name"`
	SslMode         string        `yaml:"sslmode"`
	Driver          string        `yaml:"driver"`
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
}

// Graylog - contains parameter address gelf
type Graylog struct {
	Host    string `yaml:"host"`
	Port    uint16 `yaml:"port"`
	Version string `yaml:"version"`
	Status  bool   `yaml:"status"`
}

// Grpc - contains parameter address grpc.
type Grpc struct {
	MaxConnectionIdle int64  `yaml:"maxConnectionIdle"`
	Timeout           int64  `yaml:"timeout"`
	MaxConnectionAge  int64  `yaml:"maxConnectionAge"`
	Host              string `yaml:"host"`
	Port              uint16 `yaml:"port"`
}

// Rest - contains parameter rest json connection.
type Rest struct {
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
}

// Project - contains all parameters project information.
type Project struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
	Version     string `yaml:"version"`
	CommitHash  string
	Debug       bool `yaml:"debug"`
}

// Metrics - contains all parameters metrics information.
type Metric struct {
	Host             string        `yaml:"host"`
	Port             uint16        `yaml:"port"`
	Timeout          time.Duration `yaml:"timeout"`
	Insecure         bool          `yaml:"insecure"`
	ExporterInterval time.Duration `yaml:"exporterInterval"`
	ExporterTimeout  time.Duration `yaml:"exporterTimeout"`
}

// Opentelemetry - contains all parameters metrics information.
type Tracer struct {
	Host     string        `yaml:"host"`
	Port     uint16        `yaml:"port"`
	Timeout  time.Duration `yaml:"timeout"`
	Insecure bool          `yaml:"insecure"`
}

// Profucer - producer kafka
type Producer struct {
	ReturnSuccesses bool   `yaml:"returnSuccesses"`
	RequiredAcks    int16  `yaml:"requiredAcks"`
	Compression     int8   `yaml:"compression"`
	Partitioner     string `yaml:"partitioner"`
}

// Consumer - consumer kafka
type Consumer struct {
	GroupId           string `yaml:"groupId"`
	RebalanceStrategy string `yaml:"rebalancrStrategy"`
}

// Publisher - publisher option
type Publisher struct {
	Interval     time.Duration `yaml:"interval"`
	BatchSize    uint64        `yaml:"batchSize"`
	CountWorkers uint8         `yaml:"countWorkers"`
}

// Topics - topics for kafka
type Topics struct {
	Publish string `yaml:"publish"`
}

// Kafka - contains all parameters kafka information.
type Kafka struct {
	Topics    Topics    `yaml:"topics"`
	Brokers   []string  `yaml:"brokers"`
	Producer  Producer  `yaml:"producer"`
	Consumer  Consumer  `yaml:"consumer"`
	Publisher Publisher `yaml:"publisher"`
}

// Status config for service.
type Status struct {
	Host          string `yaml:"host"`
	VersionPath   string `yaml:"versionPath"`
	LivenessPath  string `yaml:"livenessPath"`
	ReadinessPath string `yaml:"readinessPath"`
	Port          uint16 `yaml:"port"`
}

type Bot struct {
	Token       string `yaml:"token" env:"TBOT_TOKEN,required"`
	ReadTimeout int    `yaml:"readTimeout"`
}

// Config - contains all configuration parameters in config package.
type Config struct {
	Project  Project  `yaml:"project"`
	Graylog  Graylog  `yaml:"graylog"`
	Grpc     Grpc     `yaml:"grpc"`
	Rest     Rest     `yaml:"rest"`
	Database Database `yaml:"database"`
	Metric   Metric   `yaml:"metric"`
	Tracer   Tracer   `yaml:"tracer"`
	Kafka    Kafka    `yaml:"kafka"`
	Status   Status   `yaml:"status"`
	Bot      Bot      `yaml:"telegram"`
}

// ReadConfigYML - read configurations from file and init instance Config.
func ReadConfigYML(filePath string) (*Config, error) {
	var cfg *Config
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	cfg.Project.Version = version
	cfg.Project.CommitHash = commitHash

	return cfg, nil
}