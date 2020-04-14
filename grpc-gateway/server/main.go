// Package main implements a server for Greeter service.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/lpy-neo/micro-framework-test/comm_proto"
	commpb "github.com/lpy-neo/micro-framework-test/comm_proto"
	pb "github.com/lpy-neo/micro-framework-test/grpc-gateway/proto"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
}

func (s *server) GrpcReq(ctx context.Context, in *commpb.GrpcRequest) (*commpb.GrpcReply, error) {
	return processGrpcReq(ctx, in)
}

func (s *server) GrpcStream(stream pb.GrpcService_GrpcStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		rsp, err := processGrpcReq(context.Background(), req)
		if err != nil {
			return err
		}
		err = stream.Send(rsp)
		if err != nil {
			return err
		}
	}

	return nil
}

func processGrpcReq(ctx context.Context, in *commpb.GrpcRequest) (*commpb.GrpcReply, error) {
	log.Printf("Received: %v", *in)

	cmd := in.Head.Cmd
	svrAddr := ""
	serviceName := ""
	switch {
	case cmd >= 1000 && cmd < 2000:
		svrAddr = "localhost:50052"
		serviceName = "helloworld.Greeter"
	default:
		return nil, fmt.Errorf("wrong cmd %d", cmd)
	}
	conn, err := grpc.Dial(svrAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	out := new(comm_proto.GrpcReply)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = conn.Invoke(ctx, "/"+serviceName+"/GrpcReq", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func main() {
	go grpcSvr()
	go restSvr()
	select {}
}

func grpcSvr() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGrpcServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func restSvr() {
	http.HandleFunc("/grpc_gateway.RestService", httpHandler)
	http.HandleFunc("/grpc_gateway.WsService", wsHandler)
	http.ListenAndServe(":50050", nil)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	log.Printf("%s", string(body))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	var req commpb.GrpcRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	rsp, err := processGrpcReq(context.Background(), &req)
	bytes, err := json.Marshal(rsp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	_, err = w.Write(bytes)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

var upgrader = websocket.Upgrader{} // use default options
func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)

		var req commpb.GrpcRequest
		if err := json.Unmarshal(message, &req); err != nil {
			break
		}

		rsp, err := processGrpcReq(context.Background(), &req)
		bytes, err := json.Marshal(rsp)
		if err != nil {
			break
		}
		err = c.WriteMessage(mt, bytes)
		if err != nil {
			break
		}
	}
	log.Println("ws break")
}
