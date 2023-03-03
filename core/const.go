package core

type WorkerStatus int64

const (
	WorkerStatusReady      WorkerStatus = iota + 1 // 就绪
	WorkerStatusRunning                            // 运行中
	WorkerStatusShutdown                           // 关闭状态，不拉取新的任务，执行完任务后终止
	WorkerStatusStop                               // 中断状态，直接终止
	WorkerStatusTidying                            // 空闲状态
	WorkerStatusTerminated                         // 已经终止

	WorkerStatusCreate
)

func (w WorkerStatus) String() string {
	switch w {
	case WorkerStatusReady:
		return "[Ready]"
	case WorkerStatusRunning:
		return "[Running]"
	case WorkerStatusShutdown:
		return "[Shutdown]"
	case WorkerStatusTidying:
		return "[Tidying]"
	case WorkerStatusTerminated:
		return "[Terminated]"
	case WorkerStatusCreate:
		return "[Create]"
	default:
		return "[None]"
	}
}

type PoolStatus int64

func (p PoolStatus) String() string {
	switch p {
	case PoolStatusInit:
		return "[Init]"
	case PoolStatusWorking:
		return "[Working]"
	case PoolStatusShutdown:
		return "[Shutdown]"
	case PoolStatusStop:
		return "[Stop]"
	case PoolStatusTerminated:
		return "[Terminated]"
	default:
		return "[None]"
	}
}

const (
	PoolStatusInit       PoolStatus = iota + 1 // 初始化中
	PoolStatusWorking                          // 运行中
	PoolStatusShutdown                         // 关闭状态，不接受新的任务，执行完已有任务后终止
	PoolStatusStop                             // 中断状态
	PoolStatusTerminated                       // 终止状态
)
