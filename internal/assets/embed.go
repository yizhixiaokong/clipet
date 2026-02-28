// Package assets provides embedded builtin species packs and locales.
package assets

import "embed"

// BuiltinFS contains all builtin species pack data files.
//
//go:embed builtins
var BuiltinFS embed.FS

// LocalesFS contains all translation files.
//
//go:embed locales
var LocalesFS embed.FS
