package loadgen

import (
	loadgenlib "goInAction/example.v2/chapter04/loadgen/lib"
	helper "goInAction/example.v2/chapter04/loadgen/testHelper"
	"testing"
	"time"
)

var PrintDetail = false

func TestStart(t *testing.T)  {
	server := helper.NewTCPServer()
	defer server.Close()
	serverAddr := "127.0.0.1:8080"
	t.Logf("StartUp TCP server %s ...\n", serverAddr)
	err := server.Listen(serverAddr)
	if err != nil {
		t.Fatalf("TCP Server start faild at address: %s", serverAddr)
		t.FailNow()
	}

	pset := ParamSet{
		Caller: 	helper.NewTCPComm(serverAddr),
		TimeoutNS: 	50 * time.Millisecond,
		LPS:		uint32(2000),
		DurationNS: 5 * time.Second,
		ResultCh: 	make(chan *loadgenlib.CallResult, 50),
	}
	t.Logf("Initialize load generator (timeoutNS=%v, lps=%d, duration=%v)...", pset.TimeoutNS, pset.LPS, pset.DurationNS)
	gen, err := NewGenerator(pset)
	if err != nil {
		t.Fatalf("Load generator initialization faild: %s", err)
		t.FailNow()
	}
	t.Log("Start load generator...")
	gen.Start()

	countMap := make(map[loadgenlib.RetCode]int)
	for r := range pset.ResultCh {
		countMap[r.Code] = countMap[r.Code] + 1
		if PrintDetail {
			t.Logf("Result id = %d, Code = %d, Msg = %s, Elapse = %d. \n", r.ID, r.Code, r.Msg, r.Elapse)
		}
	}

	var total int
	t.Log("retCode count: ")
	for k, v := range countMap {
		codePlain := loadgenlib.GetRetCodePlain(k)
		t.Logf("Code plain %s, Count %d", codePlain, v)
		total += v
	}
	successCount := countMap[loadgenlib.RET_CODE_SUCCESS]
	t.Logf("total = %d, success count = %d", total, successCount)
	tps := float64(successCount)/ float64(pset.DurationNS/1e9)
	t.Logf("Load Per Second %d, Treatment Per Second %f", pset.LPS, tps)
}

func TestStop(t *testing.T)  {
	server := helper.NewTCPServer()
	defer server.Close()
	serverAddr := "127.0.0.1:8080"
	t.Logf("StartUp TCP server %s ...\n", serverAddr)
	err := server.Listen(serverAddr)
	if err != nil {
		t.Fatalf("TCP Server start faild at address: %s", serverAddr)
		t.FailNow()
	}

	pset := ParamSet{
		Caller: 	helper.NewTCPComm(serverAddr),
		TimeoutNS: 	50 * time.Millisecond,
		LPS:		uint32(2000),
		DurationNS: 5 * time.Second,
		ResultCh: 	make(chan *loadgenlib.CallResult, 50),
	}
	t.Logf("Initialize load generator (timeoutNS=%v, lps=%d, durationNS=%v)...",
		pset.TimeoutNS, pset.LPS, pset.DurationNS)
	gen, err := NewGenerator(pset)
	if err != nil {
		t.Fatalf("Load generator initialization failing: %s.\n",
			err)
		t.FailNow()
	}

	// 开始！
	t.Log("Start load generator...")
	gen.Start()
	timeoutNS := 2 * time.Second
	time.AfterFunc(timeoutNS, func() {
		gen.Stop()
	})

	countMap := make(map[loadgenlib.RetCode]int)
	for r := range pset.ResultCh {
		countMap[r.Code] = countMap[r.Code] + 1
		if PrintDetail {
			t.Logf("Result id = %d, Code = %d, Msg = %s, Elapse = %d. \n", r.ID, r.Code, r.Msg, r.Elapse)
		}
	}

	var total int
	t.Log("retCode count: ")
	for k, v := range countMap {
		codePlain := loadgenlib.GetRetCodePlain(k)
		t.Logf("Code plain %s, Count %d", codePlain, v)
		total += v
	}
	successCount := countMap[loadgenlib.RET_CODE_SUCCESS]
	t.Logf("total = %d, success count = %d", total, successCount)
	tps := float64(successCount)/ float64(pset.DurationNS/1e9)
	t.Logf("Load Per Second %d, Treatment Per Second %f", pset.LPS, tps)

}