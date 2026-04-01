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
//
// Example usage for simple logs:
//
//	operator := pod.NewOperator()
//	logs, err := operator.GetLogsSimple(ctx, cli, "default", "my-pod",
//	    pod.WithContainer("main"),
//	    pod.WithTailLines(100),
//	)
//
// Example usage for streaming logs:
//
//	operator := pod.NewOperator()
//	stream, err := operator.GetLogsStream(ctx, cli, "default", "my-pod",
//	    pod.WithContainer("main"),
//	    pod.WithFollow(true),
//	)
//	defer stream.Close()
//	scanner := bufio.NewScanner(stream)
//	for scanner.Scan() {
//	    fmt.Println(scanner.Text())
//	}
//
// Example usage for simple exec:
//
//	operator := pod.NewOperator()
//	result, err := operator.ExecSimple(ctx, cli, "default", "my-pod",
//	    []string{"ls", "-la"},
//	    pod.WithExecContainer("main"),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Exit code: %d\n", result.ExitCode)
//	fmt.Printf("Stdout: %s\n", result.Stdout)
//
// For more examples, see the examples/ directory.
package pod
