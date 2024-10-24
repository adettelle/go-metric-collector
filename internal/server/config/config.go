// пакет конфигурации сервера
package config

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/adettelle/go-metric-collector/internal/helpers"
)

const (
	defaultAddress      = "localhost:8080"
	defaulStoreInterval = 300
	defaultStoragePath  = "/tmp/metrics-db.json"
	defaultRestore      = true
	defaultDBParams     = "host=localhost port=5433 user=postgres password=password dbname=metrics-test sslmode=disable"
)

type Config struct {
	Address       string `json:"address"`
	DBParams      string `json:"database_dsn"`
	Key           string `json:"key"`
	Config        string // путь до json файла конфигурации
	StoragePath   string `json:"store_file"`     // по умолчанию /tmp/metrics-db.json
	CryptoKey     string `json:"crypto_key"`     // путь до приватного ключа асимметричного шифрования
	Cert          string `json:"cert"`           // путь до сертификата шифрования
	StoreInterval int    `json:"store_interval"` // по умолчанию 300 сек
	Restore       bool   `json:"restore"`        // по умолчанию true
}

// приоритет:
// сначала проверяем флаги и заполняем структуру конфига оттуда
// потом проверяем переменные окружения и перезаписываем структуру конфига оттуда
// далее проверяем, если есть json файл и дополняем структкуру конфига оттуда
func New() (*Config, error) {
	flagAddr := flag.String("a", "", "Net address localhost:port")                   // "localhost:8080"
	flagStoreInterval := flag.Int("i", 0, "store metrics to file interval, seconds") // 300
	flagStoragePath := flag.String("f", "", "file storage path")
	flagRestore := flag.Bool("r", defaultRestore, "restore or not data from file storage path")
	flagDBParams := flag.String("d", "", "db connection params")
	flagKey := flag.String("k", "", "secret key")
	flagCryptoKey := flag.String("crypto-key", "", "path to file with private key")
	flagCert := flag.String("cert", "", "path to file with certificate")
	flagConfig := flag.String("config", "", "path to file with config parametrs")

	flag.Parse()

	cfg := Config{
		Address:       getAddr(flagAddr),
		StoreInterval: getStoreInterval(flagStoreInterval),
		StoragePath:   getStoragePath(flagStoragePath),
		Restore:       getRestore(flagRestore),
		DBParams:      getDBParams(flagDBParams),
		Key:           getKey(flagKey),
		CryptoKey:     getCryptoKey(flagCryptoKey),
		Cert:          getCert(flagCert),
		Config:        getConfig(flagConfig),
	}

	if cfg.Config != "" {
		log.Println("cfg.Config: ", cfg.Config)
		cfgFromJSON, err := helpers.ReadCfgJSON[Config](cfg.Config)
		if err != nil {
			return nil, err
		}

		if cfg.Address == "" {
			cfg.Address = cfgFromJSON.Address
		}
		if cfg.StoreInterval == 0 {
			cfg.StoreInterval = cfgFromJSON.StoreInterval
		}
		if cfg.StoragePath == "" {
			cfg.StoragePath = cfgFromJSON.StoragePath
		}
		if !cfg.Restore {
			cfg.Restore = cfgFromJSON.Restore
		}
		if cfg.DBParams == "" {
			cfg.DBParams = cfgFromJSON.DBParams
		}
		if cfg.Key == "" {
			cfg.Key = cfgFromJSON.Key
		}
		if cfg.CryptoKey == "" {
			cfg.CryptoKey = cfgFromJSON.CryptoKey
		}
	}

	if cfg.Address == "" {
		cfg.Address = defaultAddress
	}
	ensureAddrFLagIsCorrect(cfg.Address)
	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = defaulStoreInterval
	}
	if cfg.StoragePath == "" {
		cfg.StoragePath = defaultStoragePath
	}
	ensureFileExists(cfg.StoragePath)
	if cfg.DBParams == "" {
		cfg.DBParams = defaultDBParams
	}

	log.Printf("config: %+v\n", cfg)
	return &cfg, nil
}

func getConfig(flagConfig *string) string {
	config := os.Getenv("CONFIG")
	if config != "" {
		return config
	}
	return *flagConfig
}

func getKey(flagKey *string) string {
	key := os.Getenv("KEY")
	if key != "" {
		return key
	}
	return *flagKey
}

func getCryptoKey(flagCryptoKey *string) string {
	cryptoKey, ok := os.LookupEnv("CRYPTO_KEY")
	if ok {
		return cryptoKey
	}
	// cryptoKey := os.Getenv("CRYPTO_KEY")
	// if cryptoKey != "" {
	// 	return cryptoKey
	// }
	return *flagCryptoKey
}

func getCert(flagCert *string) string {
	cert, ok := os.LookupEnv("CERT")
	if ok {
		return cert
	}

	return *flagCert
}

func getAddr(flagAddr *string) string {
	log.Println("flagAddr:", *flagAddr, os.Getenv("ADDRESS"))
	addr := os.Getenv("ADDRESS")
	if addr != "" {
		// ensureAddrFLagIsCorrect(addr)
		return addr
	}
	// ensureAddrFLagIsCorrect(*flagAddr)
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

func ensureFileExists(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) { // storagePath
		f, err := os.Create(path) // storagePath
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}
}

func getStoragePath(flagStoragePath *string) string {
	log.Println("flagStoragePath:", *flagStoragePath)
	storagePath, ok := os.LookupEnv("FILE_STORAGE_PATH")
	log.Println("StoragePathEnv:", storagePath)
	if !ok {
		storagePath = *flagStoragePath
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
