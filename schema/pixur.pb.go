// Code generated by protoc-gen-go.
// source: pixur.proto
// DO NOT EDIT!

/*
Package schema is a generated protocol buffer package.

It is generated from these files:
	pixur.proto

It has these top-level messages:
	Pic
	PicIdent
	AnimationInfo
	Tag
	PicTag
	User
*/
package schema

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/duration"
import google_protobuf1 "github.com/golang/protobuf/ptypes/timestamp"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
const _ = proto.ProtoPackageIsVersion1

type Pic_Mime int32

const (
	Pic_UNKNOWN Pic_Mime = 0
	Pic_JPEG    Pic_Mime = 1
	Pic_GIF     Pic_Mime = 2
	Pic_PNG     Pic_Mime = 3
	Pic_WEBM    Pic_Mime = 4
)

var Pic_Mime_name = map[int32]string{
	0: "UNKNOWN",
	1: "JPEG",
	2: "GIF",
	3: "PNG",
	4: "WEBM",
}
var Pic_Mime_value = map[string]int32{
	"UNKNOWN": 0,
	"JPEG":    1,
	"GIF":     2,
	"PNG":     3,
	"WEBM":    4,
}

func (x Pic_Mime) String() string {
	return proto.EnumName(Pic_Mime_name, int32(x))
}
func (Pic_Mime) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type Pic_DeletionStatus_Reason int32

const (
	// The reason is not know, due to limitations of proto
	Pic_DeletionStatus_UNKNOWN Pic_DeletionStatus_Reason = 0
	// No specific reason.  This is a catch-all reason.
	Pic_DeletionStatus_NONE Pic_DeletionStatus_Reason = 1
	// The pic is in violation of the rules.
	Pic_DeletionStatus_RULE_VIOLATION Pic_DeletionStatus_Reason = 2
)

var Pic_DeletionStatus_Reason_name = map[int32]string{
	0: "UNKNOWN",
	1: "NONE",
	2: "RULE_VIOLATION",
}
var Pic_DeletionStatus_Reason_value = map[string]int32{
	"UNKNOWN":        0,
	"NONE":           1,
	"RULE_VIOLATION": 2,
}

func (x Pic_DeletionStatus_Reason) String() string {
	return proto.EnumName(Pic_DeletionStatus_Reason_name, int32(x))
}
func (Pic_DeletionStatus_Reason) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{0, 0, 0}
}

type PicIdent_Type int32

const (
	PicIdent_UNKNOWN PicIdent_Type = 0
	PicIdent_SHA256  PicIdent_Type = 1
	PicIdent_SHA1    PicIdent_Type = 2
	PicIdent_MD5     PicIdent_Type = 3
	PicIdent_DCT_0   PicIdent_Type = 4
)

var PicIdent_Type_name = map[int32]string{
	0: "UNKNOWN",
	1: "SHA256",
	2: "SHA1",
	3: "MD5",
	4: "DCT_0",
}
var PicIdent_Type_value = map[string]int32{
	"UNKNOWN": 0,
	"SHA256":  1,
	"SHA1":    2,
	"MD5":     3,
	"DCT_0":   4,
}

func (x PicIdent_Type) String() string {
	return proto.EnumName(PicIdent_Type_name, int32(x))
}
func (PicIdent_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1, 0} }

type User_Capability int32

const (
	User_UNKNOWN    User_Capability = 0
	User_CREATE_PIC User_Capability = 1
	User_VIEW_PIC   User_Capability = 2
)

var User_Capability_name = map[int32]string{
	0: "UNKNOWN",
	1: "CREATE_PIC",
	2: "VIEW_PIC",
}
var User_Capability_value = map[string]int32{
	"UNKNOWN":    0,
	"CREATE_PIC": 1,
	"VIEW_PIC":   2,
}

func (x User_Capability) String() string {
	return proto.EnumName(User_Capability_name, int32(x))
}
func (User_Capability) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{5, 0} }

