// Package pod provides Pod-specific operations including log retrieval
// and command execution inside containers.
//
// This package provides two levels of abstraction:
//
// Simple API:
//   - GetLogsSimple - retrieves logs as a string
//   - ExecSimple - executes a command and returns the result
//
// Streaming API:
//   - GetLogsStream - retrieves logs as an io.ReadCloser
//   - TailLogs - convenient callback-based log tailing
//   - ExecStream - creates an interactive ExecSession
package pod

import (
	"io"
	"sync"
)

// Operator is the Pod operations operator.
type Operator struct {
}

// NewOperator creates a new Pod operator.
func NewOperator() *Operator {
	return &Operator{}
}

// ExecSession represents an interactive command execution session.
// Use ExecStream to create a session.
type ExecSession struct {
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser

	// Internal fields
	doneCh   chan struct{}
	exitCode int
	exitErr  error
	mu       sync.Mutex
}