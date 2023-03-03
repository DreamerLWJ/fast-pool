package core

import (
	"context"
	"sync/atomic"
	"time"
)

type WorkerStatusChangeHook interface {
	OnChange(workerId int64, lastStatus WorkerStatus, nextStatus WorkerStatus)
}

// Worker 执行者
type Worker struct {
	Id            int64 // 唯一 id
	parent        *Pool
	status        atomic.Value // 状态
	isTerminated  bool         // 是否终止
	FreeBeginTime time.Time    // 空闲开始时间
}

func newPoolWorker(ctx context.Context, pool *Pool) *Worker {
	pool.maxWorkerId++
	worker := Worker{
		parent: pool,
		Id:     pool.maxWorkerId,
	}
	worker.start(ctx)
	return &worker
}

func (p *Worker) start(ctx context.Context) {
	p.status.Store(WorkerStatusReady)
	p.parent.hookWorker(p.Id, WorkerStatusCreate, WorkerStatusReady)
	p.FreeBeginTime = time.Now()
	go func() {
	loop:
		for {
			lastStatus := p.status.Load().(WorkerStatus)
			select {
			case <-ctx.Done():
				// Pool 关闭
				p.parent.hookWorker(p.Id, lastStatus, WorkerStatusShutdown)
				p.status.Store(WorkerStatusShutdown)
				break loop
			case task := <-p.parent.queue:
				if task == nil {
					// skip nil task
					break
				}
				p.parent.hookWorker(p.Id, lastStatus, WorkerStatusRunning)
				p.status.Store(WorkerStatusRunning)
				task()
			default:
				p.parent.hookWorker(p.Id, lastStatus, WorkerStatusTidying)
				p.status.Store(WorkerStatusTidying)
				if p.ShouldTerminal() {
					// 空闲协程回收
					p.parent.hookWorker(p.Id, lastStatus, WorkerStatusShutdown)
					p.status.Store(WorkerStatusShutdown)
					break loop
				}
			}
		}
		lastStatus := p.status.Load().(WorkerStatus)
		p.parent.hookWorker(p.Id, lastStatus, WorkerStatusTerminated)
		p.status.Store(WorkerStatusTerminated)
	}()
}

// ShouldTerminal check over core size
func (p *Worker) ShouldTerminal() bool {
	if p.isTerminated {
		return true
	}

	dur := time.Now().Sub(p.FreeBeginTime)
	if dur > p.parent.keepAliveTime {
		return true
	}
	return false
}

func (p *Worker) ResetFreeTime() {
	p.FreeBeginTime = time.Now()
}
