proto:
	protoc -I/usr/local/include -I. \
		-I$GOPATH/src \
		--go_out=plugins=grpc,paths=source_relative:. \
		comm_proto/grpc.proto
	protoc -I/usr/local/include -I. \
		-I$GOPATH/src \
		--go_out=plugins=grpc,paths=source_relative:. \
		helloworld/proto/helloworld.proto
	protoc -I/usr/local/include -I. \
		-I$GOPATH/src \
		--go_out=plugins=grpc,paths=source_relative:. \
		grpc-gateway/proto/gateway.proto
