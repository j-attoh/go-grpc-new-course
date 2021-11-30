package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/grpc-go-new-course/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	BLOGDATABASE   = "mydb"
	BLOGCOLLECTION = "blog"
)

var (
	ctx        context.Context
	collection *mongo.Collection
)

type server struct {
	blogpb.UnimplementedBlogServiceServer
}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Creating Blog ")

	blog := req.GetBlog()

	blogData := blogItem{
		AuthorID: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}

	result, err := collection.InsertOne(ctx, blogData)

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal error : %v", err))
	}

	pObjectID, ok := result.InsertedID.(primitive.ObjectID)

	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("could not convert to ObjectID : %v", err))
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       pObjectID.Hex(),
			AuthorId: blog.GetAuthorId(),
			Content:  blog.GetContent(),
			Title:    blog.GetTitle()},
	}, nil

}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {

	fmt.Println("Reading blog ")

	blogID := req.GetBlogId()

	pObjectID, err := primitive.ObjectIDFromHex(blogID)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot Parse ID : %q, err : %v", blogID, err))

	}

	filter := bson.M{"_id": pObjectID}
	singleResult := collection.FindOne(ctx, filter)

	var blog blogItem
	if err := singleResult.Decode(&blog); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("No result Found : %v", err))
	}

	return &blogpb.ReadBlogResponse{Blog: &blogpb.Blog{
		Id:       blog.ID.Hex(),
		AuthorId: blog.AuthorID,
		Title:    blog.Title,
		Content:  blog.Content,
	}}, nil

}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {

	fmt.Println("Updating Blog")

	blog := req.GetBlog()
	blogID := blog.GetId()

	pObjectID, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "unable to Parse ID : %q Err : %v", blogID, err)
	}

	filter := bson.M{"_id": pObjectID}
	singleResult := collection.FindOne(ctx, filter)

	var data blogItem

	if err := singleResult.Decode(&data); err != nil {
		return nil, status.Errorf(codes.NotFound, "No document found : %v", err)
	}

	data.AuthorID = blog.AuthorId
	data.Content = blog.Content
	data.Title = blog.Title

	updateResult, err := collection.ReplaceOne(ctx, filter, data)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot update mongo document : %v", err)

	}

	if updateResult.ModifiedCount != 1 {

		return nil, status.Errorf(codes.Internal, "failed to update One document : %v", err)
	}
	return &blogpb.UpdateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Title:    data.Title,
			Content:  data.Content,
		},
	}, nil

}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Deleting Blog")
	blogID := req.GetBlogId()

	pObjectID, err := primitive.ObjectIDFromHex(blogID)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "cannot parse id : %q err : %v", blogID, err)
	}

	filter := bson.M{"_id": pObjectID}
	deletedResult, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "could not deleted document : %v", err)
	}

	if deletedResult.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, "blogs not found to be deleted  %v", err)
	}

	return &blogpb.DeleteBlogResponse{BlogId: req.GetBlogId()}, nil

}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("Listing Blogs")

	filter := bson.M{}
	cursor, err := collection.Find(ctx, filter)

	if err != nil {
		return status.Errorf(codes.Internal, "internal error : %v", err)
	}

	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var data blogItem

		if err := cursor.Decode(&data); err != nil {
			return status.Errorf(codes.Internal, "Error while decoding Mongo document : %v", err)
		}
		stream.Send(&blogpb.ListBlogResponse{
			Blog: &blogpb.Blog{
				Id:       data.ID.Hex(),
				AuthorId: data.AuthorID,
				Title:    data.Title,
				Content:  data.Content,
			},
		})

	}
	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, "Cursor error : %v", err)

	}
	return nil

}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	lis, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
		log.Fatalf("unable to create tcp Listener : %v", err)

	}

	//Creating DB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("unable to create mongodb client : %v", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("client unable to connect : %v", err)
	}

	collection = client.Database(BLOGDATABASE).Collection(BLOGCOLLECTION)

	grpcServer := grpc.NewServer()

	blogpb.RegisterBlogServiceServer(grpcServer, &server{})
	//use reflection for evans cli
	reflection.Register(grpcServer)

	go func() {
		log.Println("Starting the gRPC server")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("unable to initialize server : %v", err)

		}

	}()

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)

	<-ch

	log.Println("Shutting down gRPC server")
	grpcServer.Stop()

	log.Println("Shutting down Mongodb")

	if err := client.Disconnect(ctx); err != nil {
		log.Fatalf("error disconnecting from MONGODB DATABASE : %v", err)
	}

	log.Println("Final shut down ")

}
