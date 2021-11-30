package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"

	"github.com/grpc-go-new-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	calculatorpb.UnimplementedCalculatorServiceServer
}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {

	log.Printf("Server called with : %v,\n ", req)
	sumResult := req.GetFirstNumber() + req.GetSecondNumber()
	res := &calculatorpb.SumResponse{
		SumResult: sumResult,
	}
	return res, nil

}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {

	fmt.Printf("Received Request : %v \n", req)

	number := req.GetNumber()
	k := int64(2)

	for number > 1 {
		if number%k == 0 {
			res := &calculatorpb.PrimeNumberDecompositionResponse{
				PrimeFactor: k,
			}
			stream.Send(res)
			number /= k

		} else {
			k++

		}

	}

	return nil

}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {

	avg := func(nums []int64) float64 {
		length := len(nums)
		sum := int64(0)
		for _, num := range nums {
			sum += num

		}
		average := float64(sum) / float64(length)
		return average

	}

	var numbers []int64

	for {
		req, err := stream.Recv()
		if err == io.EOF {

			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Average: avg(numbers),
			})

		}
		if err != nil {
			log.Fatalf("error trying to get client stream : %v", err)
		}
		numbers = append(numbers, req.GetNumber())

	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {

	maximum := int32(math.Inf(-1))

	for {

		req, err := stream.Recv()

		if err == io.EOF {

			return nil
		}

		if err != nil {
			log.Fatalf("error trying to get client data streams : %v", err)
		}

		number := req.GetNumber()

		if number > maximum {
			maximum = number
			if err := stream.Send(&calculatorpb.FindMaximumResponse{
				Maximum: maximum,
			}); err != nil {
				log.Fatalf("error trying to send server data stream : %v", err)

			}
		}

	}

}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	number := req.GetNumber()

	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Received negative number : %v", number))

	}

	return &calculatorpb.SquareRootResponse{SqrRoot: math.Sqrt(float64(number))}, nil

}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50000")

	if err != nil {
		log.Fatalf("could not establish a listener : %v", err)

	}
	grpcServer := grpc.NewServer()

	calculatorpb.RegisterCalculatorServiceServer(grpcServer, &server{})

	//Register reflection on GRPC server
	reflection.Register(grpcServer)

	log.Fatal(grpcServer.Serve(lis))

}
