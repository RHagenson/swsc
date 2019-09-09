package ui_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/rhagenson/swsc/internal/ui"
)

func TestErrorf(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "2" {
		ui.Errorf("%s", "testing exit code 2")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestErrorf")
	cmd.Env = append(os.Environ(), "BE_CRASHER=2")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 2", err)
}
