package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCompletionCmd(rootCmd *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:

  $ source <(ebm completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ ebm completion bash > /etc/bash_completion.d/ebm
  # macOS:
  $ ebm completion bash > /usr/local/etc/bash_completion.d/ebm

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ ebm completion zsh > "${fpath[1]}/_ebm"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ ebm completion fish | source

  # To load completions for each session, execute once:
  $ ebm completion fish > ~/.config/fish/completions/ebm.fish

PowerShell:

  PS> ebm completion powershell | Out-String | Invoke-Expression

  # To load completions for each session, execute once:
  PS> ebm completion powershell > ebm.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return rootCmd.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
}
