// пакет конфигурации сервера
package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

type Config struct {
	Address       string
	StoreInterval int    // по умолчанию 300 сек
	StoragePath   string // по умолчанию /tmp/metrics-db.json
	Restore       bool   // по умолчанию true
}

func New() (*Config, error) {
	flagAddr := flag.String("a", "localhost:8080", "Net address localhost:port")
	flagStoreInterval := flag.Int("i", 300, "store metrics to file interval, seconds")
	flagStoragePath := flag.String("f", "/tmp/metrics-db.json", "file storage path")
	flagRestore := flag.Bool("r", true, "restore or not data from file storage path")
	flag.Parse()

	return &Config{
		Address:       getAddr(flagAddr),
		StoreInterval: getStoreInterval(flagStoreInterval),
		StoragePath:   getStoragePath(flagStoragePath),
		Restore:       getRestore(flagRestore),
	}, nil
}

func getAddr(flagAddr *string) string {
	addr := os.Getenv("ADDRESS")
	if addr != "" {
		ensureAddrFLagIsCorrect(addr)
		return addr
	}
	ensureAddrFLagIsCorrect(*flagAddr)
	return *flagAddr
}

func getStoreInterval(flagStoreInterval *int) int {
	envStoreInterval := os.Getenv("REPORT_INTERVAL")

	var storeInterval int

	if envStoreInterval == "" {
		storeInterval = *flagStoreInterval
	} else {
		storeInterval = parseIntOrPanic(envStoreInterval)
	}

	return storeInterval
}

func getStoragePath(flagStoragePath *string) string { // надо ли здесь ставить условия на нулевое значение????
	StoragePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		return StoragePath
	}

	return *flagStoragePath
}

func getRestore(flagRestore *bool) bool {
	envRestore := os.Getenv("RESTORE")
	if envRestore == "true" {
		return true
	} else if envRestore == "false" {
		return false
	}

	return *flagRestore
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

func parseIntOrPanic(s string) int {
	x, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return x
}
