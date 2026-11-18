package cli

import (
	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
)

var SetGridRepeatCmd = &cobra.Command{
	Use: "set-grid-repeat", Short: "Change active grid repeat mode without resetting selection",
	Args:    cobra.NoArgs,
	PreRunE: func(_ *cobra.Command, _ []string) error { return requiresRunningInstance() },
	RunE: func(cmd *cobra.Command, _ []string) error {
		value, err := cmd.Flags().GetString("state")
		if err != nil {
			return err
		}
		if value != "on" && value != "off" {
			return derrors.New(derrors.CodeInvalidInput, "--state must be on or off")
		}
		return sendCommand(cmd, domain.CommandSetGridRepeat, []string{"--state", value})
	},
}

func init() {
	SetGridRepeatCmd.Flags().String("state", "", "Repeat confirmation: on or off")
	RootCmd.AddCommand(SetGridRepeatCmd)
}
