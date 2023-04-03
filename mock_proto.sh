#!/bin/bash
rm  -rf ./pkg/proto/*.pb.go


export PATH="$PATH:$(go env GOPATH)/bin"
protoc --proto_path=. ./pkg/proto/*.proto  --go_out=. --go-grpc_out=.


mockgen -source=./pkg/infrastructure/rpc/rpc_interface.go -destination=./pkg/infrastructure/rpc/rpc_interface_mock.go -package=rpc

mockgen -source=./pkg/infrastructure/log/log_interface.go -destination=./pkg/infrastructure/log/log_interface_mock.go -package=log

mockgen -source=./pkg/infrastructure/config/config_interface.go -destination=./pkg/infrastructure/config/config_interface_mock.go -package=config


mockgen -source=./pkg/domain//node/node_interface.go -destination=./pkg/domain/node/node_interface_mock.go -package=node
mockgen -source=./pkg/domain/node/node_with_heartbeat.go -destination=./pkg/domain/node/node_with_heartbeat_mock.go -package=node



mockgen -source=./pkg/infrastructure/system/time.go -destination=./pkg/infrastructure/system/time_mock.go -package=system




mockgen -source=./pkg/domain/cluster/master_manager.go -destination=./pkg/domain/cluster/master_manager_mock.go -package=cluster

mockgen -source=./pkg/domain/cluster/server_request_handler.go -destination=./pkg/domain/cluster/server_request_handler_mock.go -package=cluster



mockgen -source=./pkg/domain/cluster/node_manager_interface.go -destination=./pkg/domain/cluster/node_manager_interface_mock.go -package=cluster