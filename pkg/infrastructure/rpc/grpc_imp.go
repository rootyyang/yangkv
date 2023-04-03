package rpc

import (
	"context"
	"fmt"
	"net"

	"github.com/rootyyang/yangkv/pkg/proto"
	"google.golang.org/grpc"
)

func init() {
	registerRPC("grpc", &gRPCClientProvider{}, &gRPCServer{})
}

type gRPCClientProvider struct {
}

func (*gRPCClientProvider) GetRPCClient(remoteNodeIP string, remoteNodePort string) RPClient {
	return &gRPCClient{remoteNodeAddr: remoteNodeIP + ":" + remoteNodePort}
}

type gRPCClient struct {
	gRPCClient     proto.NodeServiceClient
	gRPCConn       *grpc.ClientConn
	remoteNodeAddr string
}

func (rpcc *gRPCClient) Start() error {
	var err error
	rpcc.gRPCConn, err = grpc.Dial(rpcc.remoteNodeAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	rpcc.gRPCClient = proto.NewNodeServiceClient(rpcc.gRPCConn)
	return nil
}
func (rpcc *gRPCClient) Stop() error {
	rpcc.gRPCConn.Close()
	return nil
}
func (rpcc *gRPCClient) HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
	return rpcc.gRPCClient.HeartBeat(ctx, req)
}
func (rpcc *gRPCClient) LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
	return rpcc.gRPCClient.LeaveCluster(ctx, req)
}

type gRPCServer struct {
	proto.UnimplementedNodeServiceServer
	handle   RPCSeverHandleInterface
	listener net.Listener
	server   *grpc.Server
	port     string
}

func (rpcs *gRPCServer) Start(pPort string, pRPCHandleFunc RPCSeverHandleInterface) error {
	var err error
	if pRPCHandleFunc == nil {
		return fmt.Errorf("pRPCHandleFunc is nil")
	}
	rpcs.handle = pRPCHandleFunc
	rpcs.listener, err = net.Listen("tcp", ":"+pPort)
	if err != nil {
		return err
	}
	rpcs.server = grpc.NewServer()
	proto.RegisterNodeServiceServer(rpcs.server, rpcs)
	go func() {
		if err := rpcs.server.Serve(rpcs.listener); err != nil {
			rpcs.listener.Close()
		}
	}()
	return nil
}

func (rpcs *gRPCServer) Stop() error {
	rpcs.server.Stop()
	return nil
}
func (rpcs *gRPCServer) HeartBeat(ctx context.Context, req *proto.HeartBeatReq) (*proto.HeartBeatResp, error) {
	return rpcs.handle.HeartBeat(ctx, req)
}
func (rpcs *gRPCServer) LeaveCluster(ctx context.Context, req *proto.LeaveClusterReq) (*proto.LeaveClusterResp, error) {
	return rpcs.handle.LeaveCluster(ctx, req)
}
