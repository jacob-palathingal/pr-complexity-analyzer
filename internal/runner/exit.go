package runner

// CommandExitError lets Cobra return a controlled process exit code without calling
// os.Exit inside command handlers.
type CommandExitError struct {
	Code    int
	Message string
}

func NewExitError(code int, message string) CommandExitError {
	return CommandExitError{
		Code:    code,
		Message: message,
	}
}

func (e CommandExitError) Error() string {
	return e.Message
}

func (e CommandExitError) ExitCode() int {
	return e.Code
}
