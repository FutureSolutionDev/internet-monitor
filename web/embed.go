package web

import "embed"

//go:embed all:dist
var Dist embed.FS

//go:embed locales/*.json
var Locales embed.FS
