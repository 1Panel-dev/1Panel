package web

import "embed"

//go:embed index.html
var IndexHtml embed.FS

//go:embed assets/*
var Assets embed.FS

//go:embed favicon.png
var Favicon embed.FS
