package metricservice

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/adettelle/go-metric-collector/internal/security"
	"github.com/adettelle/go-metric-collector/pkg/retries"
)

type HTTPSender struct {
	Client            *http.Client
	URL               string
	EncryptionKey     string
	MaxRequestRetries int
	// publicKey         *rsa.PublicKey
}

func NewHTTPSender(client *http.Client, url string, maxRequestRetries int, encryptionKey string) *HTTPSender {
	return &HTTPSender{
		Client:            client,
		URL:               url,
		MaxRequestRetries: maxRequestRetries,
		EncryptionKey:     encryptionKey,
	}
}

// SendMetricsChunk sends chunk of metrics, id is number of chunk
func (c *HTTPSender) SendMetricsChunk(id int, chunk []MetricRequest) error {
	var err error

	log.Printf("Sending chunk on worker %d\n", id)

	data, err := json.Marshal(chunk)
	if err != nil {
		log.Printf("error %v in sending chun in worker %d\n", err, id)
		return err // прерываем итерацию, но не сам worker и не цикл
	}

	_, err = retries.RunWithRetries("Send metrics request",
		c.MaxRequestRetries,
		func() (*any, error) {
			// nolint:staticcheck
			err = c.doSend(bytes.NewBuffer(data))
			return nil, err
		}, isRetriableError)

	if err != nil {
		log.Printf("error %v in sending chun in worker %d\n", err, id)
		return err
	}

	log.Printf("chunk in worker %d sent successfully\n", id)
	return nil
}

func (c *HTTPSender) doSend(data *bytes.Buffer) error {
	req, err := http.NewRequest(http.MethodPost, c.URL, data)
	if err != nil {
		return err
	}

	if c.EncryptionKey != "" {
		// вычисляем хеш и передаем в HTTP-заголовке запроса с именем HashSHA256
		hash := security.CreateSign(data.String(), c.EncryptionKey)
		log.Println(data.String(), hash)
		req.Header.Set("HashSHA256", hash)
	}

	netAddr := "127.0.0.1"

	req.Header.Set("X-Real-IP", netAddr)

	resp, err := c.Client.Do(req)
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
