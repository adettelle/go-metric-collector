package memstorage

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"go.uber.org/zap"
)

// запись в файл слепка метрик
func WriteMetricsSnapshot(fileName string, ms *MemStorage) error {
	// открываем файл для записи в конец
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	allMetrics := MemStorageToAllMetrics(ms)

	data, err := json.Marshal(allMetrics)
	if err != nil {
		return err
	}

	// добавляем перенос строки
	data = append(data, '\n')

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	logger.Info("writing to file", zap.String("fileName", fileName))

	log.Printf("writing to file: %s", fileName)
	_, err = file.Write(data)
	return err
}

// читаем из файла и записываем в Storage
func ReadMetricsSnapshot(fileName string) (*MemStorage, error) {
	// открываем файл для чтения
	jsonFile, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic("cannot initialize zap")
	}
	defer logger.Sync()
	logger.Info("reading from to file", zap.String("fileName", fileName))
	log.Printf("reading from file: %s", fileName)
	// read our opened jsonFile as a byte array.
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		log.Fatal(err)
	}

	// we initialize our AllMetrics array
	var allMetrics AllMetrics

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'allMetrics' which we defined above
	err = json.Unmarshal(byteValue, &allMetrics)
	if err != nil {
		return nil, err
	}

	return AllMetricsToMemStorage(&allMetrics)
}

func StartSaveLoop(storeInterval time.Duration, storagePath string, ms *MemStorage) {
	ticker := time.NewTicker(storeInterval)

	for range ticker.C {
		log.Println("writing to file")
		if err := WriteMetricsSnapshot(storagePath, ms); err != nil {
			log.Fatal(err)
		}
	}
}
