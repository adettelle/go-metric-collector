package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/kelseyhightower/envconfig"
)

// const defaultMaxRequestRetries = 3

type Config struct {
	Address           string `env:"ADDRESS" flag:"a" default:"localhost:8080"`
	PollInterval      int    `env:"POLL_INTERVAL" flag:"p" default:"2"`    // по умолчанию 2 сек
	ReportInterval    int    `env:"REPORT_INTERVAL" flag:"r" default:"10"` // по умолчанию 10 сек
	Key               string `env:"KEY" flag:"k"`                          // ключ для подписи
	MaxRequestRetries int    `default:"3"`                                 // максимальное количество попыток запроса
	// количество одновременно исходящих запросов на сервер
	// (количество задач, которое одновременно происходит в worker pool)
	RateLimit int `env:"RATE_LIMIT" flag:"l" default:"1"`
}

func New() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	flag.StringVar(&cfg.Address, "a", cfg.Address, "Net address localhost:port")
	flag.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "metrics poll interval, seconds")
	flag.StringVar(&cfg.Key, "k", cfg.Key, "secret key")
	flag.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "metrics report interval, seconds")
	flag.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "number of simultaneous tasks")

	flag.Parse()

	ensureAddrFLagIsCorrect(cfg.Address) // addr

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
