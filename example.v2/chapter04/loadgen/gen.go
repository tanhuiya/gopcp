package loadgen

import (
	"bytes"
	"context"
	"errors"
	"gopcp.v2/helper/log"
	"fmt"
	"goInAction/example.v2/chapter04/loadgen/lib"
	"math"
	"sync/atomic"
	"time"
)

var logger = log.DLogger()

type myGenerator struct {
	caller		lib.Caller
	timeoutNS 	time.Duration
	lps 		uint32
	durationNS 	time.Duration
	concurrency uint32
	tickets		lib.GoTickets
	ctx 		context.Context
	cancelFunc 	context.CancelFunc
	callCount 	int64
	status 		uint32
	resultCh 	chan *lib.CallResult
}

func NewGenerator(pset ParamSet) (lib.Generator, error) {
	logger.Infoln("New a load generator...")
	if err := pset.Check(); err != nil {
		return nil, err
	}
	gen := &myGenerator{
		caller: 	pset.Caller,
		timeoutNS: 	pset.TimeoutNS,
		lps:		pset.LPS,
		durationNS: pset.DurationNS,
		status: 	lib.STATUS_ORIGINAL,
		resultCh: 	pset.ResultCh,
	}
	if err := gen.init(); err != nil {
		return nil, err
	}
	return gen, nil
}

func (gen *myGenerator)init() error {
	var buf bytes.Buffer
	buf.WriteString("Initializing the load generator ...")

	var total64 = int64(gen.timeoutNS)/int64(1e9/gen.lps) + 1
	if total64 > math.MaxInt32 {
		total64 = math.MaxInt32
	}
	gen.concurrency = uint32(total64)
	tickets, err := lib.NewGoTicket(gen.concurrency)
	if err != nil {
		return err
	}
	gen.tickets = tickets
	buf.WriteString(fmt.Sprintf("Done, (concurrency=%d)", gen.concurrency))
	logger.Infoln(buf.String())
	return nil
}

func (gen *myGenerator)genLoad(throttle <-chan time.Time)  {
	for  {
		select {
		case <-gen.ctx.Done():
			logger.Infoln("prepare to stop")
			gen.prepareToStop(gen.ctx.Err())
			return
		default:
		}
		gen.asyncCall()
		if gen.lps > 0 {
			select {
			case <-throttle:
			case <-gen.ctx.Done():
				logger.Infoln("prepare to stop")
				gen.prepareToStop(gen.ctx.Err())
				return
			}
		}
	}
}

func (gen *myGenerator)Start() bool {
	logger.Infoln("Starting load a generator...")
	if !atomic.CompareAndSwapUint32(&gen.status, lib.STATUS_ORIGINAL, lib.STATUS_STARTING) {
		if !atomic.CompareAndSwapUint32(&gen.status, lib.STATUS_STOPPED, lib.STATUS_STARTING) {
			return false
		}
	}

	// 设定节流阀
	var throttle <-chan time.Time
	if gen.lps > 0 {
		interval := time.Duration(1e9 / gen.lps)
		logger.Infoln("Setting throttle (%d)", interval)
		throttle = time.Tick(interval)
	}
	// 初始化上下文和取消函数
	gen.ctx, gen.cancelFunc = context.WithTimeout(context.Background(), gen.durationNS)
	gen.callCount = 0
	atomic.StoreUint32(&gen.status, lib.STATUS_STARTED)
	go func() {
		// 生成并发送载荷
		logger.Infoln("Generating loads...")
		gen.genLoad(throttle)
		logger.Infoln("Stoped. call count: ", gen.callCount)
	}()
	return true
}
func (gen *myGenerator)Stop() bool {
	if !atomic.CompareAndSwapUint32(&gen.status, lib.STATUS_STARTED, lib.STATUS_STOPPING) {
		return false
	}
	gen.cancelFunc()
	for {
		if atomic.LoadUint32(&gen.status) == lib.STATUS_STOPPED {
			break
		}
		time.Sleep(time.Millisecond)
	}
	return true
}
// 获取状态
func (gen *myGenerator)Status() uint32 {
	return atomic.LoadUint32(&gen.status)
}
// 获取调用计数，每次启动会重置该计数
func (gen *myGenerator)CallCount() int64 {
	return atomic.LoadInt64(&gen.callCount)
}

