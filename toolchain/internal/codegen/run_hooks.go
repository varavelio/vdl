package codegen

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type hostHookPhase string

const (
	hostHookPhasePreGenerate  hostHookPhase = "preGenerate"
	hostHookPhasePostGenerate hostHookPhase = "postGenerate"
)

var (
	hostHookCommandRunner           = runShellCommand
	hostHookWarningWriter io.Writer = os.Stderr
)

// runPreGenerateHooks executes configured pre-generate host commands in order.
// The first command failure aborts generation.
func runPreGenerateHooks(config runtimeConfig) error {
	hooks := config.Config.GetHooks()
	return runHostHooks(config, hostHookPhasePreGenerate, hooks.GetPreGenerate(), false)
}

// runPostGenerateHooks executes configured post-generate host commands in
// order. Failures are reported as warnings and do not abort generation.
func runPostGenerateHooks(config runtimeConfig) {
	hooks := config.Config.GetHooks()
	_ = runHostHooks(config, hostHookPhasePostGenerate, hooks.GetPostGenerate(), true)
}

// runHostHooks runs lifecycle hooks in definition order.
func runHostHooks(config runtimeConfig, phase hostHookPhase, commands []string, continueOnError bool) error {
	if hostHooksDisabled() || len(commands) == 0 {
		return nil
	}

	for i, rawCommand := range commands {
		command := strings.TrimSpace(rawCommand)
		if command == "" {
			err := fmt.Errorf("%s hook command %d is empty", phase, i+1)
			if continueOnError {
				printHookWarning(err)
				continue
			}
			return err
		}

		if err := hostHookCommandRunner(config.Dir, command); err != nil {
			wrapped := fmt.Errorf("%s hook command %d failed: %w", phase, i+1, err)
			if continueOnError {
				printHookWarning(wrapped)
				continue
			}
			return wrapped
		}
	}

	return nil
}

// hostHooksDisabled reports whether host hook execution is disabled for the
// current process. Cloud runtimes can set one of these variables to skip host
// shell commands.
func hostHooksDisabled() bool {
	return isTruthyEnv("VDL_SKIP_HOST_HOOKS") || isTruthyEnv("VDL_CLOUD")
}

// runShellCommand executes a shell command in the provided working directory.
func runShellCommand(workDir, command string) error {
	shell, shellFlag := shellProgram()

	cmd := exec.Command(shell, shellFlag, command)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %q in %q failed: %w", command, workDir, err)
	}

	return nil
}

// shellProgram returns the shell executable and argument used to run a command
// string on the current platform.
func shellProgram() (string, string) {
	if runtime.GOOS == "windows" {
		return "cmd", "/C"
	}

	return "sh", "-c"
}

func printHookWarning(err error) {
	_, _ = fmt.Fprintf(hostHookWarningWriter, "VDL warning: %v\n", err)
}
