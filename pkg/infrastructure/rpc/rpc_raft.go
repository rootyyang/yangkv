package rpc

import "context"

type AppendEntiresReq struct {
	LeaderTerm int32
	LeaderId   string
}

type AppendEtriesResp struct {
	Term    int32
	Success bool
}

type RequestVoteReq struct {
	Term        int32
	CandidateId string
}

type RequestVoteResp struct {
	Term        int32
	VoteGranted bool
}

type RaftServiceHandler interface {
	AppendEntires(ctx context.Context, req *AppendEntiresReq) (*AppendEtriesResp, error)
	RequestVote(ctx context.Context, req *RequestVoteReq) (*RequestVoteResp, error)
}
