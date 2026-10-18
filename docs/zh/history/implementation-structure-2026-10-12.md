---
search: false
---

# Rivu 早期目录规划

::: info 历史规划
日期：2026 年 10 月 12 日。以下为提议目录，不是已实现文件清单。当时基于 Go 1.27.2 和 MyGo v0.4.3。当前布局见[源码结构](../development/implementation-structure.md)。
:::

根包组合生命周期、窗口、视图与命令，`internal` 承担传输、持久化和原生适配。[产品设计](product-design.md)提供目标，[实施计划](implementation-plan.md)说明契约、状态与验收。生产视图使用 Go。

## 目标源码树

功能落地时再创建对应文件。测试与实现同目录，下表中的测试名只是代表，不要求机械地为每个文件建立测试。

```text
rivu/
├── AGENTS.md
├── go.mod
├── go.sum
├── mygo.json
├── Makefile
├── main.go
├── application.go
├── application_test.go
├── lifecycle.go
├── launch.go
├── cli.go
├── cli_test.go
├── commands.go
├── menu.go
├── window.go
├── window_actions.go
├── view.go
├── view_sidebar.go
├── view_toolbar.go
├── view_task_list.go
├── view_inspector.go
├── view_statusbar.go
├── view_palette.go
├── view_test.go
├── add_window.go
├── add_view.go
├── remove_window.go
├── settings_window.go
├── settings_view.go
├── settings_general.go
├── settings_downloads.go
├── settings_connections.go
├── settings_notifications.go
├── settings_updates.go
├── tray.go
├── tray_view.go
├── presentation.go
├── presentation_test.go
├── theme.go
├── icons.go
├── capture.go
├── internal/
│   ├── session/
│   │   ├── model.go
│   │   ├── manager.go
│   │   ├── command.go
│   │   ├── ports.go
│   │   ├── snapshot.go
│   │   ├── import.go
│   │   ├── scheduler.go
│   │   ├── pause.go
│   │   ├── bandwidth.go
│   │   ├── seeding.go
│   │   ├── recovery.go
│   │   ├── persistence.go
│   │   ├── remove.go
│   │   ├── manager_test.go
│   │   ├── import_test.go
│   │   ├── scheduler_test.go
│   │   ├── pause_test.go
│   │   ├── recovery_test.go
│   │   └── persistence_test.go
│   ├── torrent/
│   │   ├── engine.go
│   │   ├── source.go
│   │   ├── metadata.go
│   │   ├── storage.go
│   │   ├── observation.go
│   │   ├── engine_test.go
│   │   └── engine_integration_test.go
│   ├── config/
│   │   ├── model.go
│   │   ├── store.go
│   │   ├── migrate.go
│   │   └── store_test.go
│   ├── desktop/
│   │   ├── desktop.go
│   │   ├── dialogs.go
│   │   ├── login.go
│   │   ├── login_status_darwin.go
│   │   ├── login_status_other.go
│   │   ├── notifications.go
│   │   ├── power.go
│   │   ├── trash.go
│   │   └── desktop_test.go
│   ├── files/
│   │   ├── destination.go
│   │   ├── paths.go
│   │   ├── ownership.go
│   │   ├── space_darwin.go
│   │   ├── space_other.go
│   │   └── paths_test.go
│   ├── updater/
│   │   ├── service.go
│   │   ├── policy.go
│   │   ├── mygo.go
│   │   └── policy_test.go
│   ├── storage/
│   │   ├── atomic.go
│   │   └── atomic_test.go
│   └── tasks/
│       ├── task.go
│       └── task_test.go
├── resources/
│   └── icon.png
├── tools/genicon/main.go
├── testdata/
│   ├── torrents/
│   └── sessions/
├── docs/
│   ├── architecture.md
│   ├── desktop.md
│   ├── releasing.md
│   ├── product-design.md
│   ├── implementation-structure.md
│   └── implementation-plan.md
├── build/
└── snapshots/
```

## 职责与依赖规则

