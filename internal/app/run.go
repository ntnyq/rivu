package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"
	"github.com/ntnyq/rivu/internal/config"
	"github.com/ntnyq/rivu/internal/session"
	torrentengine "github.com/ntnyq/rivu/internal/torrent"
	"github.com/ntnyq/rivu/internal/updater"
	"github.com/ntnyq/rivu/locales"
)

const appName = "Rivu"
const version = "0.0.2"
const website = "https://rivu.ntnyq.dev"
const repository = "https://github.com/ntnyq/rivu"

// Run starts Rivu and handles process exit for launch or runtime failures.
func Run() {
	cwd, err := os.Getwd()
	if err != nil {
		slog.Error("working directory", "error", err)
		os.Exit(1)
	}
	opts, err := parseLaunch(os.Args[1:], cwd, os.Stderr)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if opts.showVersion {
		fmt.Println(appName, version)
		return
	}
	if opts.dataDir == "" {
		dir, e := os.UserConfigDir()
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			os.Exit(1)
		}
		opts.dataDir = filepath.Join(dir, "Rivu")
	}
	if opts.configPath == "" {
		opts.configPath = filepath.Join(opts.dataDir, "settings.json")
	}
	store, loadErr := config.Open(opts.configPath)
	mygo.App.SetName(appName)
	mygo.App.SetVersion(version)
	if !mygo.App.RequestSingleInstanceLock() {
		return
	}
	a := &application{store: store, pending: opts.sources, notifying: map[string]bool{}, errorsSeen: map[string]string{}}
	mygo.App.OnOpenFile(func(path string) { mygo.RunOnMain(func() { a.receive([]string{path}) }) })
	mygo.App.OnOpenURL(func(raw string) { mygo.RunOnMain(func() { a.receive([]string{raw}) }) })
	mygo.App.OnSecondInstance(func(args []string, cwd string) {
		mygo.RunOnMain(func() {
			if len(args) > 0 {
				args = args[1:]
			}
			launch, err := parseLaunch(args, cwd, io.Discard)
			if err != nil {
				a.report(err)
				return
			}
			a.openWindow()
			a.receive(launch.sources)
		})
	})
	mygo.App.OnActivate(func(visible bool) {
		if a.ready && !visible {
			a.openWindow()
		}
	})
	mygo.App.OnWindowAllClosed(func() {})
	mygo.App.OnBeforeQuit(a.beforeQuit)
	mygo.App.SetLocale(localePreference(store.Snapshot().Language))
	mygo.App.OnLocaleChanged(func(string) {
		if a.ready {
			a.refreshLanguage()
		}
	})
	mygo.App.OnQuit(func() {
		if a.cancel != nil {
			a.cancel()
		}
		if a.tray != nil {
			a.tray.Destroy()
		}
		if a.window != nil {
			a.window.close()
		}
	})
	mygo.App.WhenReady(func() {
		applyTheme(store.Snapshot().Theme)
		go func() {
			engine := torrentengine.New(store.Snapshot())
			manager, sessionErr := session.Open(opts.dataDir, engine, store.Snapshot(), mygo.Shell.TrashItem)
			mygo.RunOnMain(func() {
				a.session = manager
				a.updater = updater.New(store, manager)
				a.ready = true
				a.observe()
				mygo.App.SetMenu(a.menu())
				a.configureTray()
				if !mygo.App.WasOpenedAtLogin() || !store.Snapshot().BackgroundLogin || len(a.pending) > 0 || loadErr != nil || sessionErr != nil {
					a.openWindow()
				}
				if loadErr != nil {
					a.report(messageError{locales.SettingsReadFailed(loadErr), loadErr})
				}
				if sessionErr != nil {
					a.report(sessionErr)
				}
				sources := a.pending
				a.pending = nil
				a.receive(sources)
				mygo.Power.OnSuspend(func() {
					go func() {
						if err := a.session.Quiesce(); err != nil {
							mygo.RunOnMain(func() { a.report(err) })
						}
					}()
				})
				mygo.Power.OnResume(func() { go a.session.Continue() })
				if a.window != nil {
					a.window.captureIfRequested()
				}
			})
		}()
	})
	if err := mygo.App.Run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}
