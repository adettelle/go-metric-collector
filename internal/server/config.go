package server

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

type Config struct {
	Address string
}

func getAddr() string {
	addr := os.Getenv("ADDRESS")
	if addr != "" {
		ensureAddrFLagIsCorrect(addr)
		return addr
	}
	flagAddr := flag.String("a", "localhost:8080", "Net address localhost:port")
	flag.Parse()
	ensureAddrFLagIsCorrect(*flagAddr)
	return *flagAddr
}

func NewConfig() (*Config, error) {
	return &Config{Address: getAddr()}, nil
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
