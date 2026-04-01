package pod

import (
	"bytes"
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/guilinonline/k8s-kit/pkg/client"
)

// ExecSimple executes a command in a container and returns the result.
// This is suitable for simple commands that complete quickly.
// For interactive sessions or long-running commands, use ExecStream.
func (o *Operator) ExecSimple(
	ctx context.Context,
	cli *client.ClusterClient,
	namespace, podName string,
	command []string,
	opts ...ExecOption,
) (*ExecResult, error) {
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
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(cli.RESTConfig, "POST", execReq.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdoutBuf,
		Stderr: &stderrBuf,
	})

	result := &ExecResult{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: 0,
	}

	if err != nil {
		if exitErr, ok := err.(interface{ ExitStatus() int }); ok {
			result.ExitCode = exitErr.ExitStatus()
			return result, nil
		}
		return nil, fmt.Errorf("failed to execute command: %w", err)
	}

	return result, nil
}

// ExecInContainer is a convenience function that executes a command
// and returns stdout, stderr, and error.
func (o *Operator) ExecInContainer(
	ctx context.Context,
	cli *client.ClusterClient,
	namespace, podName, container string,
	cmd []string,
) (stdout, stderr string, exitCode int, err error) {
	result, err := o.ExecSimple(ctx, cli, namespace, podName, cmd,
		WithExecContainer(container))
	if err != nil {
		return "", "", -1, err
	}
	return result.Stdout, result.Stderr, result.ExitCode, nil
}
