package pod

// ExecResult contains the result of a simple command execution.
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}
