package retries

import (
	"log"
	"time"
)

// обобщенные методы (функции) дженерики
func RunWithRetries[T any](title string, count int, f func() (*T, error), isRetriableError func(error) bool) (*T, error) {
	delay := 1 // попытки через 1, 3, 5 сек
	for i := 0; i < count+1; i++ {
		log.Printf("Executing action '%s': attempt %d\n", title, i)
		res, err := f()
		if err == nil {
			return res, nil
		} else {
			log.Printf("error while executing action '%s': %v", title, err)
			if i == 3 || !isRetriableError(err) {
				return nil, err
			}
		}
		<-time.NewTicker(time.Duration(delay) * time.Second).C
		delay += 2
	}

	return nil, nil
}
