package test

import (
	"context"
	"fmt"
	"github.com/DreamerLWJ/fast-pool/config"
	"github.com/DreamerLWJ/fast-pool/core"
	"testing"
	"time"
)

type TestHandler struct {
}

func (t TestHandler) RejectedExecution(task func()) {
	fmt.Println("TestHandler|RejectedExecution 触发拒绝策略")
}

type TestWorkerHook struct {
}

func (t TestWorkerHook) OnChange(workerId int64, lastStatus core.WorkerStatus, nextStatus core.WorkerStatus) {
	if lastStatus != nextStatus {
		fmt.Printf("TestWorkerHook|OnChange workerId: %d, %s --> %s\n", workerId, lastStatus, nextStatus)
	}
}

type TestPoolHook struct {
}

func (t TestPoolHook) OnChange(status core.PoolStatus) {
	fmt.Printf("TestPoolHook|OnChange status: %s\n", status)
}

func TestNewPool(t *testing.T) {
	ctx := context.Background()
	pool, err := core.NewPool(ctx, config.PoolConfig{
		CorePoolSize:    2,
		MaximumPoolSize: 20,
		KeepAliveTime:   time.Second * 5,
		QueueSize:       5,
	}, TestHandler{}, []core.WorkerStatusChangeHook{TestWorkerHook{}}, []core.PoolStatusChangeHook{TestPoolHook{}})
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second)
	for i := 0; i < 30; i++ {
		pool.Go(func() {
			time.Sleep(time.Second * 5)
			fmt.Println("pool.Go|coroutine has been finish")
		})
	}

	fmt.Println("TestNewPool| all work has been dispatch")
	time.Sleep(time.Second * 2)
	pool.Shutdown()
	fmt.Println("TestNewPool| pool been shutdown")
}
