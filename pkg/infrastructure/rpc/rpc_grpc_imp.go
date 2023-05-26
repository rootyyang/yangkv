package rpc

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rootyyang/yangkv/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

func init() {
	registerRPC("grpc", &gRPCClientProvider{}, &gRPCServerProvider{})
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
	statMutex      sync.RWMutex
}

func (rpcc *gRPCClient) Start() error {
	var err error
	rpcc.statMutex.Lock()
	defer rpcc.statMutex.Unlock()
	rpcc.gRPCConn, err = grpc.Dial(rpcc.remoteNodeAddr, grpc.WithInsecure(), grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoff.Config{
		BaseDelay:  10 * time.Millisecond,
		Multiplier: 1.6,
		MaxDelay:   1 * time.Second,
	}}))
	if err != nil {
		return err
	}
	grpc.WithDefaultCallOptions()
	rpcc.gRPCClient = proto.NewNodeServiceClient(rpcc.gRPCConn)
	return nil
}
func (rpcc *gRPCClient) Stop() error {
	rpcc.statMutex.RLock()
	defer rpcc.statMutex.RUnlock()
	rpcc.gRPCConn.Close()
	return nil
}

func (rpcc *gRPCClient) AppendEntires(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error) {
	var retResp *AppendEntiresResp
	rpcc.statMutex.RLock()
	defer rpcc.statMutex.RUnlock()
	if rpcc.gRPCClient == nil {
		return nil, fmt.Errorf("rpc client not start")
	}
	resp, err := rpcc.gRPCClient.AppendEntires(ctx, &proto.AppendEntiresReq{LeaderTerm: req.LeaderTerm, LeaderId: req.LeaderId})
	if resp != nil {
		retResp = &AppendEntiresResp{Term: resp.Term, Success: resp.Success}
	}
	return retResp, err
}
func (rpcc *gRPCClient) RequestVote(ctx context.Context, req *RequestVoteReq) (*RequestVoteResp, error) {
	var retResp *RequestVoteResp
	rpcc.statMutex.RLock()
	defer rpcc.statMutex.RUnlock()
	if rpcc.gRPCClient == nil {
		return nil, fmt.Errorf("rpc client not start")
	}
	resp, err := rpcc.gRPCClient.RequestVote(ctx, &proto.RequestVoteReq{CandidateTerm: req.Term, CandidateId: req.CandidateId})
	if resp != nil {
		retResp = &RequestVoteResp{Term: resp.Term, VoteGranted: resp.VoteGranted}
	}
	return retResp, err
}

type gRPCServerProvider struct {
}

func (*gRPCServerProvider) GetRPCServer(localNodePort string) RPCServer {
	return &gRPCServer{port: localNodePort}
}

type gRPCServer struct {
	proto.UnimplementedNodeServiceServer
	handle   RaftServiceHandler
	listener net.Listener
	server   *grpc.Server
	port     string
	mutex    sync.Mutex
}

func (rpcs *gRPCServer) Start() error {
	var err error
	rpcs.mutex.Lock()
	defer rpcs.mutex.Unlock()
	rpcs.listener, err = net.Listen("tcp", ":"+rpcs.port)
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
	rpcs.mutex.Lock()
	defer rpcs.mutex.Unlock()
	if rpcs.server != nil {
		rpcs.server.Stop()
	}
	return nil
}
func (rpcs *gRPCServer) RegisterRaftHandle(pRaftHandler RaftServiceHandler) error {
	rpcs.handle = pRaftHandler
	return nil
}

func (rpcs *gRPCServer) AppendEntires(ctx context.Context, req *proto.AppendEntiresReq) (*proto.AppendEtriesResp, error) {
	var retResp *proto.AppendEtriesResp
	resp, err := rpcs.handle.AppendEntires(ctx, &AppendEntiresReq{LeaderTerm: req.LeaderTerm, LeaderId: req.LeaderId})
	if resp != nil {
		retResp = &proto.AppendEtriesResp{Term: resp.Term, Success: resp.Success}
	}
	return retResp, err
}
func (rpcs *gRPCServer) RequestVote(ctx context.Context, req *proto.RequestVoteReq) (*proto.RequestVoteResp, error) {
	var retResp *proto.RequestVoteResp
	resp, err := rpcs.handle.RequestVote(ctx, &RequestVoteReq{Term: req.CandidateTerm, CandidateId: req.CandidateId})
	if resp != nil {
		retResp = &proto.RequestVoteResp{Term: resp.Term, VoteGranted: resp.VoteGranted}
	}
	return retResp, err
}