| 模块    | 持有                                  | 不应持有                        |
| ------- | ------------------------------------- | ------------------------------- |
| 根 main | MyGo 生命周期、窗口、托盘、命令和展示 | 引擎内部对象、第二份传输状态    |
| session | 意图、队列、任务记录、编排和快照      | MyGo 或 anacrolix 对象          |
| torrent | 协议、来源解析、引擎存储适配          | 窗口生命周期、产品暂停/队列策略 |
| config  | 设置、验证和 schema 迁移              | 传输计数器和活动引擎            |
| desktop | 原生调用、能力和状态观测              | 调度或独立传输控制              |
| files   | 目录安全、空间、归属                  | UI 状态或协议执行               |
| updater | 更新生命周期和策略                    | 直接修改任务、无条件退出        |
| storage | 原子文件原语                          | 产品 schema                     |
| tasks   | 短 UI 操作、过时回调抑制              | 长期种子会话                    |

session 定义消费的引擎接口；torrent 导入普通 Go 类型并实现，session 不反向导入 torrent。main 构造实现并注入。确定性内存引擎作为测试的第二个适配器，避免通用事件总线、依赖注入框架和纯转发包。

session 可依赖 config、files、storage，后者不依赖 session。updater 从 main 获取活动观测和重启协调，无需依赖具体 Manager。视图帮助逻辑在真实第二个包需要前留在 main。

## 界面到文件的映射

| 界面             | 视图与控制                                 | 行为归属                              |
| ---------------- | ------------------------------------------ | ------------------------------------- |
| 主窗口           | `window.go`、`view*.go`、`presentation.go` | session.Manager                       |
| 概览、文件、连接 | `view_inspector.go`                        | 会话快照与选中任务详情                |
| 菜单栏           | `tray.go`、`tray_view.go`                  | 同一会话与命令注册表                  |
| 添加下载         | `add_window.go`、`add_view.go`             | session/import、torrent/source、files |
| 移除确认         | `remove_window.go`                         | session/remove、归属、原生 Trash      |
| 设置             | `settings_*.go`                            | config、会话策略、桌面状态和更新器    |
| 更新与重启       | `settings_updates.go`、`lifecycle.go`      | updater 与退出协调器                  |

## 从初始项目迁移

- `main.go` 保留启动和事件注册，归属移至 application，文件/URL 进入导入预览。
- `window.go` 移出应用、托盘、全局快捷键，以本地选择和快照替换文档/校验和状态。
- `view.go` 以任务布局替代五个示例页，复杂区域拆到根视图文件。
- `menu.go` 使用稳定命令 ID，菜单、命令面板、托盘共享处理。
- `internal/files/files.go` 在替代视图/测试完成后移除文本预览和校验和示例，增加目录与归属功能。
- `internal/tasks/task.go` 继续处理对话框和草稿，不扩展为多传输引擎。
- config 保留坏文件保护和原子发布，迁移主题并引入类型化设置；storage 保留原子替换并说明持久性限制。
- `mygo.json`、导入路径与捕获变量一起采用 Rivu 身份：模块 `github.com/ntnyq/rivu`、Bundle ID `com.ntnyq.rivu`、环境前缀 `RIVU_`。
- 架构和桌面文档始终描述当时真实代码，不能将目标目录写成已经可用。

## 运行数据目录

macOS 默认根为 `~/Library/Application Support/Rivu`，其他平台遵循系统配置约定。下载内容与应用元数据分离。

```text
Application Support/Rivu/
├── settings.json          # 设置、同意状态、普通/临时限速
├── session.json           # 权威任务、暂停批次、操作状态
├── metadata/
│   └── <task-id>.torrent   # 已验证的不可变元数据，含解析后磁力
└── engine/
    └── <task-id>/          # 可丢弃的版本化引擎缓存

~/Downloads/Rivu/
└── <torrent-root>/         # 用户内容，挂接前检查冲突
```

窗口位置继续使用 MyGo 机制，进度缓存不替代数据校验。携带认证信息的磁力与 Tracker 仅存在于权限受限记录，并在日志脱敏。下载内容不放入 docs、resources、build 或 snapshots。
