package cli

import (
	"github.com/spf13/cobra"
	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
	"time"
)

// PeekGridCmd temporarily masks the active grid without giving up its input.
var PeekGridCmd = &cobra.Command{
	Use: "peek-grid", Short: "Temporarily see through the active grid without leaving it",
	Long: `Make the active grid or recursive-grid overlay invisible for a short duration.

The mode, selection and keyboard capture remain active. You can keep typing
while it is invisible; the updated grid reappears when the deadline expires.
Another peek restarts the timeout. Leaving the mode cancels restoration.
No colors are changed and no configuration is reloaded. macOS only.

Examples:
  neru peek-grid
  neru peek-grid --duration 1s`,
	Args:    cobra.NoArgs,
	PreRunE: func(_ *cobra.Command, _ []string) error { return requiresRunningInstance() },
	RunE: func(cmd *cobra.Command, _ []string) error {
		duration, err := cmd.Flags().GetDuration("duration")
		if err != nil {
			return err
		}
		if duration < time.Millisecond || duration > 10*time.Second {
			return derrors.New(derrors.CodeInvalidInput, "peek duration must be between 1ms and 10s")
		}
		return sendCommand(cmd, domain.CommandPeekGrid, []string{"--duration=" + duration.String()})
	},
}

func init() {
	PeekGridCmd.Flags().Duration("duration", time.Second, "How long to mask the grid (1ms to 10s)")
	RootCmd.AddCommand(PeekGridCmd)
}
