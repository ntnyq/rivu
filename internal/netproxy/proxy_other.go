//go:build !windows

package netproxy

import (
	"context"
	"net/url"
)

func systemProxy(context.Context, *url.URL) (*url.URL, error) { return nil, nil }
