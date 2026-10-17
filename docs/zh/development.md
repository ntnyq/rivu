# 开发 Rivu

[文档首页](index.md) · [使用指南](usage.md) · [发布指南](releasing.md)

## 环境与依赖

原生应用使用 macOS 和 Go 1.27.2 或更高版本。MyGo 及其 CLI 工具由 `go.mod` 固定，无需 Node.js 包管理器即可构建应用。macOS 是受支持的运行平台；Windows 和 Linux 构建仅用于编译检查。

当前 `go.mod` 固定 MyGo v0.5.5，以及 anacrolix/torrent v1.61.1-0.20260927071845-d913b30f520e。Torrent 与 dht/v2 使用固定开发提交 `d913b30f520e` 和 `a9c967f9d68b`，以 Go 伪版本记录。它们兼容 missinggo/v2 v2.11.0 的 roaring/v2 位图，并移除了旧 DHT 对 `stm` 迭代器 API 的依赖。

以下命令均从仓库根目录运行。

## 本地运行

```sh
go mod download
make run
```

使用 `make dev` 启动 MyGo 开发模式。构建本地应用包而不生成磁盘映像：

```sh
go tool mygo build -skip-dmg
```

Apple Silicon 上的输出为 `build/darwin-arm64/Rivu.app`。本地构建默认采用 ad-hoc 签名，未经过 Developer ID 签名或公证。`make build` 还会生成平台安装包，通用构建与分发见[发布指南](releasing.md)。

## 启动选项

```sh
go run . /path/to/file.torrent 'magnet:?xt=urn:btih:…&tr=…'
go run . --data-dir /tmp/rivu-session
go run . --config /tmp/rivu-settings.json
go run . --version
```

`--data-dir` 修改会话和默认设置位置，`--config` 只修改设置位置。两者都不改变单实例身份或 MyGo 窗口状态位置。使用隔离测试会话前关闭已运行的实例。原生检查的实例套接字隔离见[桌面验证](desktop.md)。

## 验证

```sh
make check
make race
make integration
go tool mygo vet
make snapshots
```

`make check` 运行单元测试、Go vet 和 MyGo 词条检查。集成测试使用隔离的本地 swarm，不依赖公共种子。快照测试用演示任务渲染应用，覆盖两种外观及全部支持的语言。

原生修改还需要[桌面行为与验收](desktop.md)中的检查。渲染快照与交叉编译无法验证权限、菜单栏定位、辅助功能或签名更新。

捕获打包应用的原生空窗口并退出前，关闭其他实例并创建截图目录：

```sh
mkdir -p snapshots
RIVU_CAPTURE="$PWD/snapshots/native-empty.png" \
  build/darwin-arm64/Rivu.app/Contents/MacOS/Rivu --data-dir /tmp/rivu-smoke
```

空会话请使用新的数据目录。`--data-dir` 不隔离 MyGo 窗口状态。

## 文档站点

文档使用 VitePress 2，要求 Node.js 22.12 或更高版本，以及根目录 `package.json` 声明的 pnpm 版本。Go 应用不依赖此工具链。

```sh
pnpm install
pnpm docs:dev
pnpm docs:build
pnpm docs:preview
```

`docs:dev` 启动开发服务器；`docs:build` 生成 `docs/.vitepress/dist` 并检查站内死链；`docs:preview` 在本地预览已构建站点。

英文内容位于 `docs/`，中文对应页面位于 `docs/zh/`，两种语言保持相同文件名与目录结构。两个 `index.md` 使用 `layout: home`、`hero` 与 `features`。导航、侧边栏、搜索与界面翻译在 `docs/.vitepress/config.ts` 中配置。新增页面时同步补充中文页面与导航，确认语言切换能到达对应页面。

默认部署到域名根路径。部署到子路径时，构建和预览使用相同的 `DOCS_BASE`，路径必须以 `/` 开始和结束：

```sh
DOCS_BASE=/rivu/ pnpm docs:build
DOCS_BASE=/rivu/ pnpm docs:preview
```

将 `docs/.vitepress/dist` 部署到支持静态 HTML 的托管服务。构建站点不会配置 DNS、发布应用或启用自动更新。站点使用默认 `.html` 路由，无需配置无扩展名 URL 重写。

