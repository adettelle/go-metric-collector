// Package retries provides utility functions for executing operations with retries.
// It offers a generic function to retry a given action multiple times with an incremental delay.
package retries

import (
	"log"
	"time"
)

// RunWithRetries is a generic function.
// It executes a provided function multiple times with a delay between each retry.
// The function will be retried until it either succeeds (returns nil error)
// or the retry limit is reached.
// The delay between retries increases after each attempt. It also checks if the error is retriable.
// Parameters:
//   - title: A descriptive name of the action being retried, used for logging.
//   - count: The maximum number of retry attempts.
//   - f: The function to be executed. It should return a pointer to type T and an error.
//   - isRetriableError: A function that determines if the error returned by 'f' is retriable.
func RunWithRetries[T any](
	title string, count int,
	f func() (*T, error),
	isRetriableError func(error) bool) (*T, error) {
	delay := time.Duration(time.Second * 1)
	for i := 0; i < count+1; i++ {
		log.Printf("Executing action '%s': attempt %d\n", title, i)
		res, err := f()
		if err == nil {
			return res, nil
		} else {
			log.Printf("error while executing action '%s': %v", title, err)
			if i == 3 || !isRetriableError(err) { // дается попытки: через 1, 3, 5 сек
				return nil, err
			}
		}
		<-time.NewTicker(delay).C
		delay += delay + time.Duration(time.Second*2)
	}

	return nil, nil
}
