/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

var (
	// `gPidFile` is the path to the PID file.
	gPidFile = func() string {
		me, _ := filepath.Abs(os.Args[0])

		return filepath.Join(os.TempDir(), filepath.Base(me)+".pid")
	}()
)

// `isInstanceRunning()` checks if another instance is already running.
//
// This function checks for the existence of a process ID file and
// verifies that the process with the ID in the file is actually running.
//
// Returns:
//   - `bool`: `true` if another instance is running, `false` otherwise.
func isInstanceRunning() bool {
	// Check/read PID from file
	data, err := os.ReadFile(gPidFile) //#nosec G304
	if nil != err {
		// No PID file, assume no instance is running
		return false
	}

	// Parse PID
	pidStr := strings.TrimSpace(string(data))
	pid, err := strconv.Atoi(pidStr)
	if nil != err {
		// Invalid PID, assume no instance is running
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if nil != err {
		// Process not found, assume no instance is running
		return false
	}

	// On Unix-like systems, FindProcess always succeeds, so we need to
	// send signal 0 to check if the process actually exists
	err = process.Signal(os.Signal(syscall.Signal(0)))

	return nil == err
} // isInstanceRunning()

// `runAsDaemon()` forks the process to run as a daemon.
//
// Returns:
//   - `error`: `nil` if the process was forked successfully, the error otherwise.
func runAsDaemon() error {
	// Get the path to the current executable
	executable, err := os.Executable()
	if nil != err {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if nil != err {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Prepare command arguments, removing the daemon flag
	args := []string{}
	for _, arg := range os.Args[1:] {
		if ("--daemon" != arg) && ("-daemon" != arg) {
			args = append(args, arg)
		}
	}

	// Create command
	cmd := exec.Command(executable, args...) //#nosec G204
	cmd.Dir = cwd

	// Detach from parent process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	// Redirect standard file descriptors to /dev/null
	devNull, err := os.OpenFile("/dev/null", os.O_RDWR, 0)
	if nil != err {
		return fmt.Errorf("failed to open /dev/null: %w", err)
	}
	// Don't defer close here as the child process will use this file
	// descriptor. The OS will close it when the process exits.

	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull

	//TODO: implement log file

	// Start the daemon
	if err := cmd.Start(); nil != err {
		// Only close `devNull` if we failed to start the daemon
		_ = devNull.Close()

		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Write PID to file
	if err := os.WriteFile(gPidFile, fmt.Appendf(nil, "%d", cmd.Process.Pid), 0600); nil != err {
		//TODO: Should we try to kill the daemon process here?
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// We can close `devNull` in the parent process since the child process
	// has its own copy of the file descriptor after fork
	_ = devNull.Close()

	fmt.Printf("Started daemon with PID %d\n", cmd.Process.Pid)

	return nil
} // runAsDaemon()

/* _EoF_ */
