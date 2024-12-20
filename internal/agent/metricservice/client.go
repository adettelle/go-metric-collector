package metricservice

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/adettelle/go-metric-collector/internal/security"
	"github.com/adettelle/go-metric-collector/pkg/retries"
)

type Client struct {
	client            *http.Client
	url               string
	encryptionKey     string
	maxRequestRetries int
	publicKey         *rsa.PublicKey
}

// SendMetricsChunk sends chunk of metrics, id is number of chunk
func (c *Client) SendMetricsChunk(id int, chunk []MetricRequest) error {
	var err error

	log.Printf("Sending chunk on worker %d\n", id)

	data, err := json.Marshal(chunk)
	if err != nil {
		log.Printf("error %v in sending chun in worker %d\n", err, id)
		return err // прерываем итерацию, но не сам worker и не цикл
	}

	_, err = retries.RunWithRetries("Send metrics request",
		c.maxRequestRetries,
		func() (*any, error) {
			// nolint:staticcheck
			err = c.doSend(bytes.NewBuffer(data)) // nil
			return nil, err
		}, isRetriableError)

	if err != nil {
		log.Printf("error %v in sending chun in worker %d\n", err, id)
		return err
	}

	log.Printf("chunk in worker %d sent successfully\n", id)
	return nil
}

func (c *Client) doSend(data *bytes.Buffer) error {
	req, err := http.NewRequest(http.MethodPost, c.url, data)
	if err != nil {
		return err
	}

	if c.encryptionKey != "" {
		// вычисляем хеш и передаем в HTTP-заголовке запроса с именем HashSHA256
		hash := security.CreateSign(data.String(), c.encryptionKey)
		log.Println(data.String(), hash)
		req.Header.Set("HashSHA256", hash)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ue := UnsuccessfulStatusError{
			Message: fmt.Sprintf("response is not OK, status: %d", resp.StatusCode),
			Status:  resp.StatusCode, // статус, который пришел в ответе
		}
		return &ue
	}

	return nil
}
