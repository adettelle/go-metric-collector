package agent

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

type Config struct {
	Address        string
	ReportInterval int // секретный ключ шифрования
	PollInterval   int
}

func NewConfig() (*Config, error) { // apiPort string
	addr := os.Getenv("ADDRESS")
	envPollDelay := os.Getenv("POLL_INTERVAL")
	envReportDelay := os.Getenv("REPORT_INTERVAL")

	flagAddr := flag.String("a", "localhost:8080", "Net address localhost:port")
	flagPollDelay := flag.Int("p", 2, "metrics poll interval, seconds")
	flagReportDelay := flag.Int("r", 10, "metrics report interval, seconds")
	flag.Parse()

	if addr == "" {
		addr = *flagAddr
	}
	ensureAddrFLagIsCorrect(addr)

	var pollDelay int
	if envPollDelay == "" {
		pollDelay = *flagPollDelay
	} else {
		pollDelay = parseIntOrPanic(envPollDelay)
	}

	var reportDelay int
	if envReportDelay == "" {
		reportDelay = *flagReportDelay
	} else {
		reportDelay = parseIntOrPanic(envPollDelay)
	}

	return &Config{Address: addr, ReportInterval: reportDelay, PollInterval: pollDelay}, nil
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
