package core

import (
	"context"
	"errors"
	"github.com/DreamerLWJ/fast_pool/config"
	cmap "github.com/orcaman/concurrent-map/v2"
	"sync/atomic"
	"time"
)

type PoolStatusChangeHook interface {
	OnChange(status PoolStatus)
}

// Pool 协程池
type Pool struct {
	queue chan func() // 对于第一个版本来说，使用 channel 来做任务派发

	workers cmap.ConcurrentMap[int64, *Worker] // 工作协程

	queueSize int // 队列大小

	corePoolSize    int64              // 核心协程数
	maximumPoolSize int64              // 最大协程数
	keepAliveTime   time.Duration      // 空闲协程存活时间
	handler         RejectedHandler    // 拒绝策略
	cancelFunc      context.CancelFunc // 用于关闭协程的函数

	maxWorkerId int64

	cancelCtx context.Context
	// hooks
	workerHooks []WorkerStatusChangeHook
	poolHooks   []PoolStatusChangeHook

	workerCount  int64 // 工作协程数，应该通过 hook 来改变
	runningCount int64 // 空闲中的数量，应该通过 hook 来改变
	status       int64 // 协程池状态
}

func NewPool(ctx context.Context, cfg config.PoolConfig, handler RejectedHandler,
	workerHook []WorkerStatusChangeHook, poolHook []PoolStatusChangeHook) (*Pool, error) {
	if handler == nil {
		return nil, errors.New("rejected handler not allow nil")
	}

	cancelCtx, cancelFunc := context.WithCancel(ctx)

	workersMap := cmap.NewWithCustomShardingFunction[int64, *Worker](func(key int64) uint32 {
		return uint32(key)
	})
	pool := Pool{
		queue:           make(chan func(), cfg.QueueSize),
		workers:         workersMap,
		queueSize:       cfg.QueueSize,
		corePoolSize:    cfg.CorePoolSize,
		maximumPoolSize: cfg.MaximumPoolSize,
		keepAliveTime:   cfg.KeepAliveTime,
		handler:         handler,
		workerHooks:     workerHook,
		poolHooks:       poolHook,

		cancelFunc: cancelFunc,
		cancelCtx:  cancelCtx,
	}

	err := pool.init(cancelCtx)
	if err != nil {
		return nil, err
	}
	return &pool, nil
}

// init 初始化
func (p *Pool) init(ctx context.Context) error {
	atomic.StoreInt64(&p.status, int64(PoolStatusInit))
	p.hook(PoolStatusInit)
	err := p.initCoreGoRoutine(ctx)
	if err != nil {
		return err
	}
	return nil
}

// 初始化核心协程池
func (p *Pool) initCoreGoRoutine(ctx context.Context) error {
	var i int64 = 0
	for ; i < p.corePoolSize; i++ {
		p.joinWorker()
	}
	atomic.StoreInt64(&p.status, int64(PoolStatusWorking))
	p.hook(PoolStatusWorking)
	return nil
}

// 加入更多的协程将
func (p *Pool) joinWorker() {
	worker := newPoolWorker(p.cancelCtx, p)
	p.workers.Set(worker.Id, worker)
}

func (p *Pool) removeWorker(workerId int64) {
	p.workers.Remove(workerId)
}

// Go 增加任务
func (p *Pool) Go(f func()) {
	if atomic.LoadInt64(&p.runningCount) < atomic.LoadInt64(&p.workerCount) {
		// 当前有空闲协程，则直接分发任务
		p.queue <- f
	} else {
		// 当前没有空闲协程，判断协程数是否达到上限
		if p.workerCount < p.maximumPoolSize {
			// 没有达到上限时，增加协程
			p.joinWorker()
			p.queue <- f
		} else {
			// 达到上限时，判断工作队列是否已满
			if len(p.queue) == p.queueSize {
				// 拒绝策略
				p.handler.RejectedExecution(f)
			} else {
				// 工作队列未满
				p.queue <- f
			}
		}
	}
}

func (p *Pool) GoWithTimeout(f func(), timeout time.Duration) {
	// TODO
}

// Shutdown 关闭
func (p *Pool) Shutdown() {
	close(p.queue)
	p.cancelFunc()
}

func (p *Pool) Stop() {

}

func (p *Pool) hook(status PoolStatus) {
	if p.poolHooks != nil && len(p.poolHooks) != 0 {
		for _, hook := range p.poolHooks {
			hook.OnChange(status)
		}
	}
}

func (p *Pool) hookWorker(workerId int64, lastStatus WorkerStatus, nextStatus WorkerStatus) {
	p.preHook(workerId, lastStatus, nextStatus)
	if p.workerHooks != nil && len(p.workerHooks) != 0 {
		for _, hook := range p.workerHooks {
			hook.OnChange(workerId, lastStatus, nextStatus)
		}
	}
}

func (p *Pool) preHook(workerId int64, lastStatus WorkerStatus, nextStatus WorkerStatus) {
	if nextStatus == WorkerStatusRunning && lastStatus != WorkerStatusRunning {
		atomic.AddInt64(&p.runningCount, 1)
	} else if lastStatus == WorkerStatusRunning {
		atomic.AddInt64(&p.runningCount, -1)
	}

	switch nextStatus {
	case WorkerStatusReady:
		atomic.AddInt64(&p.workerCount, 1)
	case WorkerStatusTerminated:
		atomic.AddInt64(&p.workerCount, -1)
		p.removeWorker(workerId)
	}
}
