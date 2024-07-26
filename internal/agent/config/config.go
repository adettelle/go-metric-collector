package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

const defaultMaxRequestRetries = 3

type Config struct {
	Address           string
	ReportInterval    int // по умолчанию 10 сек
	PollInterval      int // по умолчанию 2 сек
	MaxRequestRetries int // максимальное количесвто попыток запроса
	Key               string
}

func New() (*Config, error) {
	addr := os.Getenv("ADDRESS")
	envPollInterval := os.Getenv("POLL_INTERVAL")
	envReportInterval := os.Getenv("REPORT_INTERVAL")
	envKey := os.Getenv("KEY")

	flagAddr := flag.String("a", "localhost:8080", "Net address localhost:port")
	flagPollInterval := flag.Int("p", 2, "metrics poll interval, seconds")
	flagReportInterval := flag.Int("r", 10, "metrics report interval, seconds")
	flagKey := flag.String("k", "", "secret key")
	flag.Parse()

	if addr == "" {
		addr = *flagAddr
	}
	ensureAddrFLagIsCorrect(addr)

	var pollDelay int
	if envPollInterval == "" {
		pollDelay = *flagPollInterval
	} else {
		pollDelay = parseIntOrPanic(envPollInterval)
	}

	var reportInterval int
	if envReportInterval == "" {
		reportInterval = *flagReportInterval
	} else {
		reportInterval = parseIntOrPanic(envPollInterval)
	}

	var key string
	if envKey == "" {
		key = *flagKey
	} else {
		key = envKey
	}
	return &Config{
		Address:           addr,
		ReportInterval:    reportInterval,
		PollInterval:      pollDelay,
		MaxRequestRetries: defaultMaxRequestRetries,
		Key:               key,
	}, nil
}

func parseIntOrPanic(s string) int {
	x, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return x
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
