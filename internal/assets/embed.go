// Package assets provides embedded builtin species packs.
package assets

import "embed"

// BuiltinFS contains all builtin species pack data files.
//
//go:embed builtins
var BuiltinFS embed.FS
