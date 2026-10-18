//go:build !windows

package config

import "github.com/egoist/mygo"

func downloadsDirectory() (string, error) {
	return mygo.App.Path(mygo.PathDownloads)
}
