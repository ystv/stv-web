// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.20.3
// source: storage.proto

package storage

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type STV struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ballots           []*Ballot    `protobuf:"bytes,1,rep,name=ballots,proto3" json:"ballots,omitempty"`
	Candidates        []*Candidate `protobuf:"bytes,2,rep,name=candidates,proto3" json:"candidates,omitempty"`
	Elections         []*Election  `protobuf:"bytes,3,rep,name=elections,proto3" json:"elections,omitempty"`
	Urls              []*URL       `protobuf:"bytes,4,rep,name=urls,proto3" json:"urls,omitempty"`
	Voters            []*Voter     `protobuf:"bytes,5,rep,name=voters,proto3" json:"voters,omitempty"`
	AllowRegistration bool         `protobuf:"varint,6,opt,name=allowRegistration,proto3" json:"allowRegistration,omitempty"`
}

func (x *STV) Reset() {
	*x = STV{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *STV) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*STV) ProtoMessage() {}

func (x *STV) ProtoReflect() protoreflect.Message {
	mi := &file_storage_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use STV.ProtoReflect.Descriptor instead.
func (*STV) Descriptor() ([]byte, []int) {
	return file_storage_proto_rawDescGZIP(), []int{0}
}

func (x *STV) GetBallots() []*Ballot {
	if x != nil {
		return x.Ballots
	}
	return nil
}

func (x *STV) GetCandidates() []*Candidate {
	if x != nil {
		return x.Candidates
	}
	return nil
}

func (x *STV) GetElections() []*Election {
	if x != nil {
		return x.Elections
	}
	return nil
}

func (x *STV) GetUrls() []*URL {
	if x != nil {
		return x.Urls
	}
	return nil
}

func (x *STV) GetVoters() []*Voter {
	if x != nil {
		return x.Voters
	}
	return nil
}

func (x *STV) GetAllowRegistration() bool {
	if x != nil {
		return x.AllowRegistration
	}
	return false
}

type Ballot struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Election string            `protobuf:"bytes,2,opt,name=election,proto3" json:"election,omitempty"`
	Voter    string            `protobuf:"bytes,3,opt,name=voter,proto3" json:"voter,omitempty"`
	Choice   map[uint64]string `protobuf:"bytes,4,rep,name=choice,proto3" json:"choice,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"` // map[order, candidate id]
}

func (x *Ballot) Reset() {
	*x = Ballot{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Ballot) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Ballot) ProtoMessage() {}

func (x *Ballot) ProtoReflect() protoreflect.Message {
	mi := &file_storage_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Ballot.ProtoReflect.Descriptor instead.
func (*Ballot) Descriptor() ([]byte, []int) {
	return file_storage_proto_rawDescGZIP(), []int{1}
}

func (x *Ballot) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Ballot) GetElection() string {
	if x != nil {
		return x.Election
	}
	return ""
}

func (x *Ballot) GetVoter() string {
	if x != nil {
		return x.Voter
	}
	return ""
}

func (x *Ballot) GetChoice() map[uint64]string {
	if x != nil {
		return x.Choice
	}
	return nil
}

type Candidate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Election string `protobuf:"bytes,2,opt,name=election,proto3" json:"election,omitempty"`
	Name     string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Candidate) Reset() {
	*x = Candidate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Candidate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Candidate) ProtoMessage() {}

func (x *Candidate) ProtoReflect() protoreflect.Message {
	mi := &file_storage_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Candidate.ProtoReflect.Descriptor instead.
func (*Candidate) Descriptor() ([]byte, []int) {
	return file_storage_proto_rawDescGZIP(), []int{2}
}

func (x *Candidate) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Candidate) GetElection() string {
	if x != nil {
		return x.Election
	}
	return ""
}

