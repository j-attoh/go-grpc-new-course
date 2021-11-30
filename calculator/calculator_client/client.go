package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/grpc-go-new-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {

	conn, err := grpc.Dial("localhost:50000", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("could not establish client connection %v", err)
	}

	defer conn.Close()

	client := calculatorpb.NewCalculatorServiceClient(conn)
	//doServerStreaming(client)

	//doUnary(client)
	//doClientStreaming(client)
	//BiDi(client)
	doUnaryError(client)

}

func doUnary(client calculatorpb.CalculatorServiceClient) {
	req := calculatorpb.SumRequest{
		FirstNumber:  900,
		SecondNumber: 678,
	}
	resp, err := client.Sum(context.Background(), &req)

	if err != nil {
		log.Fatalf("rpc Sum call failed : %v", err)

	}

	log.Printf("Response is  : %v", resp)

}

func doServerStreaming(client calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Number: 120,
	}
	stream, err := client.PrimeNumberDecomposition(context.Background(), req)

	if err != nil {
		log.Fatalf("rpc PrimeNumberDecomposition failed : %v", err)

	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Server Streaming of results failed : %v", err)
		}

		log.Printf(" Server Streaming result : %v \n", res.GetPrimeFactor())

	}

}
func doClientStreaming(client calculatorpb.CalculatorServiceClient) {

	numbers := []int64{1, 2, 3, 4}
	var requests []*calculatorpb.ComputeAverageRequest
	for _, number := range numbers {
		requests = append(requests, &calculatorpb.ComputeAverageRequest{
			Number: number,
		})

	}

	stream, err := client.ComputeAverage(context.Background())

	if err != nil {
		log.Fatalf("error calling rpc ComputeAverage : %v", err)

	}

	for _, req := range requests {
		fmt.Printf("sending request : %v \n", req)
		stream.Send(req)

	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error getting Compute Average Server Response : %v", err)

	}
	log.Printf(" Response : Average is %v", resp.GetAverage())

}

func BiDi(client calculatorpb.CalculatorServiceClient) {
	stream, err := client.FindMaximum(context.Background())

	if err != nil {
		log.Fatalf("error calling rpc FindMaximum : %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {

		defer wg.Done()
		numbers := []int32{4, 7, 2, 19, 4, 6, 32}

		for _, number := range numbers {
			fmt.Printf("sending number : %v \n ", number)
			stream.Send(&calculatorpb.FindMaximumRequest{
				Number: number,
			})
			time.Sleep(time.Second)

		}
		stream.CloseSend()

	}()

	go func() {
		defer wg.Done()
		for {

			resp, err := stream.Recv()

			if err == io.EOF {
				break

			}
			if err != nil {
				log.Fatalf("error reading server streams : %v", err)
			}
			log.Printf("The maximum is : %v \n", resp.GetMaximum())
		}

	}()

	wg.Wait()

}

func doUnaryError(client calculatorpb.CalculatorServiceClient) {

	errCall := func(client calculatorpb.CalculatorServiceClient, number int32) {

		req := &calculatorpb.SquareRootRequest{Number: number}
		res, err := client.SquareRoot(context.Background(), req)

		if err != nil {
			status, ok := status.FromError(err)
			if ok {
				log.Println(status.Message())

				if status.Code() == codes.InvalidArgument {
					log.Fatalf("error sending Invalid argument : %v", err)

				}

			}
			log.Fatalf("Error calling rpc  SquareRoot : %v", err)

		}

		log.Printf("The result of sqaure root of %q  is : %v", fmt.Sprint(number), res.GetSqrRoot())
	}

	//errCall(client, 15)
	errCall(client, -23)

}
