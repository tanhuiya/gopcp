package loadgen

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"goInAction/example.v2/chapter04/loadgen/lib"
	"strings"
	"time"
)

type ParamSet struct {
	Caller  lib.Caller					// 调用器
	TimeoutNS	time.Duration			// 响应超时时间，ns
	LPS 		uint32					// 每秒载荷量
	DurationNS	time.Duration			// 负载持续时间,ns
	ResultCh	chan *lib.CallResult	// 调用结果通道
}

func (pset *ParamSet) Check() error {
	var errMsg []string

	if pset.Caller == nil {
		errMsg = append(errMsg, "Invalid caller")
	}
	if pset.TimeoutNS == 0 {
		errMsg = append(errMsg, "Invalid timeoutNS")
	}
	if pset.LPS == 0 {
		errMsg = append(errMsg, "Invalid LPS")
	}
	if pset.DurationNS == 0 {
		errMsg = append(errMsg, "Invalid durationNS!")
	}
	if pset.ResultCh == nil {
		errMsg = append(errMsg, "Invalid result channel!")
	}
	var buf bytes.Buffer
	buf.WriteString("Checking the parameters...")
	if errMsg != nil {
		errMsgStr := strings.Join(errMsg, " ")
		buf.WriteString(fmt.Sprintf("Not Passed! (%s)", errMsgStr))
		logger.Infoln(buf.String())
		return errors.New(errMsgStr)
	}
	buf.WriteString(fmt.Sprintf("Passed.(timeoutNS=%s, lps=%d, durationNS=%s)", pset.TimeoutNS, pset.LPS, pset.DurationNS) )
	logger.Infoln(buf.String())
	return nil
}