package cli

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
)

// NudgeGridCmd moves the active native Grid selection without clicking.
var NudgeGridCmd = &cobra.Command{
	Use: "nudge-grid", Short: "Nudge the active native Grid selection without clicking",
	Args:    cobra.NoArgs,
	PreRunE: func(_ *cobra.Command, _ []string) error { return requiresRunningInstance() },
	RunE: func(cmd *cobra.Command, _ []string) error {
		direction, err := cmd.Flags().GetString("direction")
		if err != nil {
			return err
		}
		dir, err := domain.ParseDirection(direction)
		if err != nil {
			return err
		}
		nudges, err := cmd.Flags().GetInt("nudges-per-cell")
		if err != nil {
			return err
		}
		count, err := cmd.Flags().GetInt("count")
		if err != nil {
			return err
		}
		if nudges <= 0 || count < 0 {
			return derrors.New(derrors.CodeInvalidInput, "nudges-per-cell must be positive and count must be nonnegative")
		}
		return sendCommand(cmd, domain.CommandNudgeGrid, []string{"--direction", dir.String(), "--nudges-per-cell", strconv.Itoa(nudges), "--count", strconv.Itoa(count)})
	},
}

func init() {
	NudgeGridCmd.Flags().String("direction", "", "Direction: up, down, left, or right (required)")
	NudgeGridCmd.Flags().Int("nudges-per-cell", 4, "Nudges spanning one active cell")
	NudgeGridCmd.Flags().Int("count", 1, "Nudges to apply; zero initializes selection without movement")
	RootCmd.AddCommand(NudgeGridCmd)
}