type Pic struct {
	PicId      int64                       `protobuf:"varint,1,opt,name=pic_id,json=picId" json:"pic_id,omitempty"`
	FileSize   int64                       `protobuf:"varint,2,opt,name=file_size,json=fileSize" json:"file_size,omitempty"`
	Mime       Pic_Mime                    `protobuf:"varint,3,opt,name=mime,enum=pixur.Pic_Mime" json:"mime,omitempty"`
	Width      int64                       `protobuf:"varint,4,opt,name=width" json:"width,omitempty"`
	Height     int64                       `protobuf:"varint,5,opt,name=height" json:"height,omitempty"`
	CreatedTs  *google_protobuf1.Timestamp `protobuf:"bytes,10,opt,name=created_ts,json=createdTs" json:"created_ts,omitempty"`
	ModifiedTs *google_protobuf1.Timestamp `protobuf:"bytes,11,opt,name=modified_ts,json=modifiedTs" json:"modified_ts,omitempty"`
	// If present, the pic is on the path to removal.  When the pic is marked
	// for deletion, it is delisted from normal indexing operations.  When the
	// pic is actually "deleted" only the pic object is removed.
	DeletionStatus *Pic_DeletionStatus `protobuf:"bytes,12,opt,name=deletion_status,json=deletionStatus" json:"deletion_status,omitempty"`
	// Only present on animated images.
	AnimationInfo *AnimationInfo    `protobuf:"bytes,13,opt,name=animation_info,json=animationInfo" json:"animation_info,omitempty"`
	ViewCount     int64             `protobuf:"varint,14,opt,name=view_count,json=viewCount" json:"view_count,omitempty"`
	Source        []*Pic_FileSource `protobuf:"bytes,15,rep,name=source" json:"source,omitempty"`
	FileName      []string          `protobuf:"bytes,16,rep,name=file_name,json=fileName" json:"file_name,omitempty"`
}

func (m *Pic) Reset()                    { *m = Pic{} }
func (m *Pic) String() string            { return proto.CompactTextString(m) }
func (*Pic) ProtoMessage()               {}
func (*Pic) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Pic) GetCreatedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.CreatedTs
	}
	return nil
}

func (m *Pic) GetModifiedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.ModifiedTs
	}
	return nil
}

func (m *Pic) GetDeletionStatus() *Pic_DeletionStatus {
	if m != nil {
		return m.DeletionStatus
	}
	return nil
}

func (m *Pic) GetAnimationInfo() *AnimationInfo {
	if m != nil {
		return m.AnimationInfo
	}
	return nil
}

func (m *Pic) GetSource() []*Pic_FileSource {
	if m != nil {
		return m.Source
	}
	return nil
}

type Pic_DeletionStatus struct {
	// Represents when this Pic was marked for deletion
	MarkedDeletedTs *google_protobuf1.Timestamp `protobuf:"bytes,1,opt,name=marked_deleted_ts,json=markedDeletedTs" json:"marked_deleted_ts,omitempty"`
	// Represents when this picture will be auto deleted.  Note that the Pic
	// may exist for a short period after this time.  (may be absent)
	PendingDeletedTs *google_protobuf1.Timestamp `protobuf:"bytes,2,opt,name=pending_deleted_ts,json=pendingDeletedTs" json:"pending_deleted_ts,omitempty"`
	// Determines when Pic was actually deleted.  (present after the Pic is
	// hard deleted, a.k.a purging)
	ActualDeletedTs *google_protobuf1.Timestamp `protobuf:"bytes,3,opt,name=actual_deleted_ts,json=actualDeletedTs" json:"actual_deleted_ts,omitempty"`
	// Gives an explanation for why this pic was removed.
	Details string `protobuf:"bytes,4,opt,name=details" json:"details,omitempty"`
	// The reason the pic was removed.
	Reason Pic_DeletionStatus_Reason `protobuf:"varint,5,opt,name=reason,enum=pixur.Pic_DeletionStatus_Reason" json:"reason,omitempty"`
	// Determines if this pic can be undeleted if re uploaded.  Currently the
	// only reason is due to disk space concerns.
	Temporary bool `protobuf:"varint,6,opt,name=temporary" json:"temporary,omitempty"`
}

func (m *Pic_DeletionStatus) Reset()                    { *m = Pic_DeletionStatus{} }
func (m *Pic_DeletionStatus) String() string            { return proto.CompactTextString(m) }
func (*Pic_DeletionStatus) ProtoMessage()               {}
func (*Pic_DeletionStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

func (m *Pic_DeletionStatus) GetMarkedDeletedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.MarkedDeletedTs
	}
	return nil
}

