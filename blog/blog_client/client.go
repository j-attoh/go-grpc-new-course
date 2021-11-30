package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/grpc-go-new-course/blog/blogpb"
	"google.golang.org/grpc"
)

var ctx context.Context = context.Background()

func main() {

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	if err != nil {
		log.Fatalf("unable to connect : %v", err)
	}
	defer conn.Close()

	client := blogpb.NewBlogServiceClient(conn)

	/*
		blogID := createblog(client)
		readblog(client, blogID)
		updateblog(client, blogID)
		deleteblog(client, blogID)
	*/
	listblog(client)

}

func createblog(client blogpb.BlogServiceClient) string {

	fmt.Println("Calling CreateBlog ")
	req := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "Jonathan Attram",
			Title:    "ToBeDeleted  Blog",
			Content:  "Content of the ToBeDeleted  Blog",
		},
	}

	resp, err := client.CreateBlog(ctx, req)

	if err != nil {
		log.Fatalf("error calling rpc CreateBlog : %v", err)

	}
	log.Printf("Response : Blog Created : %v \n ", resp.GetBlog())
	return resp.GetBlog().GetId()

}

func readblog(client blogpb.BlogServiceClient, blogID string) {
	fmt.Println("Calling ReadBlog ")

	req := &blogpb.ReadBlogRequest{BlogId: blogID}

	resp, err := client.ReadBlog(ctx, req)

	if err != nil {
		log.Fatalf("error calling rpc ReadBlog : %v", err)

	}

	log.Printf("ReadBlog Response : %v \n ", resp.GetBlog())

}

func updateblog(client blogpb.BlogServiceClient, blogID string) {
	fmt.Println("calling updateblog")

	req := &blogpb.UpdateBlogRequest{
		Blog: &blogpb.Blog{
			Id:       blogID,
			AuthorId: "Jonathan Attram (edited)",
			Title:    "ToBeDeleted  blog  (edited)",
			Content:  "Content of ToBeDeleted  Blog,.. additions made",
		},
	}

	resp, err := client.UpdateBlog(ctx, req)

	if err != nil {
		log.Fatalf("error calling rpc UpdateBlog : %v", err)
	}

	log.Printf("Update Response : %v \n", resp.GetBlog())

}

func deleteblog(client blogpb.BlogServiceClient, blogID string) {
	fmt.Println("calling deleteblog ")
	req := &blogpb.DeleteBlogRequest{BlogId: blogID}

	resp, err := client.DeleteBlog(ctx, req)

	if err != nil {
		log.Fatalf("error calling rpc DeleteBlog : %v", err)
	}

	log.Printf("Deleted Blog Respnse : %v \n", resp)

}

func listblog(client blogpb.BlogServiceClient) {
	fmt.Println("calling listblog")

	req := &blogpb.ListBlogRequest{}
	stream, err := client.ListBlog(ctx, req)

	if err != nil {
		log.Fatalf("error calling ListBlog rpc : %v", err)
	}

	for {

		res, err := stream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error receiving server stream : %v", err)
		}
		log.Printf("List Blog : Response : %v \n", res.GetBlog())

	}

}
