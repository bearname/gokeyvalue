download:
	go mod donwload

build:
	go build -o gokv.exe cmd/main.go

buildHttpServer:
	go build -o goserver.exe cmd/server/main.go

buildGrpcServer:
	go build -o grpcServer.exe cmd/grpcServer/main.go

buildGrpcBench:
	go build -o grpcBench.exe cmd/grpcBench/main.go

buildGrpcClient:
	go build -o grpcClient.exe cmd/grpcClient/main.go

buildCoord:
	go build -o coord.exe cmd/coordinator/main.go

buildGrpc: buildGrpcServer buildGrpcBench buildGrpcClient buildCoord

client:
	go build -o goclient.exe cmd/client/main.go
	#goclient.exe

runMaster:
	goserver -instanceId=0 -port=8000 -isMaster=true -urls=0!localhost:8000,1!localhost:8001 -volume=inst-0/

runSlave:
	goserver -instanceId=1 -port=8001 -isMaster=false -urls=0!localhost:8000,1!localhost:8001 -volume=inst-1/

bench:
	go test -bench=. -benchtime=100x

protoc:
	#protoc --proto_path=protos --go_out=plugins=grpc:protos protos/cache-service.protos
	#protoc -I protos --go_out=plugins=grpc:protos protos/gokeyvalueServoce.protos
	protoc -I protos --go_out=plugins=grpc:protos protos/gokeyvalue.proto

stopHttp:
	taskkill /f /im goserver.exe

stopGrpc:
	taskkill /f /im grpcServer.exe