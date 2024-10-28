package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/adettelle/go-metric-collector/internal/helpers"
	"github.com/kelseyhightower/envconfig"
)

const (
	defaultAddress           = "localhost:8080"
	defaultPollInterval      = 2
	defaultMaxRequestRetries = 3
	defaultRateLimit         = 1
	defaultReportInterval    = 10
)

type Config struct {
	Address           string `envconfig:"ADDRESS" flag:"a" json:"address"`                   // default:"localhost:8080"
	Key               string `envconfig:"KEY" flag:"k" json:"key"`                           // ключ для подписи
	CryptoKey         string `envconfig:"CRYPTO_KEY" flag:"crypto-key" json:"crypto_key"`    // путь до публичного ключа асимметричного шифрования
	ClientCert        string `envconfig:"CLIENT_CERT" flag:"client-cert" json:"client_cert"` // путь до сертификата клиента
	ServerCert        string `envconfig:"SERVER_CERT" flag:"server-cert" json:"server_cert"` // путь до сертификата сервера
	Config            string `envconfig:"CONFIG" flag:"config"`                              // путь до json файла конфигурации
	GrpcUrl           string `envconfig:"GRPC" flag:"grpc"`                                  // адрес и порт grpc сервера (если указан, отправляем метрики по grpc, в противном случае - по http)
	MaxRequestRetries int    // максимальное количество попыток запроса
	PollInterval      int    `envconfig:"POLL_INTERVAL" flag:"p" json:"poll_interval"`     // по умолчанию 2 сек
	ReportInterval    int    `envconfig:"REPORT_INTERVAL" flag:"r" json:"report_interval"` // по умолчанию 10 сек
	// количество одновременно исходящих запросов на сервер
	// (количество задач, которое одновременно происходит в worker pool)
	RateLimit int `env:"RATE_LIMIT" flag:"l" json:"rate_limit"`
}

// приоритет:
// сначала проверяем флаги и заполняем структуру конфига оттуда
// потом проверяем переменные окружения и перезаписываем структуру конфига оттуда
// далее проверяем, если есть json файл и дополняем структкуру конфига оттуда
func New() (*Config, error) {
	var cfg Config

	flag.StringVar(&cfg.Address, "a", cfg.Address, "Net address localhost:port")
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "metrics poll interval, seconds")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "secret key")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "metrics report interval, seconds")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "number of simultaneous tasks")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "path to file with public key")
	flag.StringVar(&cfg.ClientCert, "client-cert", cfg.ClientCert, "path to client sertificate")
	flag.StringVar(&cfg.ServerCert, "server-cert", cfg.ServerCert, "path to server sertificate")
	flag.StringVar(&cfg.GrpcUrl, "grpc", cfg.GrpcUrl, "grpc server url")

	flag.Parse()

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	if cfg.Config != "" {
		cfgFromJSON, err := helpers.ReadCfgJSON[Config](cfg.Config)
		if err != nil {
			return nil, err
		}

		if cfg.Address == "" {
			cfg.Address = cfgFromJSON.Address
		}
		if cfg.Key == "" {
			cfg.Key = cfgFromJSON.Key
		}
		if cfg.CryptoKey == "" {
			cfg.CryptoKey = cfgFromJSON.CryptoKey
		}
		if cfg.ClientCert == "" {
			cfg.ClientCert = cfgFromJSON.ClientCert
		}
		if cfg.ServerCert == "" {
			cfg.ServerCert = cfgFromJSON.ServerCert
		}
		if cfg.MaxRequestRetries == 0 {
			cfg.MaxRequestRetries = cfgFromJSON.MaxRequestRetries
		}
		if cfg.PollInterval == 0 {
			cfg.PollInterval = cfgFromJSON.PollInterval
		}
		if cfg.ReportInterval == 0 {
			cfg.ReportInterval = cfgFromJSON.ReportInterval
		}
		if cfg.RateLimit == 0 {
			cfg.RateLimit = cfgFromJSON.RateLimit
		}
	}

	if cfg.Address == "" {
		cfg.Address = defaultAddress
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = defaultPollInterval
	}
	if cfg.MaxRequestRetries == 0 {
		cfg.MaxRequestRetries = defaultMaxRequestRetries
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = defaultRateLimit
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = defaultReportInterval
	}

	ensureAddrFLagIsCorrect(cfg.Address)

	return &cfg, nil
}

func ensureAddrFLagIsCorrect(addr string) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		log.Fatal(err)
	}

	_, err = strconv.Atoi(port)
	if err != nil {
		log.Fatal(fmt.Errorf("invalid port: '%s'", port))
	}
}
