package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/grpc-go-new-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct {
	greetpb.UnimplementedGreetServiceServer
}

func (*server) Greet(ctx context.Context, req *greetpb.GreetingRequest) (*greetpb.GreetingResponse, error) {

	log.Printf("Client Request to Server :%v \n ", req)
	firstname := req.GetGreeting().GetFirstName()
	result := "Hello " + firstname

	res := &greetpb.GreetingResponse{
		Result: result,
	}
	return res, nil

}

func (*server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			return nil, status.Errorf(codes.DeadlineExceeded, ctx.Err().Error())

		}
		time.Sleep(time.Second)

	}

	firstname := req.GetGreeting().GetFirstName()
	result := fmt.Sprintf("Hello %s !", firstname)

	return &greetpb.GreetWithDeadlineResponse{Result: result}, nil
}

// GreetManyTimes :=> Server side streaming
func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	firstname := req.GetGreeting().GetFirstName()

	for i := 0; i < 10; i++ {
		result := "Hello " + firstname + " number : " + fmt.Sprint(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(time.Second)

	}
	return nil

}

// GreetEveryOne

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("error while reading client stream : %v", err)
		}

		result := fmt.Sprintf(" Hello %s !", req.GetGreeting().GetFirstName())

		if err := stream.Send(&greetpb.GreetEveryoneResponse{
			Result: result,
		}); err != nil {
			log.Fatalf("error sending data to client : %v", err)
			return nil
		}

	}

}

// LongGreet :=> Client side Streaming
func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {

	result := ""
	for {
		req, err := stream.Recv()

		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})

		}

		if err != nil {
			log.Fatalf("client side streaming failed : %v \n", err)

		}
		result += fmt.Sprintf(" Hello %s ! ", req.GetGreeting().GetFirstName())

	}
}

func main() {

	certFile := "ssl/server.crt"
	keyFile := "ssl/server.pem"

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}

	creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
	if err != nil {
		log.Fatalf("error loading certificates : %v", err)
	}
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	greetpb.RegisterGreetServiceServer(grpcServer, &server{})

	//Register reflection on grpcServer
	reflection.Register(grpcServer)

	log.Fatal(grpcServer.Serve(lis))
}
