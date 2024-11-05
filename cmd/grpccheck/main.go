package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adettelle/go-metric-collector/internal/agent/metricservice"
	pb "github.com/adettelle/go-metric-collector/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	delta := int64(1)
	value := 11.22

	SendMetricsChunkGrpc(0, []metricservice.MetricRequest{
		{ID: "m1", MType: "counter", Delta: &delta},
		{ID: "m2", MType: "gauge", Value: &value},
	})
}

// SendMetricsChunkGrpc sends chunk of metrics, id is number of chunk
func SendMetricsChunkGrpc(id int, chunk []metricservice.MetricRequest) error {
	c, err := grpc.NewClient(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server at localhost:3200: %v", err)
	}
	defer c.Close()
	mClient := pb.NewMetricsClient(c)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// req := pb.UpdateMetricsRequest{Metrics: []*pb.Metric{
	// 	{Name: "m1", Type: "counter", Delta: 101},
	// 	{Name: "m2", Type: "gauge", Value: 1.222},
	// }}

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