func (x *Candidate) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type Election struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description string   `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Ron         bool     `protobuf:"varint,4,opt,name=ron,proto3" json:"ron,omitempty"`
	Open        bool     `protobuf:"varint,5,opt,name=open,proto3" json:"open,omitempty"`
	Closed      bool     `protobuf:"varint,6,opt,name=closed,proto3" json:"closed,omitempty"`
	Result      *Result  `protobuf:"bytes,7,opt,name=result,proto3" json:"result,omitempty"`
	Excluded    []*Voter `protobuf:"bytes,8,rep,name=excluded,proto3" json:"excluded,omitempty"`
	Voters      uint64   `protobuf:"varint,9,opt,name=voters,proto3" json:"voters,omitempty"`
}

func (x *Election) Reset() {
	*x = Election{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Election) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Election) ProtoMessage() {}

func (x *Election) ProtoReflect() protoreflect.Message {
	mi := &file_storage_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Election.ProtoReflect.Descriptor instead.
func (*Election) Descriptor() ([]byte, []int) {
	return file_storage_proto_rawDescGZIP(), []int{3}
}

func (x *Election) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Election) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Election) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Election) GetRon() bool {
	if x != nil {
		return x.Ron
	}
	return false
}

func (x *Election) GetOpen() bool {
	if x != nil {
		return x.Open
	}
	return false
}

func (x *Election) GetClosed() bool {
	if x != nil {
		return x.Closed
	}
	return false
}

func (x *Election) GetResult() *Result {
	if x != nil {
		return x.Result
	}
	return nil
}

func (x *Election) GetExcluded() []*Voter {
	if x != nil {
		return x.Excluded
	}
	return nil
}

func (x *Election) GetVoters() uint64 {
	if x != nil {
		return x.Voters
	}
	return 0
}

type Result struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Rounds uint64   `protobuf:"varint,1,opt,name=rounds,proto3" json:"rounds,omitempty"`
	Winner string   `protobuf:"bytes,2,opt,name=winner,proto3" json:"winner,omitempty"`
	Round  []*Round `protobuf:"bytes,3,rep,name=round,proto3" json:"round,omitempty"`
}

func (x *Result) Reset() {
	*x = Result{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Result) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Result) ProtoMessage() {}

func (x *Result) ProtoReflect() protoreflect.Message {
	mi := &file_storage_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Result.ProtoReflect.Descriptor instead.
func (*Result) Descriptor() ([]byte, []int) {
	return file_storage_proto_rawDescGZIP(), []int{4}
}

func (x *Result) GetRounds() uint64 {
	if x != nil {
		return x.Rounds
	}
	return 0
}

func (x *Result) GetWinner() string {
	if x != nil {
		return x.Winner
	}
	return ""
}

func (x *Result) GetRound() []*Round {
	if x != nil {
		return x.Round
	}
	return nil
}

type Round struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Round           uint64             `protobuf:"varint,1,opt,name=round,proto3" json:"round,omitempty"`
	Blanks          uint64             `protobuf:"varint,2,opt,name=blanks,proto3" json:"blanks,omitempty"`
	CandidateStatus []*CandidateStatus `protobuf:"bytes,3,rep,name=candidateStatus,proto3" json:"candidateStatus,omitempty"`
}

func (x *Round) Reset() {
	*x = Round{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Round) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Round) ProtoMessage() {}

func (x *Round) ProtoReflect() protoreflect.Message {
	mi := &file_storage_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Round.ProtoReflect.Descriptor instead.
func (*Round) Descriptor() ([]byte, []int) {
	return file_storage_proto_rawDescGZIP(), []int{5}
}

func (x *Round) GetRound() uint64 {
	if x != nil {
		return x.Round
	}
	return 0
}

func (x *Round) GetBlanks() uint64 {
	if x != nil {
		return x.Blanks
	}
	return 0
}

func (x *Round) GetCandidateStatus() []*CandidateStatus {
	if x != nil {
		return x.CandidateStatus
	}
	return nil
}

type CandidateStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CandidateRank uint64  `protobuf:"varint,1,opt,name=candidateRank,proto3" json:"candidateRank,omitempty"`
	Id            string  `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	NoOfVotes     float64 `protobuf:"fixed64,3,opt,name=noOfVotes,proto3" json:"noOfVotes,omitempty"`
	Status        string  `protobuf:"bytes,4,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *CandidateStatus) Reset() {
	*x = CandidateStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CandidateStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CandidateStatus) ProtoMessage() {}

func (x *CandidateStatus) ProtoReflect() protoreflect.Message {
	mi := &file_storage_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CandidateStatus.ProtoReflect.Descriptor instead.
func (*CandidateStatus) Descriptor() ([]byte, []int) {
	return file_storage_proto_rawDescGZIP(), []int{6}
}

func (x *CandidateStatus) GetCandidateRank() uint64 {
	if x != nil {
		return x.CandidateRank
	}
	return 0
}

func (x *CandidateStatus) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *CandidateStatus) GetNoOfVotes() float64 {
	if x != nil {
		return x.NoOfVotes
	}
	return 0
}

func (x *CandidateStatus) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

type URL struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url      string `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Election string `protobuf:"bytes,2,opt,name=election,proto3" json:"election,omitempty"`
	Voter    string `protobuf:"bytes,3,opt,name=voter,proto3" json:"voter,omitempty"`
	Voted    bool   `protobuf:"varint,4,opt,name=voted,proto3" json:"voted,omitempty"`
}

func (x *URL) Reset() {
	*x = URL{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *URL) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*URL) ProtoMessage() {}

