# 当前源码结构

结构记录更新于 2026 年 10 月 13 日。当前 `go.mod` 固定 Go 1.27.2 和 MyGo v0.5.5；早期验证记录使用 MyGo v0.4.8。根目录 Go 文件组合原生应用行为，业务逻辑位于 `internal`。

| 区域         | 文件与职责                                                                                                                                                                                                                          |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 启动         | `main.go`、`cli.go`、`application.go`：进程、唯一会话、应用生命周期                                                                                                                                                                 |
| 窗口         | `window_base.go`：原生生命周期和可取消工作；`window.go`：主窗口状态；`settings_window.go`：设置草稿；`add_window.go`：导入草稿                                                                                                      |
| 主界面       | `view.go`、`view_sidebar.go`、`view_toolbar.go`、`view_task_list.go`、`view_statusbar.go`、`view_inspector.go`、`view_palette.go`                                                                                                   |
| 对话框与输入 | `add_view.go`、`remove_window.go`、`recovery_actions.go`、`file_tree.go`                                                                                                                                                            |
| 设置         | `settings.go`：渲染；`settings_inputs.go`：单位、策略与自动化；`settings_save.go`：验证和保存；`settings_updates.go`：更新器                                                                                                        |
| 桌面组合     | `menu.go`、`tray.go`、`capture.go`、`localization.go`、`presentation.go`、`theme.go`、`icons.go`                                                                                                                                    |
| 会话         | `manager.go`：生命周期与串行化；`model.go`、`types.go`：数据契约；`validation.go`、`commands.go`：验证修改；`import.go`、`removal.go`、`persistence.go`、`recovery.go`、`observation.go`、`snapshot.go`、`scheduler.go`、`space.go` |
| 种子适配器   | `engine.go`、`source.go`、`storage.go`、`preview.go`、`diagnostics.go`                                                                                                                                                              |
| 自动化       | `internal/config/schedule.go`、`internal/watchfolder/watcher.go`、根目录 `watch_folder.go`                                                                                                                                          |
| 支撑包       | `internal/problem`、`config`、`files`、`storage`、`desktop`、`tasks`、`updater`；`locales` 存放 MyGo 词条和生成函数                                                                                                                 |

测试与实现同目录。`resources` 放置应用资源，`tools/genicon` 生成图标，`snapshots` 与 `build` 存放生成输出。原生应用不使用 Web 运行时；`docs` 下的 VitePress 站点独立使用 pnpm 构建。

状态归属与持久化契约见[架构](architecture.md)，功能行为见[验证报告](refactor-validation.md)，发布前阅读[桌面验收](desktop.md)。[10 月 12 日目录规划](history/implementation-structure-2026-10-12.md)保留为历史设计，不代表当前文件布局。
