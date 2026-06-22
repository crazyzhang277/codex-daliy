//go:build !windows

package service

import "fmt"

func Install() error   { return fmt.Errorf("windows service install is only supported on windows") }
func Uninstall() error { return fmt.Errorf("windows service uninstall is only supported on windows") }
