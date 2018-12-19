package lib

import "time"

type RawReq struct {
	ID int64
	Req []byte
}

type RawResp struct {
	ID int64
	Resp []byte
	Err  error
	Elapse	time.Duration
} 

type RetCode int

const (
	RET_CODE_SUCCESS	RetCode 	= 0
	RET_CODE_WARNING_CALL_TIMEOUT	= 1001
	RET_CODE_ERROR_CALL				= 2001
	RET_CODE_ERROR_RESPONSE			= 2002
	RET_CODE_ERROR_CALEE			= 2003
	RET_CODE_FATAL_CALL				= 3001
)

type CallResult struct {
	ID 		int64	// ID
	Req 	RawReq	// 原生请求
	Resp 	RawResp	// 原生响应
	Code 	RetCode	// 响应代码
	Msg		string	// 结果成因简述
	Elapse 	time.Duration 	// 耗时
}

const (
	STATUS_ORIGINAL 	uint32 = 0
	STATUS_STARTING		uint32 = 1
	STATUS_STARTED			    = 2
	STATUS_STOPPING				= 3
	STATUS_STOPPED				= 4
)

type Generator interface {
	Start() bool
	Stop() bool
	// 获取状态
	Status() uint32
	// 获取调用计数，每次启动会重置该计数
	CallCount() int64
}

func GetRetCodePlain(code RetCode) string {
	var codePlain string
	switch code {
	case RET_CODE_SUCCESS:
		codePlain = "Success"
	case RET_CODE_WARNING_CALL_TIMEOUT:
		codePlain = "Call Timeout Warning"
	case RET_CODE_ERROR_CALL:
		codePlain = "Call Error"
	case RET_CODE_ERROR_RESPONSE:
		codePlain = "Response Error"
	case RET_CODE_ERROR_CALEE:
		codePlain = "Callee Error"
	case RET_CODE_FATAL_CALL:
		codePlain = "Call Fatal Error"
	default:
		codePlain = "Unknown result code"
	}
	return codePlain
}