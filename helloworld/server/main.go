// Package main implements a server for Greeter service.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
	commpb "github.com/lpy-neo/micro-framework-test/comm_proto"
	pb "github.com/lpy-neo/micro-framework-test/helloworld/proto"
	"google.golang.org/grpc"
)

const (
	port = ":50052"
)

type server struct {
}

const (
	SayHelloCmd = 1000 //should put into a global file
)

func (s *server) GrpcReq(ctx context.Context, in *commpb.GrpcRequest) (*commpb.GrpcReply, error) {
	log.Printf("Received: %v", *in)

	switch in.Head.Cmd {
	case SayHelloCmd:
		var req pb.HelloRequest
		if in.Head.Encoding == 2 {
			if err := json.Unmarshal(in.Body, &req); err != nil {
				return nil, err
			}
		} else {
			if err := proto.Unmarshal(in.Body, &req); err != nil {
				return nil, err
			}
		}
		rsp, err := s.SayHello(ctx, &req)
		if err != nil {
			return nil, err
		}
		var bytes []byte
		if in.Head.Encoding == 2 {
			bytes, err = json.Marshal(rsp)
			if err != nil {
				return nil, err
			}
		} else {
			bytes, err = proto.Marshal(rsp)
			if err != nil {
				return nil, err
			}
		}
		return &commpb.GrpcReply{Cmd: in.Head.Cmd, Data: bytes}, nil
	}

	return nil, fmt.Errorf("wrong cmd %d", in.Head.Cmd)
}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", *in)

	return &pb.HelloReply{Message: in.Name}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
