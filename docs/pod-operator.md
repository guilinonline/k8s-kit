# PodOperator API Reference

## Overview

PodOperator provides Pod-specific operations including log retrieval and command execution.

## Quick Start

```go
import "github.com/guilinonline/k8s-kit/pkg/pod"

operator := pod.NewOperator()

// Get logs
logs, err := operator.GetLogsSimple(ctx, cli, "default", "my-pod",
    pod.WithContainer("main"),
    pod.WithTailLines(100),
)

// Execute command
result, err := operator.ExecSimple(ctx, cli, "default", "my-pod",
    []string{"ls", "-la"},
    pod.WithExecContainer("main"),
)
```

## Log Retrieval

### GetLogsSimple

```go
func (o *Operator) GetLogsSimple(ctx, cli, namespace, podName string, opts ...LogOption) (string, error)
```

Returns logs as a string. Suitable for short logs.

Options:
- `WithContainer(name)` - Container name
- `WithTailLines(n)` - Number of lines from end
- `WithTimestamps(bool)` - Include timestamps
- `WithSinceTime(time.Time)` - Logs after this time
- `WithLimitBytes(n)` - Maximum bytes to return

### GetLogsStream

```go
func (o *Operator) GetLogsStream(ctx, cli, namespace, podName string, opts ...LogOption) (io.ReadCloser, error)
```

Returns logs as a stream. Suitable for large logs or real-time tailing.

### TailLogs

```go
func (o *Operator) TailLogs(ctx, cli, namespace, podName string, handleLine func(line string), opts ...LogOption) error
```

Convenient callback-based log tailing.

## Command Execution

### ExecSimple

```go
func (o *Operator) ExecSimple(ctx, cli, namespace, podName string, command []string, opts ...ExecOption) (*ExecResult, error)
```

Executes a command and returns the result.

Options:
- `WithExecContainer(name)` - Container name
- `WithExecTimeout(duration)` - Execution timeout

### ExecStream

```go
func (o *Operator) ExecStream(ctx, cli, namespace, podName string, command []string, opts ...ExecOption) (*ExecSession, error)
```

Interactive execution with I/O streams.

## Error Handling

```go
import "github.com/guilinonline/k8s-kit/pkg/pod"

if pod.IsContainerNotFound(err) { ... }
if pod.IsPodNotFound(err) { ... }
if pod.IsForbidden(err) { ... }
if pod.IsTimeout(err) { ... }
if pod.IsConnectionLost(err) { ... }
```