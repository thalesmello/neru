package grid

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// The bridge clear and draw are separately queued Cocoa operations: an empty
// frame can be presented between them. Empty grids and explicit teardown may
// clear, but a full replacement must keep the old cells until the new draw.
func TestDarwinGridFullReplacement_DoesNotClearBeforeDrawing(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "overlay_darwin.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "DrawGrid" {
			continue
		}
		found = true
		for _, statement := range fn.Body.List {
			expression, ok := statement.(*ast.ExprStmt)
			if !ok {
				continue
			}
			call, ok := expression.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if ok && selector.Sel.Name == "Clear" {
				t.Error("full replacement queues a native clear before draw")
			}
		}
	}
	if !found {
		t.Fatal("DrawGrid method missing from native renderer")
	}
}
