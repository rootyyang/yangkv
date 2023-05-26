#!/bin/bash
rm  -rf ./pkg/proto/*.pb.go


export PATH="$PATH:$(go env GOPATH)/bin"
protoc --proto_path=. ./pkg/proto/*.proto  --go_out=. --go-grpc_out=.


mockgen -source=./pkg/infrastructure/rpc/rpc.go -destination=./pkg/infrastructure/rpc/rpc_mock_imp.go -package=rpc


mockgen -source=./pkg/infrastructure/rpc/rpc_raft_protocol.go -destination=./pkg/infrastructure/rpc/rpc_raft_protocol_mock.go -package=rpc

mockgen -source=./pkg/infrastructure/log/log_interface.go -destination=./pkg/infrastructure/log/log_interface_mock.go -package=log

mockgen -source=./pkg/infrastructure/config/config_interface.go -destination=./pkg/infrastructure/config/config_interface_mock.go -package=config


mockgen -source=./pkg/domain/node/node.go -destination=./pkg/domain/node/node_mock_imp.go -package=node

mockgen -source=./pkg/domain/consensus/consensus_algorithm.go -destination=./pkg/domain/consensus/consensus_algorithm_mock_imp.go -package=consensus



mockgen -source=./pkg/infrastructure/system/time.go -destination=./pkg/infrastructure/system/time_mock.go -package=system

mockgen -source=./pkg/infrastructure/system/random.go -destination=./pkg/infrastructure/system/random_mock.go -package=system


mockgen -source=./pkg/domain/cluster/master_manager.go -destination=./pkg/domain/cluster/master_manager_mock.go -package=cluster

mockgen -source=./pkg/domain/cluster/server_request_handler.go -destination=./pkg/domain/cluster/server_request_handler_mock.go -package=cluster


