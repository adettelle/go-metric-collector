package metricservice

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/adettelle/go-metric-collector/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcClient struct {
	url string
}

func NewGrpcSender(url string) *GrpcClient {
	return &GrpcClient{url: url}
}

// SendMetricsChunk sends chunk of metrics, id is number of chunk
func (c *GrpcClient) SendMetricsChunk(id int, chunk []MetricRequest) error {
	client, err := grpc.NewClient(c.url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server at localhost:3200: %v", err)
	}
	defer client.Close()
	mClient := pb.NewMetricsClient(client)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	pbMetrics := []*pb.Metric{}
	for _, mreq := range chunk {
		var pbm pb.Metric
		if mreq.MType == "counter" {
			pbm = pb.Metric{Name: mreq.ID, Type: mreq.MType, Delta: *mreq.Delta}
		} else {
			pbm = pb.Metric{Name: mreq.ID, Type: mreq.MType, Value: *mreq.Value}
		}
		pbMetrics = append(pbMetrics, &pbm)
	}
	res, err := mClient.UpdateMetrics(ctx, &pb.UpdateMetricsRequest{Metrics: pbMetrics}) // &req
	if err != nil {
		return err
	}
	if res.Error != "" {
		return fmt.Errorf("error while sending metrics chunks to grpc: %v", res.Error)
	}

	log.Printf("Response from gRPC server's UpdateMetrics function is OK")
	return nil
}
