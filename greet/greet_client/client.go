package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/grpc-go-new-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	certFile := "ssl/ca.crt"
	creds, err := credentials.NewClientTLSFromFile(certFile, "")

	if err != nil {
		log.Fatalf("error loading client certificates : %v", err)
	}
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(creds))

	if err != nil {
		log.Fatalf("unable to establish channel connection : %v", err)

	}

	defer conn.Close()

	client := greetpb.NewGreetServiceClient(conn)

	//doUnary(client)
	// doServerStreaming(client)
	//	doClientStreaming(client)
	//BiDi(client)
	//doUnaryWithDeadline(client, 4*time.Second)
	//doUnaryWithDeadline(client, 2*time.Second)
	doUnary(client)

}

func doUnary(client greetpb.GreetServiceClient) {

	req := greetpb.GreetingRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Jonathan",
			LastName:  "Attoh",
		},
	}
	resp, err := client.Greet(context.Background(), &req)

	if err != nil {
		log.Fatalf("error while calling Greet rpc : %v", err)

	}
	log.Printf("Response from Greet: %v", resp)

}

func doUnaryWithDeadline(client greetpb.GreetServiceClient, timeout time.Duration) {
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Jonathan",
			LastName:  "Attoh",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := client.GreetWithDeadline(ctx, req)

	if err != nil {

		status, ok := status.FromError(err)

		if ok {
			if status.Code() == codes.DeadlineExceeded {
				log.Println(status.Message())
				log.Fatalf("Timeout , DeadlineExceeded ! : %v", err)
			}
			log.Fatalf("unexpeceted error : %v", err)
		}
		log.Fatalf("error occured calling rpc GreetWithDeadline  : %v", err)
	}

	log.Printf(" Result of Response : %v", res.GetResult())

}

// Client side streaming

func doClientStreaming(client greetpb.GreetServiceClient) {
	var names = []string{"Jonathan", "Ama", "Nicole", "Rumaisah", "Adisa", "Akosua", "Amina"}
	var requests []*greetpb.LongGreetRequest
	for _, name := range names {

		requests = append(requests, &greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: name,
			},
		})
	}

	stream, err := client.LongGreet(context.Background())

	if err != nil {
		log.Fatalf("Error calling rpc LongGreet : %v", err)

	}

	for _, req := range requests {
		log.Printf("Sending request stream : %v \n", req)
		if err := stream.Send(req); err != nil {
			log.Fatalf("errror trying to send request stream : %v", err)

		}
	}

	resp, err := stream.CloseAndRecv()

	if err != nil {
		log.Fatalf("error trying to receive response from LongGreet : %v", err)
	}

	log.Printf("LongGreet Response : %v", resp)

}

// Server Streaming
func doServerStreaming(client greetpb.GreetServiceClient) {

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Jonathan",
			LastName:  "Attoh",
		},
	}

	stream, err := client.GreetManyTimes(context.Background(), req)

	if err != nil {
		log.Fatalf("rpc GreetManyTimes call failed : %v", err)

	}

	// Now time for Streaming

	for {
		resp, err := stream.Recv()

		if err == io.EOF {
			break

		}
		if err != nil {
			log.Fatalf(" Streaming server side results failed : %v", err)

		}

		log.Printf("Server side Result Stream : %v \n", resp)

	}

}

func BiDi(client greetpb.GreetServiceClient) {
	var names = []string{"Jonathan", "Ama", "Nicole", "Rumaisah", "Adisa", "Akosua", "Amina"}
	var requests []*greetpb.GreetEveryoneRequest

	stream, err := client.GreetEveryone(context.Background())

	if err != nil {
		log.Fatalf("rpc Call to GreetEveryone failed : %v", err)

	}

	for _, name := range names {
		requests = append(requests, &greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: name,
			},
		})
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {

		defer wg.Done()
		for _, req := range requests {
			fmt.Printf("sending message : %v \n", req)
			stream.Send(req)
			time.Sleep(500 * time.Millisecond)

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
				log.Fatalf("error trying to receive data streams from server : %v", err)
			}
			fmt.Printf("receiving message : %v \n", resp)

		}

	}()
	wg.Wait()

}
