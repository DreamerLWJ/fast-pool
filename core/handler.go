package core

// RejectedHandler rejected handler
type RejectedHandler interface {
	RejectedExecution(task func())
}
