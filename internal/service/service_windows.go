//go:build windows

package service

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	taskName = "NvidiaRateLimiter"
)

func Install() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command("schtasks.exe",
		"/Create",
		"/TN", taskName,
		"/SC", "ONSTART",
		"/RU", "SYSTEM",
		"/RL", "HIGHEST",
		"/F",
		"/TR", fmt.Sprintf("\"%s\" install-run", exe),
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("install failed: %v: %s", err, string(out))
	}
	return nil
}

func Uninstall() error {
	cmd := exec.Command("schtasks.exe",
		"/Delete",
		"/TN", taskName,
		"/F",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("uninstall failed: %v: %s", err, string(out))
	}
	return nil
}
