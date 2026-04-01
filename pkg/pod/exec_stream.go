package pod

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/seaman/k8s-kit/pkg/client"
)

// ExecStream executes a command in interactive mode.
// Returns an ExecSession that provides stdin, stdout, stderr streams.
func (o *Operator) ExecStream(
	ctx context.Context,
	cli *client.ClusterClient,
	namespace, podName string,
	command []string,
	opts ...ExecOption,
) (*ExecSession, error) {
	options := &ExecOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Apply timeout
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	execReq := cli.Clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	execReq.VersionedParams(&corev1.PodExecOptions{
		Container: options.Container,
		Command:   command,
		Stdin:     options.Stdin,
		Stdout:    true,
		Stderr:    true,
		TTY:       options.TTY,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(cli.RESTConfig, "POST", execReq.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	// Create pipes
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	stderrR, stderrW := io.Pipe()

	session := &ExecSession{
		Stdin:  stdinW,
		Stdout: stdoutR,
		Stderr: stderrR,
		doneCh: make(chan struct{}),
	}

	// Execute in background
	go func() {
		defer close(session.doneCh)

		streamOpts := remotecommand.StreamOptions{
			Stdin:  stdinR,
			Stdout: stdoutW,
			Stderr: stderrW,
			Tty:    options.TTY,
		}

		err := executor.StreamWithContext(ctx, streamOpts)

		session.mu.Lock()
		if err != nil {
			if exitErr, ok := err.(interface{ ExitStatus() int }); ok {
				session.exitCode = exitErr.ExitStatus()
			} else {
				session.exitErr = err
			}
		}
		session.mu.Unlock()

		stdoutW.Close()
		stderrW.Close()
	}()

	return session, nil
}

// Wait waits for the command execution to complete and returns exit code and error.
func (s *ExecSession) Wait() (int, error) {
	<-s.doneCh
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exitCode, s.exitErr
}

// Close closes the session and releases resources.
func (s *ExecSession) Close() error {
	if s.Stdin != nil {
		s.Stdin.Close()
	}
	<-s.doneCh
	return nil
}

// ExecShell opens an interactive shell session in the specified container.
func (o *Operator) ExecShell(
	ctx context.Context,
	cli *client.ClusterClient,
	namespace, podName, container string,
) (*ExecSession, error) {
	return o.ExecStream(ctx, cli, namespace, podName, []string{"/bin/sh"},
		WithExecContainer(container),
		WithTTY(true),
		WithStdin(true),
	)
}