## 文档截图

`docs/images/` 中的产品截图来自应用 UI 渲染器，名称、速度和进度均为测试数据，README 已作说明。

相关 UI 修改后更新截图：

```sh
make snapshots
cp snapshots/light-1280.png docs/images/main-light.png
cp snapshots/dark-1280.png docs/images/main-dark.png
cp snapshots/menu-bar-light.png docs/images/menu-bar.png
```

提交前检查每张图片。`snapshots/` 保持忽略，仅提交 `docs/images/` 中筛选的图片。原生截图是独立验证步骤。

站点图标 `docs/public/icon.png` 来自 `resources/icons/rivu-256.png`，应用图标更新后同步复制。

## 图标

源文件是 [resources/icon.svg](https://github.com/ntnyq/rivu/blob/main/resources/icon.svg)。重新生成：

```sh
make icon
```

输出包括 16–1024 尺寸 PNG、`resources/rivu.icns`、`resources/rivu.ico`、macOS iconset 和模板菜单栏图标。修改 SVG 源文件，而非生成的位图。

## 本地化

`language` 设置接受默认值 `system`，以及 `zh-Hans`、`zh-Hant`、`en`、`ja`、`ko`。缺少该字段的旧设置继续跟随系统。

MyGo 嵌入 `locales/` 下的五份 ARB JSON 词条，英语为源语言和回退语言。使用带语义前缀的扁平小驼峰键，例如 `settingsSave` 和 `transferEtaSeconds`。文案使用完整的 ICU MessageFormat 模板和命名占位符，在英语词条的 `@message` 元数据中声明参数类型。语义不同的消息保留独立键。保留的 `mygo.*` 键提供框架菜单、控件和对话框翻译，包括从 starter 引入的日语和韩语补充文案。

同步修改五份词条后运行：

```sh
make generate
make i18n
```

随词条提交 `locales/catalog_gen.go`，不要手动编辑生成文件。直接调用生成函数，例如 `locales.SettingsSave()`、`locales.TransferEtaSeconds(seconds)`；指定语言时使用 `locales.In(tag)`。MyGo CLI 的可选 TypeScript 输出位于已忽略的 `.mygo/`，避免在文档工作区生成原生应用源码。

Rivu 保存的语言设置继续作为选择来源，设置成功保存后通过 `App.SetLocale` 应用，`system` 对应 MyGo 的空覆盖值。MyGo 按系统首选语言顺序匹配，并统一更新原生文本、控件和角色菜单。`App.OnLocaleChanged` 刷新应用菜单、托盘内容及窗口标题，也处理系统语言变化。`MYGO_LOCALE` 可临时预览语言而不保存。

阶段标签在渲染时翻译。结构化错误携带类型化的 `func(locales.Messages) string` 渲染函数及显式原始错误，在展示边界使用当前语言；`Error()` 使用英语以稳定诊断，`Unwrap()` 保留错误链。异步工作返回的展示文本不应再次翻译。词条测试检查完整翻译及未使用文案，`mygo i18n check` 校验 ICU 模板与占位符兼容性，Go 编译检查函数名和参数类型。UI 测试覆盖五种语言的正常与最小窗口，并为框架控件设置测试语言。替换全局语言提供者的测试串行执行，不写入原生语言偏好。系统对话框和第三方消息可能使用其自身语言。

## 集合性能基准

```sh
go test . ./internal/session -run '^$' -bench 'Benchmark(Main1000Tasks|FileTree10000Files|CachedFileTree10000Files|Snapshot1000Tasks10000Files)$' -benchmem
```

测量、持久化迁移和功能语义见[重构报告](refactor-validation.md)。升级引擎后必须通过本地 Tracker 集成测试，验证诊断输出契约。

原生捕获还支持 `RIVU_CAPTURE_SURFACE=settings-downloads` 或 `settings-automation`。

## 延伸阅读

- [架构](architecture.md)：会话归属、引擎集成和持久化。
- [实现状态](implementation-status.md)：已实现行为与发布门槛。
- [桌面验证](desktop.md)：原生行为与验收。
- [发布指南](releasing.md)：打包、签名、更新与许可证。
- [设计参考](design/README.md)：仅供评审的 HTML 原型。
