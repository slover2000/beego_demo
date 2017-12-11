package services

import (
	"fmt"
	"time"
	"strconv"
	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/slover2000/prisma"
	"github.com/slover2000/prisma/discovery"
	pb "github.com/slover2000/beego_demo/helloworld"
)

var grpcConn *grpc.ClientConn

func InitHelloServiceClient(addr string, interceptorClient *prisma.InterceptorClient) error {
    // create a resolver
    resolver := discovery.NewEtcdResolver(
		discovery.WithResolverSystem(discovery.GRPCSystem),
		discovery.WithResolverService("hello_service"),
		discovery.WithEnvironment(discovery.Product),
		discovery.WithDialTimeout(5 * time.Second))
	b := grpc.RoundRobin(resolver)

	ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)
    conn, err := grpc.DialContext(
        ctx, 
        addr, 
        grpc.WithInsecure(), 
        grpc.WithBalancer(b),
        grpc.WithUnaryInterceptor(interceptorClient.GRPCUnaryClientInterceptor()), 
        grpc.WithStreamInterceptor(interceptorClient.GRPCStreamClientInterceptor()))
    if err != nil {
        return err
	}

	grpcConn = conn
	return nil
}

func CloseHelloServiceClient() {
	if grpcConn != nil {
		grpcConn.Close();
	}
}

func QueryGrpcDemo() {
	t := time.Now().Second()
	client := pb.NewGreeterClient(grpcConn)
	reqCtx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	resp, err := client.SayHello(reqCtx, &pb.HelloRequest{Name: "world " + strconv.Itoa(t)})
	cancel()
	if err == nil {
		fmt.Printf("%d: Reply is %s\n", t, resp.Message)
	}	
}