func (x *URL) ProtoReflect() protoreflect.Message {
	mi := &file_storage_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use URL.ProtoReflect.Descriptor instead.
func (*URL) Descriptor() ([]byte, []int) {
	return file_storage_proto_rawDescGZIP(), []int{7}
}

func (x *URL) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *URL) GetElection() string {
	if x != nil {
		return x.Election
	}
	return ""
}

func (x *URL) GetVoter() string {
	if x != nil {
		return x.Voter
	}
	return ""
}

func (x *URL) GetVoted() bool {
	if x != nil {
		return x.Voted
	}
	return false
}

type Voter struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Email string `protobuf:"bytes,1,opt,name=email,proto3" json:"email,omitempty"`
	Name  string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *Voter) Reset() {
	*x = Voter{}
	if protoimpl.UnsafeEnabled {
		mi := &file_storage_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Voter) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Voter) ProtoMessage() {}

func (x *Voter) ProtoReflect() protoreflect.Message {
	mi := &file_storage_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Voter.ProtoReflect.Descriptor instead.
func (*Voter) Descriptor() ([]byte, []int) {
	return file_storage_proto_rawDescGZIP(), []int{8}
}

func (x *Voter) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *Voter) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

var File_storage_proto protoreflect.FileDescriptor

var file_storage_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x22, 0x8d, 0x02, 0x0a, 0x03, 0x53, 0x54, 0x56,
	0x12, 0x29, 0x0a, 0x07, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x0f, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x42, 0x61, 0x6c, 0x6c,
	0x6f, 0x74, 0x52, 0x07, 0x62, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x73, 0x12, 0x32, 0x0a, 0x0a, 0x63,
	0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x12, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64,
	0x61, 0x74, 0x65, 0x52, 0x0a, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x73, 0x12,
	0x2f, 0x0a, 0x09, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x03, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x11, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x45, 0x6c, 0x65,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x09, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x12, 0x20, 0x0a, 0x04, 0x75, 0x72, 0x6c, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c,
	0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x55, 0x52, 0x4c, 0x52, 0x04, 0x75, 0x72,
	0x6c, 0x73, 0x12, 0x26, 0x0a, 0x06, 0x76, 0x6f, 0x74, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x56, 0x6f, 0x74,
	0x65, 0x72, 0x52, 0x06, 0x76, 0x6f, 0x74, 0x65, 0x72, 0x73, 0x12, 0x2c, 0x0a, 0x11, 0x61, 0x6c,
	0x6c, 0x6f, 0x77, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x11, 0x61, 0x6c, 0x6c, 0x6f, 0x77, 0x52, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0xba, 0x01, 0x0a, 0x06, 0x42, 0x61, 0x6c,
	0x6c, 0x6f, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x6f, 0x74, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x6f, 0x74, 0x65, 0x72, 0x12, 0x33, 0x0a, 0x06, 0x63, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x18,
	0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e,
	0x42, 0x61, 0x6c, 0x6c, 0x6f, 0x74, 0x2e, 0x43, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x06, 0x63, 0x68, 0x6f, 0x69, 0x63, 0x65, 0x1a, 0x39, 0x0a, 0x0b, 0x43, 0x68,
	0x6f, 0x69, 0x63, 0x65, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x4b, 0x0a, 0x09, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x22, 0xfb, 0x01, 0x0a, 0x08, 0x45, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x10, 0x0a, 0x03, 0x72, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x03, 0x72, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x6f, 0x70, 0x65, 0x6e, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x6f, 0x70, 0x65, 0x6e, 0x12, 0x16, 0x0a, 0x06, 0x63,
	0x6c, 0x6f, 0x73, 0x65, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x63, 0x6c, 0x6f,
	0x73, 0x65, 0x64, 0x12, 0x27, 0x0a, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x52, 0x06, 0x72, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x2a, 0x0a, 0x08,
	0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x64, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x56, 0x6f, 0x74, 0x65, 0x72, 0x52, 0x08,
	0x65, 0x78, 0x63, 0x6c, 0x75, 0x64, 0x65, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x6f, 0x74, 0x65,
	0x72, 0x73, 0x18, 0x09, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x76, 0x6f, 0x74, 0x65, 0x72, 0x73,
	0x22, 0x5e, 0x0a, 0x06, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x6f,
	0x75, 0x6e, 0x64, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x72, 0x6f, 0x75, 0x6e,
	0x64, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x69, 0x6e, 0x6e, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x77, 0x69, 0x6e, 0x6e, 0x65, 0x72, 0x12, 0x24, 0x0a, 0x05, 0x72, 0x6f,
	0x75, 0x6e, 0x64, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x73, 0x74, 0x6f, 0x72,
	0x61, 0x67, 0x65, 0x2e, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64,
	0x22, 0x79, 0x0a, 0x05, 0x52, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75,
	0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x62, 0x6c, 0x61, 0x6e, 0x6b, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52,
	0x06, 0x62, 0x6c, 0x61, 0x6e, 0x6b, 0x73, 0x12, 0x42, 0x0a, 0x0f, 0x63, 0x61, 0x6e, 0x64, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x18, 0x2e, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x2e, 0x43, 0x61, 0x6e, 0x64, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x0f, 0x63, 0x61, 0x6e, 0x64,
	0x69, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x7d, 0x0a, 0x0f, 0x43,
	0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x24,
	0x0a, 0x0d, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x52, 0x61, 0x6e, 0x6b, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0d, 0x63, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x52, 0x61, 0x6e, 0x6b, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x6f, 0x4f, 0x66, 0x56, 0x6f, 0x74, 0x65,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x09, 0x6e, 0x6f, 0x4f, 0x66, 0x56, 0x6f, 0x74,
	0x65, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x5f, 0x0a, 0x03, 0x55, 0x52,
	0x4c, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x75, 0x72, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x6f, 0x74, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x76, 0x6f, 0x74, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x6f, 0x74, 0x65, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x76, 0x6f, 0x74, 0x65, 0x64, 0x22, 0x31, 0x0a, 0x05, 0x56,
	0x6f, 0x74, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6d, 0x61, 0x69, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x42, 0x21,
	0x5a, 0x1f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x79, 0x73, 0x74,
	0x76, 0x2f, 0x73, 0x74, 0x76, 0x2d, 0x77, 0x65, 0x62, 0x2f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67,
	0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_storage_proto_rawDescOnce sync.Once
	file_storage_proto_rawDescData = file_storage_proto_rawDesc
)

