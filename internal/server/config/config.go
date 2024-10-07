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
	DBParams      string
	Key           string
}

func New() (*Config, error) {
	flagAddr := flag.String("a", "localhost:8080", "Net address localhost:port")
	flagStoreInterval := flag.Int("i", 300, "store metrics to file interval, seconds")
	flagStoragePath := flag.String("f", "/tmp/metrics-db.json", "file storage path")
	flagRestore := flag.Bool("r", true, "restore or not data from file storage path")
	flagDBParams := flag.String(
		"d",
		"host=localhost port=5433 user=postgres password=password dbname=metrics-test sslmode=disable",
		"db connection params")
	flagKey := flag.String("k", "", "secret key")

	flag.Parse()

	return &Config{
		Address:       getAddr(flagAddr),
		StoreInterval: getStoreInterval(flagStoreInterval),
		StoragePath:   getStoragePath(flagStoragePath),
		Restore:       getRestore(flagRestore),
		DBParams:      getDBParams(flagDBParams),
		Key:           getKey(flagKey),
	}, nil
}

func getKey(flagKey *string) string {
	key := os.Getenv("KEY")
	if key != "" {
		return key
	}
	return *flagKey
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

func getStoragePath(flagStoragePath *string) string {
	log.Println("flagStoragePath:", *flagStoragePath)
	storagePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	log.Println("StoragePathEnv:", storagePath)
	if !ok {
		storagePath = *flagStoragePath
	}

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		f, err := os.Create(storagePath)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}

	return storagePath
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

func (config *Config) ShouldRestore() bool {

	if !config.Restore {
		return false
	}

	fileStoragePath, err := os.Stat(config.StoragePath)

	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Fatal(err)
	}

	// в этом месте мы знаем, что файл существует, и что Restore = true,
	// значит надо убедится в размере файла
	return fileStoragePath.Size() > 0
}

func getDBParams(flagDBParams *string) string {
	envDBParams := os.Getenv("DATABASE_DSN")

	if envDBParams != "" {
		return envDBParams
	}
	return *flagDBParams
}
