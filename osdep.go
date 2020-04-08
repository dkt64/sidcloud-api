// +build darwin

package osdep

func prepareBackgroundCommand(cmd *exec.Cmd) {
    cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// +build !darwin

package osdep

func prepareBackgroundCommand(cmd *exec.Cmd) {
}
