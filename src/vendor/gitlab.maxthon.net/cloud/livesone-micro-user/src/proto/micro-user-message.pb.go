// Code generated by protoc-gen-go. DO NOT EDIT.
// source: micro-user-message.proto

package microuser

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ResMsg struct {
	Result               ResCode  `protobuf:"varint,1,opt,name=result,proto3,enum=microuser.ResCode" json:"result,omitempty"`
	Msg                  string   `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ResMsg) Reset()         { *m = ResMsg{} }
func (m *ResMsg) String() string { return proto.CompactTextString(m) }
func (*ResMsg) ProtoMessage()    {}
func (*ResMsg) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{0}
}
func (m *ResMsg) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResMsg.Unmarshal(m, b)
}
func (m *ResMsg) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResMsg.Marshal(b, m, deterministic)
}
func (dst *ResMsg) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResMsg.Merge(dst, src)
}
func (m *ResMsg) XXX_Size() int {
	return xxx_messageInfo_ResMsg.Size(m)
}
func (m *ResMsg) XXX_DiscardUnknown() {
	xxx_messageInfo_ResMsg.DiscardUnknown(m)
}

var xxx_messageInfo_ResMsg proto.InternalMessageInfo

func (m *ResMsg) GetResult() ResCode {
	if m != nil {
		return m.Result
	}
	return ResCode_OK
}

func (m *ResMsg) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

type RegUserInfo struct {
	Pwd                  string   `protobuf:"bytes,1,opt,name=pwd,proto3" json:"pwd,omitempty"`
	Country              int32    `protobuf:"varint,2,opt,name=country,proto3" json:"country,omitempty"`
	Phone                string   `protobuf:"bytes,3,opt,name=phone,proto3" json:"phone,omitempty"`
	Email                string   `protobuf:"bytes,4,opt,name=email,proto3" json:"email,omitempty"`
	Type                 int32    `protobuf:"varint,5,opt,name=type,proto3" json:"type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegUserInfo) Reset()         { *m = RegUserInfo{} }
func (m *RegUserInfo) String() string { return proto.CompactTextString(m) }
func (*RegUserInfo) ProtoMessage()    {}
func (*RegUserInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{1}
}
func (m *RegUserInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegUserInfo.Unmarshal(m, b)
}
func (m *RegUserInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegUserInfo.Marshal(b, m, deterministic)
}
func (dst *RegUserInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegUserInfo.Merge(dst, src)
}
func (m *RegUserInfo) XXX_Size() int {
	return xxx_messageInfo_RegUserInfo.Size(m)
}
func (m *RegUserInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_RegUserInfo.DiscardUnknown(m)
}

var xxx_messageInfo_RegUserInfo proto.InternalMessageInfo

func (m *RegUserInfo) GetPwd() string {
	if m != nil {
		return m.Pwd
	}
	return ""
}

func (m *RegUserInfo) GetCountry() int32 {
	if m != nil {
		return m.Country
	}
	return 0
}

func (m *RegUserInfo) GetPhone() string {
	if m != nil {
		return m.Phone
	}
	return ""
}

func (m *RegUserInfo) GetEmail() string {
	if m != nil {
		return m.Email
	}
	return ""
}

func (m *RegUserInfo) GetType() int32 {
	if m != nil {
		return m.Type
	}
	return 0
}

type SetUserInfoReq struct {
	Uid                  int64     `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Field                UserField `protobuf:"varint,2,opt,name=field,proto3,enum=microuser.UserField" json:"field,omitempty"`
	Value                string    `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *SetUserInfoReq) Reset()         { *m = SetUserInfoReq{} }
func (m *SetUserInfoReq) String() string { return proto.CompactTextString(m) }
func (*SetUserInfoReq) ProtoMessage()    {}
func (*SetUserInfoReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{2}
}
func (m *SetUserInfoReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SetUserInfoReq.Unmarshal(m, b)
}
func (m *SetUserInfoReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SetUserInfoReq.Marshal(b, m, deterministic)
}
func (dst *SetUserInfoReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetUserInfoReq.Merge(dst, src)
}
func (m *SetUserInfoReq) XXX_Size() int {
	return xxx_messageInfo_SetUserInfoReq.Size(m)
}
func (m *SetUserInfoReq) XXX_DiscardUnknown() {
	xxx_messageInfo_SetUserInfoReq.DiscardUnknown(m)
}

var xxx_messageInfo_SetUserInfoReq proto.InternalMessageInfo

func (m *SetUserInfoReq) GetUid() int64 {
	if m != nil {
		return m.Uid
	}
	return 0
}

func (m *SetUserInfoReq) GetField() UserField {
	if m != nil {
		return m.Field
	}
	return UserField_NICKNAME
}

func (m *SetUserInfoReq) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type LoginUserReq struct {
	Account              string   `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	PwdHash              string   `protobuf:"bytes,2,opt,name=pwdHash,proto3" json:"pwdHash,omitempty"`
	Key                  string   `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LoginUserReq) Reset()         { *m = LoginUserReq{} }
func (m *LoginUserReq) String() string { return proto.CompactTextString(m) }
func (*LoginUserReq) ProtoMessage()    {}
func (*LoginUserReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{3}
}
func (m *LoginUserReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LoginUserReq.Unmarshal(m, b)
}
func (m *LoginUserReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LoginUserReq.Marshal(b, m, deterministic)
}
func (dst *LoginUserReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LoginUserReq.Merge(dst, src)
}
func (m *LoginUserReq) XXX_Size() int {
	return xxx_messageInfo_LoginUserReq.Size(m)
}
func (m *LoginUserReq) XXX_DiscardUnknown() {
	xxx_messageInfo_LoginUserReq.DiscardUnknown(m)
}

var xxx_messageInfo_LoginUserReq proto.InternalMessageInfo

func (m *LoginUserReq) GetAccount() string {
	if m != nil {
		return m.Account
	}
	return ""
}

func (m *LoginUserReq) GetPwdHash() string {
	if m != nil {
		return m.PwdHash
	}
	return ""
}

func (m *LoginUserReq) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

type LoginUserRes struct {
	Result               ResCode  `protobuf:"varint,1,opt,name=result,proto3,enum=microuser.ResCode" json:"result,omitempty"`
	Uid                  string   `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid,omitempty"`
	Token                string   `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`
	Expire               int64    `protobuf:"varint,4,opt,name=expire,proto3" json:"expire,omitempty"`
	LimitTime            int64    `protobuf:"varint,5,opt,name=limitTime,proto3" json:"limitTime,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LoginUserRes) Reset()         { *m = LoginUserRes{} }
func (m *LoginUserRes) String() string { return proto.CompactTextString(m) }
func (*LoginUserRes) ProtoMessage()    {}
func (*LoginUserRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{4}
}
func (m *LoginUserRes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LoginUserRes.Unmarshal(m, b)
}
func (m *LoginUserRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LoginUserRes.Marshal(b, m, deterministic)
}
func (dst *LoginUserRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LoginUserRes.Merge(dst, src)
}
func (m *LoginUserRes) XXX_Size() int {
	return xxx_messageInfo_LoginUserRes.Size(m)
}
func (m *LoginUserRes) XXX_DiscardUnknown() {
	xxx_messageInfo_LoginUserRes.DiscardUnknown(m)
}

var xxx_messageInfo_LoginUserRes proto.InternalMessageInfo

func (m *LoginUserRes) GetResult() ResCode {
	if m != nil {
		return m.Result
	}
	return ResCode_OK
}

func (m *LoginUserRes) GetUid() string {
	if m != nil {
		return m.Uid
	}
	return ""
}

func (m *LoginUserRes) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *LoginUserRes) GetExpire() int64 {
	if m != nil {
		return m.Expire
	}
	return 0
}

func (m *LoginUserRes) GetLimitTime() int64 {
	if m != nil {
		return m.LimitTime
	}
	return 0
}

type GetLoginInfoReq struct {
	TokenHash            string   `protobuf:"bytes,1,opt,name=tokenHash,proto3" json:"tokenHash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetLoginInfoReq) Reset()         { *m = GetLoginInfoReq{} }
func (m *GetLoginInfoReq) String() string { return proto.CompactTextString(m) }
func (*GetLoginInfoReq) ProtoMessage()    {}
func (*GetLoginInfoReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{5}
}
func (m *GetLoginInfoReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetLoginInfoReq.Unmarshal(m, b)
}
func (m *GetLoginInfoReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetLoginInfoReq.Marshal(b, m, deterministic)
}
func (dst *GetLoginInfoReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetLoginInfoReq.Merge(dst, src)
}
func (m *GetLoginInfoReq) XXX_Size() int {
	return xxx_messageInfo_GetLoginInfoReq.Size(m)
}
func (m *GetLoginInfoReq) XXX_DiscardUnknown() {
	xxx_messageInfo_GetLoginInfoReq.DiscardUnknown(m)
}

var xxx_messageInfo_GetLoginInfoReq proto.InternalMessageInfo

func (m *GetLoginInfoReq) GetTokenHash() string {
	if m != nil {
		return m.TokenHash
	}
	return ""
}

type GetLoginInfoRes struct {
	Result               ResCode  `protobuf:"varint,1,opt,name=result,proto3,enum=microuser.ResCode" json:"result,omitempty"`
	Uid                  string   `protobuf:"bytes,2,opt,name=uid,proto3" json:"uid,omitempty"`
	Key                  string   `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	Token                string   `protobuf:"bytes,4,opt,name=token,proto3" json:"token,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetLoginInfoRes) Reset()         { *m = GetLoginInfoRes{} }
func (m *GetLoginInfoRes) String() string { return proto.CompactTextString(m) }
func (*GetLoginInfoRes) ProtoMessage()    {}
func (*GetLoginInfoRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{6}
}
func (m *GetLoginInfoRes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetLoginInfoRes.Unmarshal(m, b)
}
func (m *GetLoginInfoRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetLoginInfoRes.Marshal(b, m, deterministic)
}
func (dst *GetLoginInfoRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetLoginInfoRes.Merge(dst, src)
}
func (m *GetLoginInfoRes) XXX_Size() int {
	return xxx_messageInfo_GetLoginInfoRes.Size(m)
}
func (m *GetLoginInfoRes) XXX_DiscardUnknown() {
	xxx_messageInfo_GetLoginInfoRes.DiscardUnknown(m)
}

var xxx_messageInfo_GetLoginInfoRes proto.InternalMessageInfo

func (m *GetLoginInfoRes) GetResult() ResCode {
	if m != nil {
		return m.Result
	}
	return ResCode_OK
}

func (m *GetLoginInfoRes) GetUid() string {
	if m != nil {
		return m.Uid
	}
	return ""
}

func (m *GetLoginInfoRes) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *GetLoginInfoRes) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

type GetUserInfoReq struct {
	Uid                  int64     `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Field                UserField `protobuf:"varint,2,opt,name=field,proto3,enum=microuser.UserField" json:"field,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *GetUserInfoReq) Reset()         { *m = GetUserInfoReq{} }
func (m *GetUserInfoReq) String() string { return proto.CompactTextString(m) }
func (*GetUserInfoReq) ProtoMessage()    {}
func (*GetUserInfoReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{7}
}
func (m *GetUserInfoReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetUserInfoReq.Unmarshal(m, b)
}
func (m *GetUserInfoReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetUserInfoReq.Marshal(b, m, deterministic)
}
func (dst *GetUserInfoReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetUserInfoReq.Merge(dst, src)
}
func (m *GetUserInfoReq) XXX_Size() int {
	return xxx_messageInfo_GetUserInfoReq.Size(m)
}
func (m *GetUserInfoReq) XXX_DiscardUnknown() {
	xxx_messageInfo_GetUserInfoReq.DiscardUnknown(m)
}

var xxx_messageInfo_GetUserInfoReq proto.InternalMessageInfo

func (m *GetUserInfoReq) GetUid() int64 {
	if m != nil {
		return m.Uid
	}
	return 0
}

func (m *GetUserInfoReq) GetField() UserField {
	if m != nil {
		return m.Field
	}
	return UserField_NICKNAME
}

type GetUserInfoRes struct {
	Result               ResCode  `protobuf:"varint,1,opt,name=result,proto3,enum=microuser.ResCode" json:"result,omitempty"`
	Value                string   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetUserInfoRes) Reset()         { *m = GetUserInfoRes{} }
func (m *GetUserInfoRes) String() string { return proto.CompactTextString(m) }
func (*GetUserInfoRes) ProtoMessage()    {}
func (*GetUserInfoRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{8}
}
func (m *GetUserInfoRes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetUserInfoRes.Unmarshal(m, b)
}
func (m *GetUserInfoRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetUserInfoRes.Marshal(b, m, deterministic)
}
func (dst *GetUserInfoRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetUserInfoRes.Merge(dst, src)
}
func (m *GetUserInfoRes) XXX_Size() int {
	return xxx_messageInfo_GetUserInfoRes.Size(m)
}
func (m *GetUserInfoRes) XXX_DiscardUnknown() {
	xxx_messageInfo_GetUserInfoRes.DiscardUnknown(m)
}

var xxx_messageInfo_GetUserInfoRes proto.InternalMessageInfo

func (m *GetUserInfoRes) GetResult() ResCode {
	if m != nil {
		return m.Result
	}
	return ResCode_OK
}

func (m *GetUserInfoRes) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type UserIdReq struct {
	Uid                  int64    `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *UserIdReq) Reset()         { *m = UserIdReq{} }
func (m *UserIdReq) String() string { return proto.CompactTextString(m) }
func (*UserIdReq) ProtoMessage()    {}
func (*UserIdReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{9}
}
func (m *UserIdReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UserIdReq.Unmarshal(m, b)
}
func (m *UserIdReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UserIdReq.Marshal(b, m, deterministic)
}
func (dst *UserIdReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UserIdReq.Merge(dst, src)
}
func (m *UserIdReq) XXX_Size() int {
	return xxx_messageInfo_UserIdReq.Size(m)
}
func (m *UserIdReq) XXX_DiscardUnknown() {
	xxx_messageInfo_UserIdReq.DiscardUnknown(m)
}

var xxx_messageInfo_UserIdReq proto.InternalMessageInfo

func (m *UserIdReq) GetUid() int64 {
	if m != nil {
		return m.Uid
	}
	return 0
}

type GetUserAllInfoRes struct {
	Result               ResCode  `protobuf:"varint,1,opt,name=result,proto3,enum=microuser.ResCode" json:"result,omitempty"`
	Nickname             string   `protobuf:"bytes,2,opt,name=nickname,proto3" json:"nickname,omitempty"`
	Country              int64    `protobuf:"varint,3,opt,name=country,proto3" json:"country,omitempty"`
	Phone                string   `protobuf:"bytes,4,opt,name=phone,proto3" json:"phone,omitempty"`
	Status               int64    `protobuf:"varint,5,opt,name=status,proto3" json:"status,omitempty"`
	CreditScore          int64    `protobuf:"varint,6,opt,name=creditScore,proto3" json:"creditScore,omitempty"`
	ActiveDays           int64    `protobuf:"varint,7,opt,name=ActiveDays,proto3" json:"ActiveDays,omitempty"`
	WalletAddress        string   `protobuf:"bytes,8,opt,name=walletAddress,proto3" json:"walletAddress,omitempty"`
	AvatarUrl            string   `protobuf:"bytes,9,opt,name=avatarUrl,proto3" json:"avatarUrl,omitempty"`
	Uid                  int64    `protobuf:"varint,10,opt,name=uid,proto3" json:"uid,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetUserAllInfoRes) Reset()         { *m = GetUserAllInfoRes{} }
func (m *GetUserAllInfoRes) String() string { return proto.CompactTextString(m) }
func (*GetUserAllInfoRes) ProtoMessage()    {}
func (*GetUserAllInfoRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{10}
}
func (m *GetUserAllInfoRes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetUserAllInfoRes.Unmarshal(m, b)
}
func (m *GetUserAllInfoRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetUserAllInfoRes.Marshal(b, m, deterministic)
}
func (dst *GetUserAllInfoRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetUserAllInfoRes.Merge(dst, src)
}
func (m *GetUserAllInfoRes) XXX_Size() int {
	return xxx_messageInfo_GetUserAllInfoRes.Size(m)
}
func (m *GetUserAllInfoRes) XXX_DiscardUnknown() {
	xxx_messageInfo_GetUserAllInfoRes.DiscardUnknown(m)
}

var xxx_messageInfo_GetUserAllInfoRes proto.InternalMessageInfo

func (m *GetUserAllInfoRes) GetResult() ResCode {
	if m != nil {
		return m.Result
	}
	return ResCode_OK
}

func (m *GetUserAllInfoRes) GetNickname() string {
	if m != nil {
		return m.Nickname
	}
	return ""
}

func (m *GetUserAllInfoRes) GetCountry() int64 {
	if m != nil {
		return m.Country
	}
	return 0
}

func (m *GetUserAllInfoRes) GetPhone() string {
	if m != nil {
		return m.Phone
	}
	return ""
}

func (m *GetUserAllInfoRes) GetStatus() int64 {
	if m != nil {
		return m.Status
	}
	return 0
}

func (m *GetUserAllInfoRes) GetCreditScore() int64 {
	if m != nil {
		return m.CreditScore
	}
	return 0
}

func (m *GetUserAllInfoRes) GetActiveDays() int64 {
	if m != nil {
		return m.ActiveDays
	}
	return 0
}

func (m *GetUserAllInfoRes) GetWalletAddress() string {
	if m != nil {
		return m.WalletAddress
	}
	return ""
}

func (m *GetUserAllInfoRes) GetAvatarUrl() string {
	if m != nil {
		return m.AvatarUrl
	}
	return ""
}

func (m *GetUserAllInfoRes) GetUid() int64 {
	if m != nil {
		return m.Uid
	}
	return 0
}

type TokenReq struct {
	TokenHash            string   `protobuf:"bytes,1,opt,name=tokenHash,proto3" json:"tokenHash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TokenReq) Reset()         { *m = TokenReq{} }
func (m *TokenReq) String() string { return proto.CompactTextString(m) }
func (*TokenReq) ProtoMessage()    {}
func (*TokenReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{11}
}
func (m *TokenReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TokenReq.Unmarshal(m, b)
}
func (m *TokenReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TokenReq.Marshal(b, m, deterministic)
}
func (dst *TokenReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TokenReq.Merge(dst, src)
}
func (m *TokenReq) XXX_Size() int {
	return xxx_messageInfo_TokenReq.Size(m)
}
func (m *TokenReq) XXX_DiscardUnknown() {
	xxx_messageInfo_TokenReq.DiscardUnknown(m)
}

var xxx_messageInfo_TokenReq proto.InternalMessageInfo

func (m *TokenReq) GetTokenHash() string {
	if m != nil {
		return m.TokenHash
	}
	return ""
}

type AutoLoginReq struct {
	TokenHash            string   `protobuf:"bytes,1,opt,name=tokenHash,proto3" json:"tokenHash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AutoLoginReq) Reset()         { *m = AutoLoginReq{} }
func (m *AutoLoginReq) String() string { return proto.CompactTextString(m) }
func (*AutoLoginReq) ProtoMessage()    {}
func (*AutoLoginReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{12}
}
func (m *AutoLoginReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AutoLoginReq.Unmarshal(m, b)
}
func (m *AutoLoginReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AutoLoginReq.Marshal(b, m, deterministic)
}
func (dst *AutoLoginReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AutoLoginReq.Merge(dst, src)
}
func (m *AutoLoginReq) XXX_Size() int {
	return xxx_messageInfo_AutoLoginReq.Size(m)
}
func (m *AutoLoginReq) XXX_DiscardUnknown() {
	xxx_messageInfo_AutoLoginReq.DiscardUnknown(m)
}

var xxx_messageInfo_AutoLoginReq proto.InternalMessageInfo

func (m *AutoLoginReq) GetTokenHash() string {
	if m != nil {
		return m.TokenHash
	}
	return ""
}

type AutoLoginRes struct {
	Result               ResCode  `protobuf:"varint,1,opt,name=result,proto3,enum=microuser.ResCode" json:"result,omitempty"`
	Expire               int64    `protobuf:"varint,2,opt,name=expire,proto3" json:"expire,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AutoLoginRes) Reset()         { *m = AutoLoginRes{} }
func (m *AutoLoginRes) String() string { return proto.CompactTextString(m) }
func (*AutoLoginRes) ProtoMessage()    {}
func (*AutoLoginRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{13}
}
func (m *AutoLoginRes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AutoLoginRes.Unmarshal(m, b)
}
func (m *AutoLoginRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AutoLoginRes.Marshal(b, m, deterministic)
}
func (dst *AutoLoginRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AutoLoginRes.Merge(dst, src)
}
func (m *AutoLoginRes) XXX_Size() int {
	return xxx_messageInfo_AutoLoginRes.Size(m)
}
func (m *AutoLoginRes) XXX_DiscardUnknown() {
	xxx_messageInfo_AutoLoginRes.DiscardUnknown(m)
}

var xxx_messageInfo_AutoLoginRes proto.InternalMessageInfo

func (m *AutoLoginRes) GetResult() ResCode {
	if m != nil {
		return m.Result
	}
	return ResCode_OK
}

func (m *AutoLoginRes) GetExpire() int64 {
	if m != nil {
		return m.Expire
	}
	return 0
}

type WalletBindReq struct {
	Uid                  int64    `protobuf:"varint,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Address              string   `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WalletBindReq) Reset()         { *m = WalletBindReq{} }
func (m *WalletBindReq) String() string { return proto.CompactTextString(m) }
func (*WalletBindReq) ProtoMessage()    {}
func (*WalletBindReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{14}
}
func (m *WalletBindReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WalletBindReq.Unmarshal(m, b)
}
func (m *WalletBindReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WalletBindReq.Marshal(b, m, deterministic)
}
func (dst *WalletBindReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WalletBindReq.Merge(dst, src)
}
func (m *WalletBindReq) XXX_Size() int {
	return xxx_messageInfo_WalletBindReq.Size(m)
}
func (m *WalletBindReq) XXX_DiscardUnknown() {
	xxx_messageInfo_WalletBindReq.DiscardUnknown(m)
}

var xxx_messageInfo_WalletBindReq proto.InternalMessageInfo

func (m *WalletBindReq) GetUid() int64 {
	if m != nil {
		return m.Uid
	}
	return 0
}

func (m *WalletBindReq) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

type WalletAddr struct {
	Address              string   `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	CreateTime           int64    `protobuf:"varint,3,opt,name=createTime,proto3" json:"createTime,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WalletAddr) Reset()         { *m = WalletAddr{} }
func (m *WalletAddr) String() string { return proto.CompactTextString(m) }
func (*WalletAddr) ProtoMessage()    {}
func (*WalletAddr) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{15}
}
func (m *WalletAddr) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WalletAddr.Unmarshal(m, b)
}
func (m *WalletAddr) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WalletAddr.Marshal(b, m, deterministic)
}
func (dst *WalletAddr) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WalletAddr.Merge(dst, src)
}
func (m *WalletAddr) XXX_Size() int {
	return xxx_messageInfo_WalletAddr.Size(m)
}
func (m *WalletAddr) XXX_DiscardUnknown() {
	xxx_messageInfo_WalletAddr.DiscardUnknown(m)
}

var xxx_messageInfo_WalletAddr proto.InternalMessageInfo

func (m *WalletAddr) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *WalletAddr) GetCreateTime() int64 {
	if m != nil {
		return m.CreateTime
	}
	return 0
}

type WalletQueryRes struct {
	Result               ResCode       `protobuf:"varint,1,opt,name=result,proto3,enum=microuser.ResCode" json:"result,omitempty"`
	Wallets              []*WalletAddr `protobuf:"bytes,2,rep,name=wallets,proto3" json:"wallets,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *WalletQueryRes) Reset()         { *m = WalletQueryRes{} }
func (m *WalletQueryRes) String() string { return proto.CompactTextString(m) }
func (*WalletQueryRes) ProtoMessage()    {}
func (*WalletQueryRes) Descriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_message_5de22529307f0a38, []int{16}
}
func (m *WalletQueryRes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WalletQueryRes.Unmarshal(m, b)
}
func (m *WalletQueryRes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WalletQueryRes.Marshal(b, m, deterministic)
}
func (dst *WalletQueryRes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WalletQueryRes.Merge(dst, src)
}
func (m *WalletQueryRes) XXX_Size() int {
	return xxx_messageInfo_WalletQueryRes.Size(m)
}
func (m *WalletQueryRes) XXX_DiscardUnknown() {
	xxx_messageInfo_WalletQueryRes.DiscardUnknown(m)
}

var xxx_messageInfo_WalletQueryRes proto.InternalMessageInfo

func (m *WalletQueryRes) GetResult() ResCode {
	if m != nil {
		return m.Result
	}
	return ResCode_OK
}

func (m *WalletQueryRes) GetWallets() []*WalletAddr {
	if m != nil {
		return m.Wallets
	}
	return nil
}

func init() {
	proto.RegisterType((*ResMsg)(nil), "microuser.ResMsg")
	proto.RegisterType((*RegUserInfo)(nil), "microuser.RegUserInfo")
	proto.RegisterType((*SetUserInfoReq)(nil), "microuser.SetUserInfoReq")
	proto.RegisterType((*LoginUserReq)(nil), "microuser.LoginUserReq")
	proto.RegisterType((*LoginUserRes)(nil), "microuser.LoginUserRes")
	proto.RegisterType((*GetLoginInfoReq)(nil), "microuser.GetLoginInfoReq")
	proto.RegisterType((*GetLoginInfoRes)(nil), "microuser.GetLoginInfoRes")
	proto.RegisterType((*GetUserInfoReq)(nil), "microuser.GetUserInfoReq")
	proto.RegisterType((*GetUserInfoRes)(nil), "microuser.GetUserInfoRes")
	proto.RegisterType((*UserIdReq)(nil), "microuser.UserIdReq")
	proto.RegisterType((*GetUserAllInfoRes)(nil), "microuser.GetUserAllInfoRes")
	proto.RegisterType((*TokenReq)(nil), "microuser.TokenReq")
	proto.RegisterType((*AutoLoginReq)(nil), "microuser.AutoLoginReq")
	proto.RegisterType((*AutoLoginRes)(nil), "microuser.AutoLoginRes")
	proto.RegisterType((*WalletBindReq)(nil), "microuser.WalletBindReq")
	proto.RegisterType((*WalletAddr)(nil), "microuser.WalletAddr")
	proto.RegisterType((*WalletQueryRes)(nil), "microuser.WalletQueryRes")
}

func init() {
	proto.RegisterFile("micro-user-message.proto", fileDescriptor_micro_user_message_5de22529307f0a38)
}

var fileDescriptor_micro_user_message_5de22529307f0a38 = []byte{
	// 628 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x55, 0xdb, 0x4e, 0x14, 0x41,
	0x10, 0xcd, 0xee, 0xec, 0x85, 0x29, 0x60, 0xd5, 0x0e, 0x90, 0x09, 0x41, 0x43, 0x3a, 0x3e, 0x6c,
	0x88, 0x40, 0x82, 0x8f, 0x3e, 0xad, 0x1a, 0xd0, 0x44, 0x4d, 0x6c, 0x20, 0x3c, 0xb7, 0x33, 0xc5,
	0xd2, 0x61, 0x2e, 0x6b, 0x77, 0x0f, 0xeb, 0xfe, 0x89, 0xff, 0xe2, 0xcf, 0x99, 0xae, 0xe9, 0xd9,
	0x1d, 0x2e, 0x86, 0xac, 0xfa, 0xd6, 0x55, 0x7d, 0xba, 0x4e, 0x9d, 0xba, 0xcc, 0x40, 0x94, 0xa9,
	0x58, 0x17, 0xfb, 0xa5, 0x41, 0xbd, 0x9f, 0xa1, 0x31, 0x72, 0x8c, 0x07, 0x13, 0x5d, 0xd8, 0x82,
	0x85, 0x74, 0xe3, 0x2e, 0xb6, 0x37, 0x1b, 0x20, 0xcc, 0xcb, 0xac, 0x42, 0xf0, 0x63, 0xe8, 0x09,
	0x34, 0x9f, 0xcd, 0x98, 0xed, 0x41, 0x4f, 0xa3, 0x29, 0x53, 0x1b, 0xb5, 0x76, 0x5b, 0xc3, 0xc1,
	0x11, 0x3b, 0x98, 0x3f, 0x3e, 0x10, 0x68, 0xde, 0x15, 0x09, 0x0a, 0x8f, 0x60, 0x4f, 0x21, 0xc8,
	0xcc, 0x38, 0x6a, 0xef, 0xb6, 0x86, 0xa1, 0x70, 0x47, 0x3e, 0x83, 0x55, 0x81, 0xe3, 0x73, 0x83,
	0xfa, 0x63, 0x7e, 0x59, 0x38, 0xc0, 0x64, 0x9a, 0x50, 0xa4, 0x50, 0xb8, 0x23, 0x8b, 0xa0, 0x1f,
	0x17, 0x65, 0x6e, 0xf5, 0x8c, 0x9e, 0x75, 0x45, 0x6d, 0xb2, 0x0d, 0xe8, 0x4e, 0xae, 0x8a, 0x1c,
	0xa3, 0x80, 0xd0, 0x95, 0xe1, 0xbc, 0x98, 0x49, 0x95, 0x46, 0x9d, 0xca, 0x4b, 0x06, 0x63, 0xd0,
	0xb1, 0xb3, 0x09, 0x46, 0x5d, 0x0a, 0x41, 0x67, 0x9e, 0xc0, 0xe0, 0x14, 0x6d, 0x4d, 0x2d, 0xf0,
	0xbb, 0x63, 0x2f, 0x55, 0xc5, 0x1e, 0x08, 0x77, 0x64, 0x7b, 0xd0, 0xbd, 0x54, 0x98, 0x26, 0xc4,
	0x3d, 0x38, 0xda, 0x68, 0x68, 0x73, 0x0f, 0x8f, 0xdd, 0x9d, 0xa8, 0x20, 0x8e, 0xf9, 0x46, 0xa6,
	0xe5, 0x3c, 0x1f, 0x32, 0xf8, 0x19, 0xac, 0x7d, 0x2a, 0xc6, 0x2a, 0x77, 0x70, 0xc7, 0x11, 0x41,
	0x5f, 0xc6, 0x24, 0xc1, 0xab, 0xac, 0x4d, 0x77, 0x33, 0x99, 0x26, 0x1f, 0xa4, 0xb9, 0xf2, 0x05,
	0xaa, 0x4d, 0x97, 0xd7, 0x35, 0xce, 0x7c, 0x5c, 0x77, 0xe4, 0x3f, 0x5b, 0xb7, 0xc2, 0x9a, 0x65,
	0xbb, 0xe0, 0x64, 0xfa, 0x2e, 0x38, 0x99, 0x1b, 0xd0, 0xb5, 0xc5, 0x35, 0xe6, 0x75, 0xea, 0x64,
	0xb0, 0x2d, 0xe8, 0xe1, 0x8f, 0x89, 0xd2, 0x48, 0xb5, 0x0c, 0x84, 0xb7, 0xd8, 0x0e, 0x84, 0xa9,
	0xca, 0x94, 0x3d, 0x53, 0x59, 0x55, 0xd1, 0x40, 0x2c, 0x1c, 0xfc, 0x10, 0x9e, 0x9c, 0xa0, 0xa5,
	0xe4, 0xea, 0xba, 0xee, 0x40, 0x48, 0x11, 0x49, 0x5b, 0xa5, 0x7a, 0xe1, 0xe0, 0xd3, 0xbb, 0x0f,
	0xfe, 0x55, 0xcd, 0xbd, 0x72, 0x2d, 0xf4, 0x75, 0x1a, 0xfa, 0xf8, 0x17, 0x18, 0x9c, 0xfc, 0xc7,
	0x01, 0xe0, 0xe2, 0x4e, 0xbc, 0xe5, 0x74, 0xcc, 0xc7, 0xa7, 0xdd, 0x1c, 0x9f, 0xe7, 0x10, 0x52,
	0xc0, 0xe4, 0xc1, 0xf4, 0xf8, 0xaf, 0x36, 0x3c, 0xf3, 0x9c, 0xa3, 0x34, 0xfd, 0x1b, 0xda, 0x6d,
	0x58, 0xc9, 0x55, 0x7c, 0x9d, 0xcb, 0xac, 0x66, 0x9e, 0xdb, 0xcd, 0xdd, 0x0b, 0x88, 0xf3, 0xfe,
	0xee, 0x75, 0x9a, 0xbb, 0xb7, 0x05, 0x3d, 0x63, 0xa5, 0x2d, 0x8d, 0x9f, 0x0a, 0x6f, 0xb1, 0x5d,
	0x58, 0x8d, 0x35, 0x26, 0xca, 0x9e, 0xc6, 0x85, 0xc6, 0xa8, 0x47, 0x97, 0x4d, 0x17, 0x7b, 0x01,
	0x30, 0x8a, 0xad, 0xba, 0xc1, 0xf7, 0x72, 0x66, 0xa2, 0x3e, 0x01, 0x1a, 0x1e, 0xf6, 0x12, 0xd6,
	0xa7, 0x32, 0x4d, 0xd1, 0x8e, 0x92, 0x44, 0xa3, 0x31, 0xd1, 0x0a, 0xf1, 0xde, 0x76, 0xba, 0x39,
	0x93, 0x37, 0xd2, 0x4a, 0x7d, 0xae, 0xd3, 0x28, 0xac, 0xe6, 0x6c, 0xee, 0xa8, 0xab, 0x07, 0x8b,
	0xea, 0x0d, 0x61, 0xe5, 0xcc, 0x4d, 0xc2, 0xe3, 0x33, 0xfa, 0x0a, 0xd6, 0x46, 0xa5, 0x2d, 0x68,
	0x48, 0x1f, 0x47, 0x8b, 0x5b, 0xe8, 0xe5, 0xfa, 0xb1, 0x58, 0xba, 0x76, 0x73, 0xe9, 0xf8, 0x1b,
	0x58, 0xbf, 0x20, 0xb1, 0x6f, 0x55, 0xfe, 0xf0, 0x30, 0xd0, 0xa7, 0xc5, 0x97, 0xc7, 0x7f, 0x40,
	0xbc, 0xc9, 0x8f, 0x01, 0x2e, 0xe6, 0x95, 0xfa, 0x33, 0xce, 0xb5, 0x21, 0xd6, 0x28, 0x2d, 0xd2,
	0x6a, 0x57, 0x3d, 0x6f, 0x78, 0x78, 0x06, 0x83, 0x2a, 0xce, 0xd7, 0x12, 0xf5, 0x6c, 0x59, 0x69,
	0x87, 0xd0, 0xaf, 0xfa, 0xe5, 0x78, 0x83, 0xe1, 0xea, 0xd1, 0x66, 0x03, 0xbc, 0xc8, 0x4f, 0xd4,
	0xa8, 0x6f, 0x3d, 0xfa, 0xd7, 0xbc, 0xfe, 0x1d, 0x00, 0x00, 0xff, 0xff, 0xc4, 0x64, 0x62, 0x84,
	0xa9, 0x06, 0x00, 0x00,
}
