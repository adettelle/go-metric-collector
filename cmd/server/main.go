// Server for collecting rantime metrics, collects reports from agents by HTTP protocol.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/adettelle/go-metric-collector/internal/api"
	database "github.com/adettelle/go-metric-collector/internal/db"
	"github.com/adettelle/go-metric-collector/internal/grpcserver"
	"github.com/adettelle/go-metric-collector/internal/migrator"

	"github.com/adettelle/go-metric-collector/internal/server/config"
	"github.com/adettelle/go-metric-collector/internal/storage/dbstorage"
	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	var err error

	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)

	cfg, err := config.New(false)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("config:", cfg)

	storager, err := initStorager(cfg)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	mAPI := api.NewMetricHandlers(storager, cfg, &wg)
	router := api.NewMetricRouter(storager, mAPI)
	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}
	go func() {
		fmt.Printf("Starting server on %s\n", cfg.Address)

		err := srv.ListenAndServeTLS(cfg.Cert, cfg.CryptoKey)
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		err := grpcserver.StartServer(storager, cfg.GrpcPort)
		if err != nil {
			log.Fatal(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	done := make(chan bool, 1)

	go func() {
		s := <-c
		log.Printf("Got termination signal: %s. Graceful shutdown", s)

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatal(err) // failure/timeout shutting down the server gracefully
		}

		mAPI.Finalizing = true

		err = storager.Finalize()
		if err != nil {
			log.Println(err)
			log.Println("unable to write to file")
		}
		mAPI.Wg.Wait() //  ждем завершение update'ов
		done <- true
	}()
	<-done
}

// initStorager not only constructs, but also starts related processes
// depending on which storager we choose.
func initStorager(cfg *config.Config) (api.Storager, error) {
	// log.Println("config in initStorager:", cfg)
	var storager api.Storager

	if cfg.DBParams != "" {

		db, err := database.NewDBConnection(cfg.DBParams).Connect()
		if err != nil {
			log.Fatal(err)
		}
		storager = &dbstorage.DBStorage{
			Ctx: context.Background(),
			DB:  db,
		}

		migrator.MustApplyMigrations(cfg.DBParams)

	} else {
		var ms *memstorage.MemStorage
		log.Println("cfg.StoragePath in initStorager:", cfg.StoragePath)
		ms, err := memstorage.New(cfg.ShouldRestore(), cfg.StoragePath)
		log.Println("ms.StoragePath in initStorager:", ms.FileName)
		if err != nil {
			return nil, err
		}

		if cfg.StoreInterval > 0 {
			go memstorage.StartSaveLoop(time.Second*time.Duration(cfg.StoreInterval),
				cfg.StoragePath, ms)
		} else if cfg.StoreInterval == 0 {
			// если config.StoreInterval равен 0, то мы назначаем MemStorage FileName, чтобы
			// он мог синхронно писать изменения
			ms.FileName = cfg.StoragePath
		}

		storager = ms
	}
	return storager, nil
}
