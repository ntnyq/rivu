# Current implementation structure

Structure recorded October 17, 2026. The current `go.mod` pins Go 1.27.2 and MyGo v0.5.5; earlier verification records used MyGo v0.4.8. The root `main.go` delegates to `internal/app.Run`. Native application composition stays in one package, while business logic retains its existing boundaries.

```text
rivu/
├── main.go
├── mygo.json
├── Makefile
├── internal/
│   ├── app/
│   ├── config/
│   ├── desktop/
│   ├── files/
│   ├── problem/
│   ├── session/
│   ├── storage/
│   ├── torrent/
│   ├── updater/
│   ├── watchfolder/
│   └── windowtask/
├── resources/
│   ├── embed.go
│   └── ...
├── locales/
├── tools/genicon/
└── docs/
    ├── guide/
    ├── development/
    ├── history/
    ├── images/
    ├── public/
    ├── zh/
    └── .vitepress/
```

Application filenames below are relative to `internal/app/`; session and torrent filenames are relative to their respective packages.

| Area                       | Files and ownership                                                                                                                                                                                                                                         |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Bootstrap                  | Root `main.go`: thin entry; app `run.go`, `cli.go`, `application.go`: process, single session, app lifecycle                                                                                                                                                |
| Windows                    | `window_base.go`: native lifetime and cancellable work; `window.go`: main state; `window_actions.go`: main actions; `settings_window.go`: settings draft; `add_window.go`: import draft                                                                     |
| Main UI                    | `view.go`, `view_sidebar.go`, `view_toolbar.go`, `view_task_list.go`, `view_statusbar.go`, `view_inspector.go`, `view_palette.go`                                                                                                                           |
| Dialogs and inputs         | `add_view.go`, `remove_window.go`, `recovery_actions.go`, `file_tree.go`                                                                                                                                                                                    |
| Preferences                | `settings.go`: rendering; `settings_inputs.go`: units, policy controls and automation; `settings_save.go`: validation/persistence; `settings_updates.go`: updater                                                                                           |
| Desktop composition        | `menu.go`, `tray.go`, `capture.go`, `localization.go`, `presentation.go`, `theme.go`; `icons.go`: SVG parsing/cache; `controls.go`: UI controls                                                                                                             |
| Session                    | `manager.go`: lifecycle/serialization; `model.go`, `types.go`: data contracts; `validation.go`, `commands.go`: validated mutations; `import.go`, `removal.go`, `persistence.go`, `recovery.go`, `observation.go`, `snapshot.go`, `scheduler.go`, `space.go` |
| Torrent adapter            | `engine.go`, `source.go`, `storage.go`, `preview.go`, `diagnostics.go`                                                                                                                                                                                      |
| Automation                 | `internal/config/schedule.go`, `internal/watchfolder/watcher.go`, `internal/app/watch_folder.go`                                                                                                                                                            |
| Supporting packages        | `internal/problem`, `config`, `files`, `storage`, `desktop`, `updater`; `internal/windowtask` manages cancellable window jobs, separate from transfers                                                                                                      |
| Resources and localization | `resources/embed.go`: embedded artwork and relative-path reads; `locales`: MyGo catalogs and generated accessors                                                                                                                                            |

Tests remain beside implementations, including application tests in `internal/app/`. `resources` contains shipped assets, `tools/genicon` generates icons, `snapshots` and `build` contain generated output. The native application has no web runtime. The VitePress documentation in `docs` uses a separate pnpm toolchain.

`go run .` and MyGo dev/build commands still use the root entry. `make snapshots` targets `internal/app`; the release workflow updates the version in `internal/app/run.go`.

See [architecture](architecture.md) for ownership and persistence contracts, [validation report](refactor-validation.md) for feature behavior, and [desktop acceptance](desktop.md) before release. The [October 12 proposed tree](../history/implementation-structure-2026-10-12.md) is retained as historical design, not current implementation documentation.