func (gen *myGenerator)prepareToStop(err error) {
	logger.Infoln("Prepare to stop load generate , cause: %s", err)
	atomic.CompareAndSwapUint32(&gen.status, lib.STATUS_STARTED, lib.STATUS_STOPPING)
	logger.Infoln("Closing result channel")
	close(gen.resultCh)
	atomic.CompareAndSwapUint32(&gen.status, lib.STATUS_STOPPING, lib.STATUS_STOPPED)
}

// asyncSend 会异步地调用承受方接口
func (gen *myGenerator)asyncCall () {
	gen.tickets.Take()
	go func() {
		defer func() {
			if p := recover(); p != nil {
				err, ok := interface{}(p).(error)
				var errMsg string
				if ok {
					errMsg = fmt.Sprintf("Async Call Panic : error : %s", err)
				} else {
					errMsg = fmt.Sprintf("Async Call Panic : clue : %#v", p)
				}
				logger.Errorln(errMsg)
				result := &lib.CallResult{
					ID:		-1,
					Msg:	errMsg,
					Code:  	lib.RET_CODE_FATAL_CALL,
				}
				gen.sendResult(result)
			}
			gen.tickets.Return()
		}()
		rawReq := gen.caller.BuildReq()
		// 调用状态： 0-未调用， 1-调用完成， 2-调用超时
		var callStatus uint32
		timer := time.AfterFunc(gen.timeoutNS, func() {
			if !atomic.CompareAndSwapUint32(&callStatus, 0, 2) {
				return
			}
			result := &lib.CallResult{
				ID: rawReq.ID,
				Req:rawReq,
				Code:lib.RET_CODE_WARNING_CALL_TIMEOUT,
				Msg: fmt.Sprintf("Timeout , expect < (%v)", gen.timeoutNS),
				Elapse: gen.timeoutNS,
			}
			gen.sendResult(result)
		})
		rawResp := gen.callOne(&rawReq)
		if !atomic.CompareAndSwapUint32(&callStatus, 0, 1) {
			return
		}
		timer.Stop()
		var result *lib.CallResult
		if rawResp.Err != nil {
			result = &lib.CallResult{
				ID:	rawResp.ID,
				Req:rawReq,
				Code:lib.RET_CODE_ERROR_CALL,
				Elapse:rawResp.Elapse,
				Msg:rawResp.Err.Error(),
			}
		} else {
			result = gen.caller.CheckResp(rawReq, *rawResp)
			result.Elapse = rawResp.Elapse
		}
		gen.sendResult(result)
	}()
}

func (gen *myGenerator)sendResult(result *lib.CallResult) bool {
	if atomic.LoadUint32(&gen.status) != lib.STATUS_STARTED {
		gen.printIgnoreResult(result, "stopped load generator")
		return false
	}
	select {
	case gen.resultCh <- result:
		return true
	default:
		gen.printIgnoreResult(result, "full result channel")
		return false
	}
}

func (gen *myGenerator)printIgnoreResult(result *lib.CallResult, cause string)  {
	resultMsg := fmt.Sprintf("ID=%d, Code=%d, Msg=%s, Elapse=%d", result.ID, result.Code, result.Msg, result.Elapse)
	logger.Warnf("ignore result %s, cause: %s", resultMsg, cause)
}

// 向载荷方发起一次调用
func (gen *myGenerator)callOne(rawReq *lib.RawReq) *lib.RawResp {
	atomic.StoreInt64(&gen.callCount, gen.callCount + 1)
	if rawReq == nil {
		return &lib.RawResp{
			ID: -1,
			Err:errors.New("Invalid raw request"),
		}
	}
	start := time.Now().UnixNano()
	resp, err := gen.caller.Call(rawReq.Req, gen.timeoutNS)
	end := time.Now().UnixNano()
	elapsed := time.Duration(end - start)
	var rawResp lib.RawResp
	if err != nil {
		errMsg := fmt.Sprintf("Sync Call Error, %s", err)
		rawResp = lib.RawResp{
			ID: rawReq.ID,
			Elapse:elapsed,
			Err: errors.New(errMsg),
		}
	} else {
		rawResp = lib.RawResp{
			ID: rawReq.ID,
			Resp:resp,
			Elapse:elapsed,
			Err: nil,
		}
	}
	return &rawResp
}