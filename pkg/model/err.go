package model

import "github.com/peakedshout/go-pandorasbox/tool/xerror"

var (
	ErrElement = xerror.New("Element <%s> err: %v")
	ErrPage    = xerror.New("Page [%s] err: %v")
)
