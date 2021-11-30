

blog-server:
	go run blog/blog_server/server.go  
blog-client:
	go run blog/blog_client/client.go 

greet: greet/greetpb/greet.proto  
	protoc  --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  greet/greetpb/greet.proto


calculator: calculator/calculatorpb/calculator.proto
	protoc  --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  calculator/calculatorpb/calculator.proto 


blog: blog/blogpb/blog.proto 
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative blog/blogpb/blog.proto 

