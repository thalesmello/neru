package cli

import (
	"fmt"
	"math"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/y3owk1n/neru/internal/derrors"
	"github.com/y3owk1n/neru/internal/domain"
)

// SetGridOpacityCmd overrides the grid cell-fill alpha for the current daemon
// session without touching configuration or mode state.
var SetGridOpacityCmd = &cobra.Command{
	Use:   "set-grid-opacity",
	Short: "Override the grid cell fill alpha",
	Long: `Override the cell background alpha used by grid and recursive-grid mode.

The override applies immediately and lasts until the daemon restarts. It is a
runtime visual adjustment only: configuration is not written, the active mode
is not exited, and no grid reset or mode reload occurs. The command is allowed
in idle as well as while a grid mode is running, so a binding can pre-set the
opacity before activating grid.

Opacity 0 makes cells fully transparent; opacity 1 makes them fully opaque.
Label badges, text, separators, and non-grid modes are unaffected.

Returns ERR_NOT_SUPPORTED when the active overlay backend does not implement
GridAppearancePort (all non-macOS platforms for now).

Examples:
  neru set-grid-opacity --value 0.45
  neru set-grid-opacity --value 0
  neru set-grid-opacity --value 1`,
	Args: cobra.NoArgs,
	PreRunE: func(_ *cobra.Command, _ []string) error {
		return requiresRunningInstance()
	},
	RunE: setGridOpacityRun,
}

var setGridOpacityValue float64

func setGridOpacityRun(cmd *cobra.Command, _ []string) error {
	if !cmd.Flags().Changed("value") {
		return derrors.New(derrors.CodeInvalidInput, "--value is required")
	}

	if math.IsNaN(setGridOpacityValue) || math.IsInf(setGridOpacityValue, 0) {
		return derrors.New(derrors.CodeInvalidInput, "--value must be a finite number")
	}

	if setGridOpacityValue < 0 || setGridOpacityValue > 1 {
		return derrors.Newf(
			derrors.CodeInvalidInput,
			"--value %s is out of range [0, 1]",
			strconv.FormatFloat(setGridOpacityValue, 'f', -1, 64),
		)
	}

	arg := fmt.Sprintf("--value=%s",
		strconv.FormatFloat(setGridOpacityValue, 'f', -1, 64))

	return sendCommand(cmd, domain.CommandSetGridOpacity, []string{arg})
}

func init() {
	SetGridOpacityCmd.Flags().Float64Var(
		&setGridOpacityValue,
		"value",
		0,
		"Cell fill alpha: a finite number in [0, 1]",
	)

	RootCmd.AddCommand(SetGridOpacityCmd)
}
