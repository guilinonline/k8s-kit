package pod

import (
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
)

// LogOptions contains options for log retrieval.
type LogOptions struct {
	Container    string
	Follow       bool
	Previous     bool
	TailLines    *int64
	SinceTime    *time.Time
	SinceSeconds *int64
	Timestamps   bool
	LimitBytes   *int64
	Namespace    string
	LabelSelector labels.Selector
	FieldSelector fields.Selector
}

// LogOption is the functional option type for log options.
type LogOption func(*LogOptions)

// WithContainer sets the container name.
func WithContainer(name string) LogOption {
	return func(o *LogOptions) {
		o.Container = name
	}
}

// WithFollow sets whether to follow log output like tail -f.
func WithFollow(follow bool) LogOption {
	return func(o *LogOptions) {
		o.Follow = follow
	}
}

// WithPrevious sets whether to return the previous container's logs.
func WithPrevious(previous bool) LogOption {
	return func(o *LogOptions) {
		o.Previous = previous
	}
}

// WithTailLines sets the number of lines from the end to show.
func WithTailLines(lines int64) LogOption {
	return func(o *LogOptions) {
		o.TailLines = &lines
	}
}

// WithSinceTime sets the time after which to return logs.
func WithSinceTime(t time.Time) LogOption {
	return func(o *LogOptions) {
		o.SinceTime = &t
	}
}

// WithSinceSeconds sets the duration in seconds to go back.
func WithSinceSeconds(seconds int64) LogOption {
	return func(o *LogOptions) {
		o.SinceSeconds = &seconds
	}
}

// WithTimestamps sets whether to include timestamps in log output.
func WithTimestamps(timestamps bool) LogOption {
	return func(o *LogOptions) {
		o.Timestamps = timestamps
	}
}

// WithLimitBytes sets the maximum bytes of logs to return.
func WithLimitBytes(bytes int64) LogOption {
	return func(o *LogOptions) {
		o.LimitBytes = &bytes
	}
}

// WithLogNamespace sets the namespace for log listing.
func WithLogNamespace(namespace string) LogOption {
	return func(o *LogOptions) {
		o.Namespace = namespace
	}
}

// WithLogLabelSelector sets the label selector for log listing.
func WithLogLabelSelector(selector labels.Selector) LogOption {
	return func(o *LogOptions) {
		o.LabelSelector = selector
	}
}

// WithLogFieldSelector sets the field selector for log listing.
func WithLogFieldSelector(selector fields.Selector) LogOption {
	return func(o *LogOptions) {
		o.FieldSelector = selector
	}
}

// ExecOptions contains options for command execution.
type ExecOptions struct {
	Container string
	TTY       bool
	Stdin     bool
	Timeout   time.Duration
}

// ExecOption is the functional option type for exec options.
type ExecOption func(*ExecOptions)

// WithExecContainer sets the container name for exec.
func WithExecContainer(name string) ExecOption {
	return func(o *ExecOptions) {
		o.Container = name
	}
}

// WithTTY sets whether to allocate a TTY for the session.
func WithTTY(tty bool) ExecOption {
	return func(o *ExecOptions) {
		o.TTY = tty
	}
}

// WithStdin sets whether to attach stdin for input.
func WithStdin(stdin bool) ExecOption {
	return func(o *ExecOptions) {
		o.Stdin = stdin
	}
}

// WithExecTimeout sets the execution timeout.
func WithExecTimeout(timeout time.Duration) ExecOption {
	return func(o *ExecOptions) {
		o.Timeout = timeout
	}
}