func (m *Pic_DeletionStatus) GetPendingDeletedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.PendingDeletedTs
	}
	return nil
}

func (m *Pic_DeletionStatus) GetActualDeletedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.ActualDeletedTs
	}
	return nil
}

type Pic_FileSource struct {
	Url       string                      `protobuf:"bytes,1,opt,name=url" json:"url,omitempty"`
	Referrer  string                      `protobuf:"bytes,2,opt,name=referrer" json:"referrer,omitempty"`
	CreatedTs *google_protobuf1.Timestamp `protobuf:"bytes,3,opt,name=created_ts,json=createdTs" json:"created_ts,omitempty"`
}

func (m *Pic_FileSource) Reset()                    { *m = Pic_FileSource{} }
func (m *Pic_FileSource) String() string            { return proto.CompactTextString(m) }
func (*Pic_FileSource) ProtoMessage()               {}
func (*Pic_FileSource) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 1} }

func (m *Pic_FileSource) GetCreatedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.CreatedTs
	}
	return nil
}

// A picture identifier
type PicIdent struct {
	PicId int64         `protobuf:"varint,1,opt,name=pic_id,json=picId" json:"pic_id,omitempty"`
	Type  PicIdent_Type `protobuf:"varint,2,opt,name=type,enum=pixur.PicIdent_Type" json:"type,omitempty"`
	Value []byte        `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	// dct0 are the upper 8x8 corner of the 32x32 dct of the image
	Dct0Values []float32 `protobuf:"fixed32,4,rep,packed,name=dct0_values,json=dct0Values" json:"dct0_values,omitempty"`
}

func (m *PicIdent) Reset()                    { *m = PicIdent{} }
func (m *PicIdent) String() string            { return proto.CompactTextString(m) }
func (*PicIdent) ProtoMessage()               {}
func (*PicIdent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type AnimationInfo struct {
	// How long this animated image in time.  There must be more than 1 frame
	// for this value to be set.
	Duration *google_protobuf.Duration `protobuf:"bytes,1,opt,name=duration" json:"duration,omitempty"`
}

func (m *AnimationInfo) Reset()                    { *m = AnimationInfo{} }
func (m *AnimationInfo) String() string            { return proto.CompactTextString(m) }
func (*AnimationInfo) ProtoMessage()               {}
func (*AnimationInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *AnimationInfo) GetDuration() *google_protobuf.Duration {
	if m != nil {
		return m.Duration
	}
	return nil
}

type Tag struct {
	TagId      int64                       `protobuf:"varint,1,opt,name=tag_id,json=tagId" json:"tag_id,omitempty"`
	Name       string                      `protobuf:"bytes,2,opt,name=name" json:"name,omitempty"`
	UsageCount int64                       `protobuf:"varint,3,opt,name=usage_count,json=usageCount" json:"usage_count,omitempty"`
	CreatedTs  *google_protobuf1.Timestamp `protobuf:"bytes,6,opt,name=created_ts,json=createdTs" json:"created_ts,omitempty"`
	ModifiedTs *google_protobuf1.Timestamp `protobuf:"bytes,7,opt,name=modified_ts,json=modifiedTs" json:"modified_ts,omitempty"`
}

func (m *Tag) Reset()                    { *m = Tag{} }
func (m *Tag) String() string            { return proto.CompactTextString(m) }
func (*Tag) ProtoMessage()               {}
func (*Tag) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Tag) GetCreatedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.CreatedTs
	}
	return nil
}

func (m *Tag) GetModifiedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.ModifiedTs
	}
	return nil
}

type PicTag struct {
	PicId      int64                       `protobuf:"varint,1,opt,name=pic_id,json=picId" json:"pic_id,omitempty"`
	TagId      int64                       `protobuf:"varint,2,opt,name=tag_id,json=tagId" json:"tag_id,omitempty"`
	Name       string                      `protobuf:"bytes,3,opt,name=name" json:"name,omitempty"`
	CreatedTs  *google_protobuf1.Timestamp `protobuf:"bytes,6,opt,name=created_ts,json=createdTs" json:"created_ts,omitempty"`
	ModifiedTs *google_protobuf1.Timestamp `protobuf:"bytes,7,opt,name=modified_ts,json=modifiedTs" json:"modified_ts,omitempty"`
}

func (m *PicTag) Reset()                    { *m = PicTag{} }
func (m *PicTag) String() string            { return proto.CompactTextString(m) }
func (*PicTag) ProtoMessage()               {}
func (*PicTag) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PicTag) GetCreatedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.CreatedTs
	}
	return nil
}

func (m *PicTag) GetModifiedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.ModifiedTs
	}
	return nil
}

type User struct {
	UserId int64 `protobuf:"varint,1,opt,name=user_id,json=userId" json:"user_id,omitempty"`
	// Hashed secret token
	Secret     []byte                      `protobuf:"bytes,2,opt,name=secret,proto3" json:"secret,omitempty"`
	Email      string                      `protobuf:"bytes,3,opt,name=email" json:"email,omitempty"`
	CreatedTs  *google_protobuf1.Timestamp `protobuf:"bytes,4,opt,name=created_ts,json=createdTs" json:"created_ts,omitempty"`
	ModifiedTs *google_protobuf1.Timestamp `protobuf:"bytes,5,opt,name=modified_ts,json=modifiedTs" json:"modified_ts,omitempty"`
	LastSeenTs *google_protobuf1.Timestamp `protobuf:"bytes,6,opt,name=last_seen_ts,json=lastSeenTs" json:"last_seen_ts,omitempty"`
	Capability []User_Capability           `protobuf:"varint,7,rep,name=capability,enum=pixur.User_Capability" json:"capability,omitempty"`
}

func (m *User) Reset()                    { *m = User{} }
func (m *User) String() string            { return proto.CompactTextString(m) }
func (*User) ProtoMessage()               {}
func (*User) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *User) GetCreatedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.CreatedTs
	}
	return nil
}

func (m *User) GetModifiedTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.ModifiedTs
	}
	return nil
}

func (m *User) GetLastSeenTs() *google_protobuf1.Timestamp {
	if m != nil {
		return m.LastSeenTs
	}
	return nil
}

func init() {
	proto.RegisterType((*Pic)(nil), "pixur.Pic")
	proto.RegisterType((*Pic_DeletionStatus)(nil), "pixur.Pic.DeletionStatus")
	proto.RegisterType((*Pic_FileSource)(nil), "pixur.Pic.FileSource")
	proto.RegisterType((*PicIdent)(nil), "pixur.PicIdent")
	proto.RegisterType((*AnimationInfo)(nil), "pixur.AnimationInfo")
	proto.RegisterType((*Tag)(nil), "pixur.Tag")
	proto.RegisterType((*PicTag)(nil), "pixur.PicTag")
	proto.RegisterType((*User)(nil), "pixur.User")
	proto.RegisterEnum("pixur.Pic_Mime", Pic_Mime_name, Pic_Mime_value)
	proto.RegisterEnum("pixur.Pic_DeletionStatus_Reason", Pic_DeletionStatus_Reason_name, Pic_DeletionStatus_Reason_value)
	proto.RegisterEnum("pixur.PicIdent_Type", PicIdent_Type_name, PicIdent_Type_value)
	proto.RegisterEnum("pixur.User_Capability", User_Capability_name, User_Capability_value)
}

var fileDescriptor0 = []byte{
	// 949 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xc4, 0x55, 0xcd, 0x6e, 0xdb, 0x46,
	0x10, 0x2e, 0x45, 0x8a, 0x92, 0x46, 0x32, 0xcd, 0x2c, 0x92, 0x94, 0x51, 0x7f, 0x22, 0x28, 0x17,
	0x5d, 0xaa, 0xa4, 0x0a, 0x9c, 0x36, 0x48, 0x7b, 0x90, 0xf5, 0x63, 0xab, 0x8d, 0x65, 0x61, 0x45,
	0xdb, 0x40, 0x2f, 0xc4, 0x9a, 0x5c, 0xc9, 0x8b, 0xf2, 0x0f, 0xe4, 0x32, 0xa9, 0xf3, 0x16, 0x7d,
	0x99, 0x3e, 0x40, 0x81, 0x1e, 0xfa, 0x3e, 0x7d, 0x80, 0x62, 0x97, 0xd4, 0x9f, 0x93, 0xc0, 0x35,
	0x72, 0xe8, 0x8d, 0xdf, 0xcc, 0x37, 0x83, 0xf9, 0x66, 0x76, 0x86, 0x50, 0x8f, 0xd9, 0x6f, 0x59,
	0xd2, 0x8d, 0x93, 0x88, 0x47, 0xa8, 0x2c, 0x41, 0xf3, 0xeb, 0x65, 0x14, 0x2d, 0x7d, 0xfa, 0x54,
	0x1a, 0x2f, 0xb3, 0xc5, 0x53, 0x2f, 0x4b, 0x08, 0x67, 0x51, 0x98, 0xd3, 0x9a, 0x8f, 0x6f, 0xfa,
	0x39, 0x0b, 0x68, 0xca, 0x49, 0x10, 0xe7, 0x84, 0xf6, 0x1f, 0x55, 0x50, 0x67, 0xcc, 0x45, 0x0f,
	0x40, 0x8f, 0x99, 0xeb, 0x30, 0xcf, 0x52, 0x5a, 0x4a, 0x47, 0xc5, 0xe5, 0x98, 0xb9, 0x13, 0x0f,
	0x7d, 0x01, 0xb5, 0x05, 0xf3, 0xa9, 0x93, 0xb2, 0x77, 0xd4, 0x2a, 0x49, 0x4f, 0x55, 0x18, 0xe6,
	0xec, 0x1d, 0x45, 0x4f, 0x40, 0x0b, 0x58, 0x40, 0x2d, 0xb5, 0xa5, 0x74, 0x8c, 0xde, 0x7e, 0x37,
	0xaf, 0x6f, 0xc6, 0xdc, 0xee, 0x09, 0x0b, 0x28, 0x96, 0x4e, 0x74, 0x1f, 0xca, 0x6f, 0x99, 0xc7,
	0xaf, 0x2c, 0x2d, 0xcf, 0x2b, 0x01, 0x7a, 0x08, 0xfa, 0x15, 0x65, 0xcb, 0x2b, 0x6e, 0x95, 0xa5,
	0xb9, 0x40, 0xe8, 0x25, 0x80, 0x9b, 0x50, 0xc2, 0xa9, 0xe7, 0xf0, 0xd4, 0x82, 0x96, 0xd2, 0xa9,
	0xf7, 0x9a, 0xdd, 0x5c, 0x44, 0x77, 0x25, 0xa2, 0x6b, 0xaf, 0x44, 0xe0, 0x5a, 0xc1, 0xb6, 0x53,
	0xf4, 0x0a, 0xea, 0x41, 0xe4, 0xb1, 0x05, 0xcb, 0x63, 0xeb, 0xb7, 0xc6, 0xc2, 0x8a, 0x6e, 0xa7,
	0xe8, 0x10, 0xf6, 0x3d, 0xea, 0x53, 0xd1, 0x39, 0x27, 0xe5, 0x84, 0x67, 0xa9, 0xd5, 0x90, 0x09,
	0x1e, 0x6d, 0xa9, 0x1a, 0x16, 0x8c, 0xb9, 0x24, 0x60, 0xc3, 0xdb, 0xc1, 0xe8, 0x15, 0x18, 0x24,
	0x64, 0x81, 0x6c, 0xbf, 0xc3, 0xc2, 0x45, 0x64, 0xed, 0xc9, 0x14, 0xf7, 0x8b, 0x14, 0xfd, 0x95,
	0x73, 0x12, 0x2e, 0x22, 0xbc, 0x47, 0xb6, 0x21, 0xfa, 0x0a, 0xe0, 0x0d, 0xa3, 0x6f, 0x1d, 0x37,
	0xca, 0x42, 0x6e, 0x19, 0xb2, 0x29, 0x35, 0x61, 0x19, 0x08, 0x03, 0xfa, 0x06, 0xf4, 0x34, 0xca,
	0x12, 0x97, 0x5a, 0xfb, 0x2d, 0xb5, 0x53, 0xef, 0x3d, 0xd8, 0x2a, 0x6b, 0x2c, 0xe6, 0x21, 0x9d,
	0xb8, 0x20, 0xad, 0xc7, 0x16, 0x92, 0x80, 0x5a, 0x66, 0x4b, 0xed, 0xd4, 0xf2, 0xb1, 0x4d, 0x49,
	0x40, 0x9b, 0xbf, 0xab, 0x60, 0xec, 0x4a, 0x41, 0x63, 0xb8, 0x17, 0x90, 0xe4, 0x57, 0xea, 0x39,
	0x52, 0x53, 0xde, 0x41, 0xe5, 0xd6, 0x0e, 0xee, 0xe7, 0x41, 0xc3, 0x3c, 0xc6, 0x4e, 0xd1, 0x31,
	0xa0, 0x98, 0x86, 0x1e, 0x0b, 0x97, 0xdb, 0x89, 0x4a, 0xb7, 0x26, 0x32, 0x8b, 0xa8, 0x4d, 0xa6,
	0x31, 0xdc, 0x23, 0x2e, 0xcf, 0x88, 0xbf, 0x9d, 0x48, 0xbd, 0xbd, 0xa2, 0x3c, 0x68, 0x93, 0xc7,
	0x82, 0x8a, 0x47, 0x39, 0x61, 0x7e, 0x2a, 0x1f, 0x60, 0x0d, 0xaf, 0x20, 0xfa, 0x1e, 0xf4, 0x84,
	0x92, 0x34, 0x0a, 0xe5, 0x13, 0x34, 0x7a, 0xad, 0x8f, 0x4e, 0xba, 0x8b, 0x25, 0x0f, 0x17, 0x7c,
	0xf4, 0x25, 0xd4, 0x38, 0x0d, 0xe2, 0x28, 0x21, 0xc9, 0xb5, 0xa5, 0xb7, 0x94, 0x4e, 0x15, 0x6f,
	0x0c, 0xed, 0xe7, 0xa0, 0xe7, 0x7c, 0x54, 0x87, 0xca, 0xd9, 0xf4, 0xe7, 0xe9, 0xe9, 0xc5, 0xd4,
	0xfc, 0x0c, 0x55, 0x41, 0x9b, 0x9e, 0x4e, 0x47, 0xa6, 0x82, 0x10, 0x18, 0xf8, 0xec, 0xf5, 0xc8,
	0x39, 0x9f, 0x9c, 0xbe, 0xee, 0xdb, 0x93, 0xd3, 0xa9, 0x59, 0x6a, 0x66, 0x00, 0x9b, 0x31, 0x22,
	0x13, 0xd4, 0x2c, 0xf1, 0xe5, 0x00, 0x6a, 0x58, 0x7c, 0xa2, 0x26, 0x54, 0x13, 0xba, 0xa0, 0x49,
	0x42, 0x13, 0xd9, 0xce, 0x1a, 0x5e, 0xe3, 0x1b, 0x3b, 0xa3, 0xde, 0x61, 0x67, 0xda, 0x2f, 0x41,
	0x13, 0xab, 0xfa, 0x5e, 0xa5, 0x3f, 0xcd, 0x46, 0x47, 0xa6, 0x82, 0x2a, 0xa0, 0x1e, 0x4d, 0xc6,
	0x66, 0x49, 0x7c, 0xcc, 0xa6, 0x47, 0xa6, 0x2a, 0x7c, 0x17, 0xa3, 0xc3, 0x13, 0x53, 0x6b, 0xff,
	0xa5, 0x40, 0x75, 0x26, 0x6e, 0x04, 0x0d, 0xf9, 0xc7, 0xae, 0x47, 0x07, 0x34, 0x7e, 0x1d, 0xe7,
	0x87, 0xc3, 0x58, 0xef, 0xc1, 0x2a, 0xaa, 0x6b, 0x5f, 0xc7, 0x14, 0x4b, 0x86, 0xb8, 0x12, 0x6f,
	0x88, 0x9f, 0xe5, 0xb7, 0xa4, 0x81, 0x73, 0x80, 0x9e, 0x40, 0xdd, 0x73, 0xf9, 0x33, 0x47, 0x22,
	0x31, 0x40, 0xb5, 0x53, 0x3a, 0x2c, 0x99, 0x0a, 0x06, 0x61, 0x3e, 0x97, 0xd6, 0xf6, 0x8f, 0xa0,
	0x89, 0x44, 0xbb, 0x1a, 0x00, 0xf4, 0xf9, 0x71, 0xbf, 0x77, 0xf0, 0xc2, 0x54, 0x44, 0xcd, 0xf3,
	0xe3, 0xfe, 0xb7, 0xb9, 0x8c, 0x93, 0xe1, 0x81, 0xa9, 0xa2, 0x1a, 0x94, 0x87, 0x03, 0xdb, 0x79,
	0x66, 0x6a, 0xed, 0x31, 0xec, 0xed, 0x2c, 0x26, 0x3a, 0x80, 0xea, 0xea, 0x88, 0x16, 0x2b, 0xf0,
	0xe8, 0xbd, 0x66, 0x0e, 0x0b, 0x02, 0x5e, 0x53, 0xdb, 0x7f, 0x2b, 0xa0, 0xda, 0x64, 0x29, 0x5a,
	0xc1, 0xc9, 0x72, 0xab, 0x15, 0x9c, 0x2c, 0x27, 0x1e, 0x42, 0xa0, 0xc9, 0x65, 0xcc, 0x87, 0x27,
	0xbf, 0xd1, 0x63, 0xa8, 0x67, 0x29, 0x59, 0xd2, 0x62, 0xe9, 0x55, 0xc9, 0x07, 0x69, 0xca, 0xb7,
	0x7e, 0x77, 0xb2, 0xfa, 0x27, 0x5c, 0xc3, 0xca, 0x5d, 0xae, 0x61, 0xfb, 0x4f, 0x05, 0xf4, 0x19,
	0x73, 0x0b, 0x39, 0x1f, 0x9a, 0xec, 0x46, 0x65, 0xe9, 0x43, 0x2a, 0xd5, 0x2d, 0x95, 0xff, 0x97,
	0x88, 0x7f, 0x4a, 0xa0, 0x9d, 0xa5, 0x34, 0x41, 0x9f, 0x43, 0x25, 0x4b, 0x69, 0xb2, 0xd1, 0xa0,
	0x0b, 0x38, 0xf1, 0xc4, 0x4f, 0x28, 0xa5, 0x6e, 0x42, 0xb9, 0x14, 0xd1, 0xc0, 0x05, 0x12, 0x8f,
	0x91, 0x06, 0x84, 0xf9, 0x85, 0x8c, 0x1c, 0xdc, 0xd0, 0xa1, 0x7d, 0x82, 0x8e, 0xf2, 0x9d, 0x7e,
	0x4d, 0x3f, 0x40, 0xc3, 0x27, 0x29, 0x77, 0x52, 0x4a, 0xc3, 0xff, 0xd6, 0x41, 0x10, 0xfc, 0x39,
	0xa5, 0xa1, 0x9d, 0xa2, 0x17, 0x00, 0x2e, 0x89, 0xc9, 0x25, 0xf3, 0x19, 0xbf, 0xb6, 0x2a, 0x2d,
	0xb5, 0x63, 0xf4, 0x1e, 0x16, 0x8b, 0x28, 0xba, 0xd3, 0x1d, 0xac, 0xbd, 0x78, 0x8b, 0xd9, 0xfe,
	0x0e, 0x60, 0xe3, 0xd9, 0xdd, 0x2d, 0x03, 0x60, 0x80, 0x47, 0x7d, 0x7b, 0xe4, 0xcc, 0x26, 0x03,
	0x53, 0x41, 0x0d, 0xa8, 0x9e, 0x4f, 0x46, 0x17, 0x12, 0x95, 0x0e, 0xab, 0xbf, 0xe8, 0xa9, 0x7b,
	0x45, 0x03, 0x72, 0xa9, 0xcb, 0xd2, 0x9e, 0xff, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x71, 0x6d, 0xee,
	0x4e, 0xb8, 0x08, 0x00, 0x00,
}
