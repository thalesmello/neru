package overlay_test

import (
	"context"
	"image"
	"testing"

	"github.com/y3owk1n/neru/internal/adapter/overlay"
	"github.com/y3owk1n/neru/internal/config"
	"github.com/y3owk1n/neru/internal/domain"
	domainGrid "github.com/y3owk1n/neru/internal/domain/grid"
	"github.com/y3owk1n/neru/internal/ports"
	"go.uber.org/zap"
)

type gridReplacementManager struct {
	screenManager
	replacements int
}

func (m *gridReplacementManager) PrepareGridReplacement() { m.replacements++ }

func TestAdapter_RedrawFrame_ReplacesSubgridWithoutClearingWindow(t *testing.T) {
	m := &gridReplacementManager{}
	cfg := config.DefaultConfig()
	a := overlay.NewAdapter(m, overlay.NewStyleResolver(m, cfg, nil, nil), nil)
	g := domainGrid.NewGridWithLabels("abcd", "", "", image.Rect(0, 0, 800, 600), zap.NewNop())
	frame := ports.GridFrame{Grid: g}
	if err := a.ShowFrame(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	a.ShowGridSubgrid(g.Cells()[0], ports.GridPointer{})
	clears, resizes := m.cleared, m.resizes
	if err := a.RedrawFrame(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	if m.replacements != 1 || m.cleared != clears || m.resizes != resizes || !m.visible {
		t.Fatalf("reset did not preserve surface: replacements=%d clears=%d resizes=%d visible=%v", m.replacements, m.cleared, m.resizes, m.visible)
	}
	if err := a.RedrawFrame(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	if m.replacements != 1 {
		t.Fatal("replacement cache invalidated twice")
	}
}

func TestAdapter_RedrawFrame_RecursiveResetKeepsWindow(t *testing.T) {
	m := &screenManager{}
	a := overlay.NewAdapter(m, overlay.NewStyleResolver(m, config.DefaultConfig(), nil, nil), nil)
	frame := ports.RecursiveGridFrame{Bounds: image.Rect(0, 0, 800, 600), Depth: 0,
		Layout: ports.RecursiveGridLayout{Keys: "asdf", Dimensions: domain.GridDimensions{Rows: 2, Cols: 2}}}
	if err := a.ShowFrame(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	clears, resizes := m.cleared, m.resizes
	if err := a.RedrawFrame(context.Background(), frame); err != nil {
		t.Fatal(err)
	}
	if m.cleared != clears || m.resizes != resizes || !m.visible {
		t.Fatal("recursive reset took down window")
	}
}
