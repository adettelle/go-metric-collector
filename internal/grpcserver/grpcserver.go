package grpcserver

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/adettelle/go-metric-collector/internal/api"
	pb "github.com/adettelle/go-metric-collector/proto"
	"google.golang.org/grpc"
)

// MetricsServer поддерживает все необходимые методы сервера.
type MServer struct {
	pb.UnimplementedMetricsServer
	Storager api.Storager
}

// UpdatesMetric реализует интерфейс обновления метрик.
func (ms *MServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	var resp pb.UpdateMetricsResponse
	log.Println("resieved metrics: ", in.Metrics)

	for _, metric := range in.Metrics {
		switch {
		case metric.Type == "gauge":
			if err := ms.Storager.AddGaugeMetric(metric.Name, metric.Value); err != nil {
				resp.Error = err.Error()
				return &resp, err
			}

		case metric.Type == "counter":
			if err := ms.Storager.AddCounterMetric(metric.Name, metric.Delta); err != nil {
				resp.Error = err.Error()
				return &resp, err
			}

		default:
			resp.Error = "No such metric"
			return &resp, fmt.Errorf("err %v", resp.Error)
		}
	}

	resp.Error = ""
	return &resp, nil
}

func StartServer(storager api.Storager, port string) error {
	// определяем порт для сервера
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	// создаём gRPC-сервер без зарегистрированной службы
	s := grpc.NewServer()

	// регистрируем сервис
	pb.RegisterMetricsServer(s, &MServer{Storager: storager})
	log.Printf("Starting grpc server on port: %s", port)

	// получаем запрос gRPC
	if err := s.Serve(listen); err != nil {
		return err
	}
	return nil
}
