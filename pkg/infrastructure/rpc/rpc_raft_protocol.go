package rpc

import "context"

type AppendEntiresReq struct {
	LeaderTerm   uint64
	LeaderId     string
	PrevLogIndex uint64
	PrevLogTerm  uint64
	Entries      []string
}

type AppendEntiresResp struct {
	Term    uint64
	Success bool
}

type RequestVoteReq struct {
	Term         uint64
	CandidateId  string
	LastLogIndex uint64
	LastLogTerm  uint64
}

type RequestVoteResp struct {
	Term        uint64
	VoteGranted bool
}

type RaftServiceHandler interface {
	AppendEntires(ctx context.Context, req *AppendEntiresReq) (*AppendEntiresResp, error)
	RequestVote(ctx context.Context, req *RequestVoteReq) (*RequestVoteResp, error)
}
