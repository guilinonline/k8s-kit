// PodOperator Example
//
// This example demonstrates how to use the PodOperator for log retrieval and command execution.
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/seaman/k8s-kit/pkg/client"
	"github.com/seaman/k8s-kit/pkg/pod"
)

func main() {
	ctx := context.Background()

	// Create client (in production, get from ClusterManager)
	clientFactory := client.NewFactory()
	kubeconfig, err := os.ReadFile("C:\\Users\\chenguilin\\code\\cglk8s-kit\\docs\\config-mock-1")
	if err != nil {
		log.Fatal(err)
	}
	cli, err := clientFactory.CreateFromKubeconfig(kubeconfig)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Create PodOperator
	operator := pod.NewOperator()

	// Example 1: Get simple logs
	fmt.Println("=== Get Logs Simple ===")
	logs, err := operator.GetLogsSimple(ctx, cli, "default", "nginx2-845cb975b5-2xjcl",
		//pod.WithContainer("main"),
		pod.WithTailLines(100),
		pod.WithTimestamps(true),
	)
	if err != nil {
		log.Printf("Failed to get logs: %v", err)
	} else {
		fmt.Println(logs)
	}

	// Example 2: Stream logs
	fmt.Println("=== Stream Logs ===")
	stream, err := operator.GetLogsStream(ctx, cli, "default", "nginx2-845cb975b5-2xjcl",
		//pod.WithContainer("main"),
		pod.WithFollow(true),
	)
	if err != nil {
		log.Printf("Failed to get log stream: %v", err)
	} else {
		defer stream.Close()
		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}

	// Example 3: Simple exec
	fmt.Println("=== Simple Exec ===")
	result, err := operator.ExecSimple(ctx, cli, "default", "nginx2-845cb975b5-2xjcl",
		[]string{"ls", "-la", "/"},
		//pod.WithExecContainer("main"),
		pod.WithExecTimeout(10*time.Second),
	)
	if err != nil {
		log.Printf("Failed to exec: %v", err)
	} else {
		fmt.Printf("Exit code: %d\n", result.ExitCode)
		fmt.Printf("Stdout:\n%s\n", result.Stdout)
		if result.Stderr != "" {
			fmt.Printf("Stderr:\n%s\n", result.Stderr)
		}
	}

	// Example 4: Interactive exec
	fmt.Println("=== Interactive Exec ===")
	session, err := operator.ExecStream(ctx, cli, "default", "nginx2-845cb975b5-2xjcl",
		[]string{"/bin/sh"},
		//pod.WithExecContainer("main"),
		pod.WithTTY(true),
		pod.WithStdin(true),
	)
	if err != nil {
		log.Printf("Failed to create exec session: %v", err)
	} else {
		defer session.Close()

		// Send commands in background
		go func() {
			session.Stdin.Write([]byte("echo 'Hello from k8s-kit'\n"))
			time.Sleep(1 * time.Second)
			session.Stdin.Write([]byte("ls -la\n"))
			time.Sleep(1 * time.Second)
			session.Stdin.Write([]byte("exit\n"))
		}()

		// Read output
		buf := make([]byte, 1024)
		for {
			n, err := session.Stdout.Read(buf)
			if err != nil {
				break
			}
			fmt.Print(string(buf[:n]))
		}

		// Wait for completion
		exitCode, _ := session.Wait()
		fmt.Printf("Exit code: %d\n", exitCode)
	}
}
