package cli

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCompletionCmd(t *testing.T) {
	rootCmd := &cobra.Command{Use: "ebm"}
	cmd := NewCompletionCmd(rootCmd)

	if cmd.Use != "completion [bash|zsh|fish|powershell]" {
		t.Errorf("Unexpected use string: %s", cmd.Use)
	}

	// Test generation
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetArgs([]string{shell})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("Failed to generate %s completion: %v", shell, err)
			}
			if buf.Len() == 0 {
				t.Errorf("%s completion is empty", shell)
			}
		})
	}

	t.Run("invalid shell", func(t *testing.T) {
		cmd.SetArgs([]string{"invalid"})
		err := cmd.Execute()
		if err == nil {
			t.Error("Expected error for invalid shell")
		}
	})
}