func file_storage_proto_rawDescGZIP() []byte {
	file_storage_proto_rawDescOnce.Do(func() {
		file_storage_proto_rawDescData = protoimpl.X.CompressGZIP(file_storage_proto_rawDescData)
	})
	return file_storage_proto_rawDescData
}

var file_storage_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_storage_proto_goTypes = []interface{}{
	(*STV)(nil),             // 0: storage.STV
	(*Ballot)(nil),          // 1: storage.Ballot
	(*Candidate)(nil),       // 2: storage.Candidate
	(*Election)(nil),        // 3: storage.Election
	(*Result)(nil),          // 4: storage.Result
	(*Round)(nil),           // 5: storage.Round
	(*CandidateStatus)(nil), // 6: storage.CandidateStatus
	(*URL)(nil),             // 7: storage.URL
	(*Voter)(nil),           // 8: storage.Voter
	nil,                     // 9: storage.Ballot.ChoiceEntry
}
var file_storage_proto_depIdxs = []int32{
	1,  // 0: storage.STV.ballots:type_name -> storage.Ballot
	2,  // 1: storage.STV.candidates:type_name -> storage.Candidate
	3,  // 2: storage.STV.elections:type_name -> storage.Election
	7,  // 3: storage.STV.urls:type_name -> storage.URL
	8,  // 4: storage.STV.voters:type_name -> storage.Voter
	9,  // 5: storage.Ballot.choice:type_name -> storage.Ballot.ChoiceEntry
	4,  // 6: storage.Election.result:type_name -> storage.Result
	8,  // 7: storage.Election.excluded:type_name -> storage.Voter
	5,  // 8: storage.Result.round:type_name -> storage.Round
	6,  // 9: storage.Round.candidateStatus:type_name -> storage.CandidateStatus
	10, // [10:10] is the sub-list for method output_type
	10, // [10:10] is the sub-list for method input_type
	10, // [10:10] is the sub-list for extension type_name
	10, // [10:10] is the sub-list for extension extendee
	0,  // [0:10] is the sub-list for field type_name
}

func init() { file_storage_proto_init() }
func file_storage_proto_init() {
	if File_storage_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_storage_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*STV); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_storage_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Ballot); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_storage_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Candidate); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_storage_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Election); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_storage_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Result); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_storage_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Round); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_storage_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CandidateStatus); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_storage_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*URL); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_storage_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Voter); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_storage_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_storage_proto_goTypes,
		DependencyIndexes: file_storage_proto_depIdxs,
		MessageInfos:      file_storage_proto_msgTypes,
	}.Build()
	File_storage_proto = out.File
	file_storage_proto_rawDesc = nil
	file_storage_proto_goTypes = nil
	file_storage_proto_depIdxs = nil
}
