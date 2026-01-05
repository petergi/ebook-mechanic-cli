package cli

import (
	"testing"
)

func init() {
	// Mock osExit to prevent tests from exiting
	osExit = func(code int) {
		// Do nothing or record the code if needed
	}
}

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()
	if cmd.Use != "ebm" {
		t.Errorf("expected command name 'ebm', got %s", cmd.Use)
	}

	if !cmd.HasSubCommands() {
		t.Error("expected root command to have subcommands")
	}
}

func TestExecute(t *testing.T) {
	// Execute without args should show help or return error
	// but since we mocked osExit, it won't crash
	_ = Execute()
}

func TestRootFlags_Defaults(t *testing.T) {
	cmd := NewRootCmd()

	format, _ := cmd.PersistentFlags().GetString("format")
	if format != "text" {
		t.Errorf("expected default format 'text', got %s", format)
	}

	color, _ := cmd.PersistentFlags().GetBool("color")
	if !color {
		t.Error("expected default color to be true")
	}
}
