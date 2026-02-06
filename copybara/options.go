package copybara

// Option configures the interpreter.
type Option func(*Interpreter)

// WithDryRun enables dry-run mode where no actual changes are made.
func WithDryRun(dryRun bool) Option {
	return func(i *Interpreter) {
		i.dryRun = dryRun
	}
}

// WithWorkdir sets the working directory for file operations.
func WithWorkdir(dir string) Option {
	return func(i *Interpreter) {
		i.workDir = dir
	}
}
