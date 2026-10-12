package main

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type launchOptions struct {
	configPath, dataDir string
	sources             []string
	showVersion         bool
}

func parseLaunch(args []string, cwd string, output io.Writer) (launchOptions, error) {
	var o launchOptions
	f := flag.NewFlagSet("rivu", flag.ContinueOnError)
	f.SetOutput(output)
	f.BoolVar(&o.showVersion, "version", false, "print version")
	f.StringVar(&o.configPath, "config", "", "preferences file (does not isolate session)")
	f.StringVar(&o.dataDir, "data-dir", "", "application data directory (isolates session)")
	f.Usage = func() {
		fmt.Fprintln(output, "Usage: rivu [--config path] [--data-dir path] [torrent files or magnet links…]")
		f.PrintDefaults()
	}
	if err := f.Parse(args); err != nil {
		return o, err
	}
	resolve := func(p string) string {
		if p != "" && !filepath.IsAbs(p) {
			return filepath.Join(cwd, p)
		}
		return p
	}
	for _, src := range f.Args() {
		if strings.HasPrefix(strings.ToLower(src), "magnet:") {
			o.sources = append(o.sources, src)
		} else {
			o.sources = append(o.sources, resolve(src))
		}
	}
	o.configPath = resolve(o.configPath)
	o.dataDir = resolve(o.dataDir)
	return o, nil
}
