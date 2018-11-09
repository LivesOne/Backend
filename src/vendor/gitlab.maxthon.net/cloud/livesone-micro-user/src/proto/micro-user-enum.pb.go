// Code generated by protoc-gen-go. DO NOT EDIT.
// source: micro-user-enum.proto

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

type ResCode int32

const (
	ResCode_OK                ResCode = 0
	ResCode_ERR_SYSTEM        ResCode = 1
	ResCode_ERR_PARAM         ResCode = 2
	ResCode_ERR_NOTFOUND      ResCode = 3
	ResCode_ERR_LIMITED       ResCode = 4
	ResCode_ERR_INVALID_TOKEN ResCode = 5
	ResCode_ERR_DUP_DATA      ResCode = 6
	ResCode_ERR_FAILED        ResCode = 7
)

var ResCode_name = map[int32]string{
	0: "OK",
	1: "ERR_SYSTEM",
	2: "ERR_PARAM",
	3: "ERR_NOTFOUND",
	4: "ERR_LIMITED",
	5: "ERR_INVALID_TOKEN",
	6: "ERR_DUP_DATA",
	7: "ERR_FAILED",
}
var ResCode_value = map[string]int32{
	"OK":                0,
	"ERR_SYSTEM":        1,
	"ERR_PARAM":         2,
	"ERR_NOTFOUND":      3,
	"ERR_LIMITED":       4,
	"ERR_INVALID_TOKEN": 5,
	"ERR_DUP_DATA":      6,
	"ERR_FAILED":        7,
}

func (x ResCode) String() string {
	return proto.EnumName(ResCode_name, int32(x))
}
func (ResCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_enum_f065b21cf7e21daf, []int{0}
}

type UserField int32

const (
	UserField_NICKNAME         UserField = 0
	UserField_EMAIL            UserField = 1
	UserField_COUNTRY          UserField = 2
	UserField_PHONE            UserField = 3
	UserField_LOGIN_PASSWORD   UserField = 4
	UserField_PAYMENT_PASSWORD UserField = 5
	UserField_LANGUAGE         UserField = 6
	UserField_REGION           UserField = 7
	UserField_FROM             UserField = 8
	UserField_STATUS           UserField = 9
	UserField_LEVEL            UserField = 10
	UserField_CREDIT_SCORE     UserField = 11
	UserField_ACTIVE_DAYS      UserField = 12
	UserField_WX               UserField = 13
	UserField_TG               UserField = 14
	UserField_AVATAR_URL       UserField = 15
	UserField_REGISTER_TIME    UserField = 16
	UserField_UPDATE_TIME      UserField = 17
	UserField_WALLET_ADDRESS   UserField = 18
)

var UserField_name = map[int32]string{
	0:  "NICKNAME",
	1:  "EMAIL",
	2:  "COUNTRY",
	3:  "PHONE",
	4:  "LOGIN_PASSWORD",
	5:  "PAYMENT_PASSWORD",
	6:  "LANGUAGE",
	7:  "REGION",
	8:  "FROM",
	9:  "STATUS",
	10: "LEVEL",
	11: "CREDIT_SCORE",
	12: "ACTIVE_DAYS",
	13: "WX",
	14: "TG",
	15: "AVATAR_URL",
	16: "REGISTER_TIME",
	17: "UPDATE_TIME",
	18: "WALLET_ADDRESS",
}
var UserField_value = map[string]int32{
	"NICKNAME":         0,
	"EMAIL":            1,
	"COUNTRY":          2,
	"PHONE":            3,
	"LOGIN_PASSWORD":   4,
	"PAYMENT_PASSWORD": 5,
	"LANGUAGE":         6,
	"REGION":           7,
	"FROM":             8,
	"STATUS":           9,
	"LEVEL":            10,
	"CREDIT_SCORE":     11,
	"ACTIVE_DAYS":      12,
	"WX":               13,
	"TG":               14,
	"AVATAR_URL":       15,
	"REGISTER_TIME":    16,
	"UPDATE_TIME":      17,
	"WALLET_ADDRESS":   18,
}

func (x UserField) String() string {
	return proto.EnumName(UserField_name, int32(x))
}
func (UserField) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_enum_f065b21cf7e21daf, []int{1}
}

type PwdCheckType int32

const (
	PwdCheckType_LOGIN_PWD   PwdCheckType = 0
	PwdCheckType_PAYMENT_PWD PwdCheckType = 1
)

var PwdCheckType_name = map[int32]string{
	0: "LOGIN_PWD",
	1: "PAYMENT_PWD",
}
var PwdCheckType_value = map[string]int32{
	"LOGIN_PWD":   0,
	"PAYMENT_PWD": 1,
}

func (x PwdCheckType) String() string {
	return proto.EnumName(PwdCheckType_name, int32(x))
}
func (PwdCheckType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_micro_user_enum_f065b21cf7e21daf, []int{2}
}

func init() {
	proto.RegisterEnum("microuser.ResCode", ResCode_name, ResCode_value)
	proto.RegisterEnum("microuser.UserField", UserField_name, UserField_value)
	proto.RegisterEnum("microuser.PwdCheckType", PwdCheckType_name, PwdCheckType_value)
}

func init() {
	proto.RegisterFile("micro-user-enum.proto", fileDescriptor_micro_user_enum_f065b21cf7e21daf)
}

