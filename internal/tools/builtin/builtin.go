// Package builtin registers all built-in tools.
package builtin

import (
	"github.com/shtdu/ohgo/internal/tools"
	"github.com/shtdu/ohgo/internal/tools/bash"
	"github.com/shtdu/ohgo/internal/tools/edit"
	"github.com/shtdu/ohgo/internal/tools/glob"
	"github.com/shtdu/ohgo/internal/tools/grep"
	"github.com/shtdu/ohgo/internal/tools/lsp"
	"github.com/shtdu/ohgo/internal/tools/read"
	"github.com/shtdu/ohgo/internal/tools/webfetch"
	"github.com/shtdu/ohgo/internal/tools/websearch"
	"github.com/shtdu/ohgo/internal/tools/write"
)

// RegisterAll registers all built-in tools into the registry.
func RegisterAll(r *tools.Registry) {
	r.Register(read.ReadTool{})
	r.Register(write.WriteTool{})
	r.Register(edit.EditTool{})
	r.Register(bash.BashTool{})
	r.Register(glob.GlobTool{})
	r.Register(grep.GrepTool{})
	r.Register(webfetch.FetchTool{})
	r.Register(websearch.SearchTool{})
	r.Register(lsp.LspTool{})
}
