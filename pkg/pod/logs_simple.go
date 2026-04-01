package pod

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/guilinonline/k8s-kit/pkg/client"
)

// GetLogsSimple retrieves Pod logs as a simple string.
// This is suitable for short logs. For large logs or real-time tailing,
// use GetLogsStream instead.
func (o *Operator) GetLogsSimple(
	ctx context.Context,
	cli *client.ClusterClient,
	namespace, podName string,
	opts ...LogOption,
) (string, error) {
	options := &LogOptions{}
	for _, opt := range opts {
		opt(options)
	}

	podLogOpts := &corev1.PodLogOptions{
		Container:  options.Container,
		Previous:   options.Previous,
		Timestamps: options.Timestamps,
		Follow:     false,
	}

	if options.TailLines != nil {
		podLogOpts.TailLines = options.TailLines
	}
	if options.SinceTime != nil {
		t := metav1.NewTime(*options.SinceTime)
		podLogOpts.SinceTime = &t
	}
	if options.SinceSeconds != nil {
		podLogOpts.SinceSeconds = options.SinceSeconds
	}
	if options.LimitBytes != nil {
		podLogOpts.LimitBytes = options.LimitBytes
	}

	req := cli.Clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get log stream: %w", err)
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		return "", fmt.Errorf("failed to read log: %w", err)
	}

	return string(data), nil
}

// GetLogsStream retrieves Pod logs as a stream.
// Returns io.ReadCloser for streaming log retrieval.
// Caller is responsible for closing the stream.
func (o *Operator) GetLogsStream(
	ctx context.Context,
	cli *client.ClusterClient,
	namespace, podName string,
	opts ...LogOption,
) (io.ReadCloser, error) {
	options := &LogOptions{}
	for _, opt := range opts {
		opt(options)
	}

	podLogOpts := &corev1.PodLogOptions{
		Container:  options.Container,
		Previous:   options.Previous,
		Timestamps: options.Timestamps,
		Follow:     options.Follow,
	}

	if options.TailLines != nil {
		podLogOpts.TailLines = options.TailLines
	}
	if options.SinceTime != nil {
		t := metav1.NewTime(*options.SinceTime)
		podLogOpts.SinceTime = &t
	}
	if options.SinceSeconds != nil {
		podLogOpts.SinceSeconds = options.SinceSeconds
	}
	if options.LimitBytes != nil {
		podLogOpts.LimitBytes = options.LimitBytes
	}

	req := cli.Clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	return req.Stream(ctx)
}

// TailLogs provides convenient real-time log tailing with a callback.
// The callback is invoked for each new log line.
func (o *Operator) TailLogs(
	ctx context.Context,
	cli *client.ClusterClient,
	namespace, podName string,
	handleLine func(line string),
	opts ...LogOption,
) error {
	options := &LogOptions{Follow: true}
	for _, opt := range opts {
		opt(options)
	}

	stream, err := o.GetLogsStream(ctx, cli, namespace, podName, func(o *LogOptions) {
		*o = *options
	})
	if err != nil {
		return err
	}
	defer stream.Close()

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			handleLine(string(buf[:n]))
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

// GetPodLogsOptions returns default log options with sensible defaults.
func GetPodLogsOptions() LogOptions {
	return LogOptions{
		TailLines: new(int64),
	}
}