var fileDescriptor_micro_user_enum_f065b21cf7e21daf = []byte{
	// 395 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x44, 0x91, 0x5f, 0x6e, 0x13, 0x31,
	0x10, 0xc6, 0x93, 0xb4, 0xd9, 0x64, 0x27, 0x7f, 0x3a, 0xb1, 0xe8, 0x25, 0x22, 0xb5, 0x2f, 0x9c,
	0x60, 0xb4, 0x9e, 0x2c, 0x56, 0xbc, 0xf6, 0xca, 0xf6, 0x66, 0xc9, 0xd3, 0x4a, 0x34, 0x2b, 0x51,
	0x41, 0x49, 0x95, 0x50, 0x21, 0xee, 0xc0, 0x11, 0x38, 0x2c, 0x9a, 0x42, 0xe8, 0x93, 0x35, 0x3f,
	0xc9, 0xdf, 0xe7, 0x9f, 0x07, 0x6e, 0x9f, 0x1e, 0x1f, 0x4e, 0xc7, 0xbb, 0x97, 0x73, 0x7f, 0xba,
	0xeb, 0xbf, 0xbd, 0x3c, 0xdd, 0x3f, 0x9f, 0x8e, 0xdf, 0x8f, 0x2a, 0x7f, 0xc5, 0x42, 0xd7, 0xbf,
	0x86, 0x30, 0x09, 0xfd, 0xb9, 0x38, 0x1e, 0x7a, 0x95, 0xc1, 0xc8, 0x6f, 0x71, 0xa0, 0x96, 0x00,
	0x1c, 0x42, 0x17, 0xf7, 0x31, 0x71, 0x85, 0x43, 0xb5, 0x80, 0x5c, 0xe6, 0x9a, 0x02, 0x55, 0x38,
	0x52, 0x08, 0x73, 0x19, 0x9d, 0x4f, 0x1b, 0xdf, 0x38, 0x8d, 0x57, 0xea, 0x06, 0x66, 0x42, 0xac,
	0xa9, 0x4c, 0x62, 0x8d, 0xd7, 0xea, 0x16, 0x56, 0x02, 0x8c, 0xdb, 0x91, 0x35, 0xba, 0x4b, 0x7e,
	0xcb, 0x0e, 0xc7, 0x97, 0x9b, 0xba, 0xa9, 0x3b, 0x4d, 0x89, 0x30, 0xbb, 0x54, 0x6d, 0xc8, 0x58,
	0xd6, 0x38, 0x59, 0xff, 0x1e, 0x41, 0xde, 0x9c, 0xfb, 0xd3, 0xe6, 0xb1, 0xff, 0x7a, 0x50, 0x73,
	0x98, 0x3a, 0x53, 0x6c, 0x1d, 0x55, 0x8c, 0x03, 0x95, 0xc3, 0x98, 0x2b, 0x32, 0x16, 0x87, 0x6a,
	0x06, 0x93, 0xc2, 0x37, 0x2e, 0x85, 0x3d, 0x8e, 0x84, 0xd7, 0x1f, 0xbc, 0x63, 0xbc, 0x52, 0x0a,
	0x96, 0xd6, 0x97, 0xc6, 0x75, 0x35, 0xc5, 0xd8, 0xfa, 0x20, 0x6f, 0x79, 0x07, 0x58, 0xd3, 0xbe,
	0x62, 0x97, 0xde, 0xe8, 0x58, 0xa2, 0x2d, 0xb9, 0xb2, 0xa1, 0x92, 0x31, 0x53, 0x00, 0x59, 0xe0,
	0xd2, 0x78, 0x87, 0x13, 0x35, 0x85, 0xeb, 0x4d, 0xf0, 0x15, 0x4e, 0x85, 0xc6, 0x44, 0xa9, 0x89,
	0x98, 0x4b, 0x89, 0xe5, 0x1d, 0x5b, 0x04, 0xb1, 0x28, 0x02, 0x6b, 0x93, 0xba, 0x58, 0xf8, 0xc0,
	0x38, 0x13, 0x7f, 0x2a, 0x92, 0xd9, 0x71, 0xa7, 0x69, 0x1f, 0x71, 0x2e, 0x3f, 0xd9, 0x7e, 0xc4,
	0x85, 0x9c, 0xa9, 0xc4, 0xa5, 0x68, 0xd2, 0x8e, 0x12, 0x85, 0xae, 0x09, 0x16, 0x6f, 0xd4, 0x0a,
	0x16, 0xd2, 0x17, 0x13, 0x87, 0x2e, 0x99, 0x8a, 0x11, 0x25, 0xa3, 0xa9, 0x35, 0x25, 0xfe, 0x0b,
	0x56, 0xe2, 0xd2, 0x92, 0xb5, 0x9c, 0x3a, 0xd2, 0x3a, 0x70, 0x8c, 0xa8, 0xd6, 0xf7, 0x30, 0xaf,
	0x7f, 0x1c, 0x8a, 0xcf, 0xfd, 0xc3, 0x97, 0xf4, 0xf3, 0xb9, 0x97, 0xcd, 0xfc, 0xf3, 0x6d, 0x35,
	0x0e, 0x24, 0xe3, 0xbf, 0x6a, 0xab, 0x71, 0xf8, 0x29, 0x7b, 0xdd, 0xf7, 0xfb, 0x3f, 0x01, 0x00,
	0x00, 0xff, 0xff, 0x27, 0x32, 0xfc, 0xa8, 0x08, 0x02, 0x00, 0x00,
}
