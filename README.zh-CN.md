<p align="center">
  <img src="resources/icons/rivu-128.png" alt="Rivu 图标" width="80" height="80">
</p>

<h1 align="center">Rivu</h1>

<p align="center"><a href="README.md">English</a> · 简体中文</p>

<p align="center"><strong>让下载安静流动。</strong></p>

<p align="center">
  为 macOS 打造的原生 BitTorrent 下载器。<br>
  选择需要的文件，整理下载队列，把传输交给 Rivu。
</p>

<p align="center">
  <a href="#开始使用">开始使用</a> ·
  <a href="docs/zh/usage.md">使用指南</a> ·
  <a href="https://github.com/ntnyq/rivu/issues">反馈问题</a>
</p>

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/images/main-dark.png">
  <img src="docs/images/main-light.png" alt="Rivu 下载队列，展示传输进度、速度和所选任务的详细信息">
</picture>

## 下载状态，一目了然

查看正在下载、做种、已完成或需要处理的任务。搜索与排序传输列表，打开详情面板查看进度、文件与连接信息。

- **自由添加。** 打开 `.torrent` 文件、粘贴磁力链接，或将种子拖入窗口。支持 v1、v2 和混合种子。
- **只下载所需文件。** 开始前选择文件、设置优先级与保存位置。磁力链接通过可取消的预览解析元数据。
- **按自己的节奏传输。** 限制带宽与并发下载数，暂停单个任务或整个队列。恢复批量暂停时，保留之前单独暂停的任务。
- **继续上次的进度。** 重启后保留任务与暂停意图，恢复传输时重新校验已下载数据。
- **自主设定做种策略。** 下载完成后停止、达到目标分享率后停止，或不限分享率持续做种；也可设置累计做种时长。
- **让整理更省心。** 按每日时段限速，使用标签管理任务，或监视文件夹自动导入新种子。
- **复用已有数据。** 明确接管现有下载，或修改任务位置；Rivu 校验数据，并保留已有文件的删除权限边界。
- **保留下载文件。** 移除任务默认保留文件；移到废纸篓是单独的选项。

## 关闭窗口，继续下载

Rivu 在后台运行时会继续传输。通过菜单栏面板快速查看速度与活动任务，暂停传输、切换限速或添加下载。

<p align="center">
  <img src="docs/images/menu-bar.png" alt="Rivu 菜单栏面板，包含传输进度、暂停控制与限速开关" width="360">
</p>

## 融入你的 Mac

Rivu 提供原生界面、浅色与深色外观、熟悉的菜单、键盘快捷键和命令面板。应用使用 Go 与 MyGo 构建，运行时无需 WebView 或 JavaScript。

支持 **English、简体中文、繁體中文、日本語和 한국어**。默认跟随系统语言，也可在 **设置 → 通用 → 语言** 中修改。

<details>
  <summary>查看深色外观</summary>

  <img src="docs/images/main-dark.png" alt="深色外观下的 Rivu 下载队列和传输详情">
</details>

<sub>截图由应用界面渲染器生成，展示简体中文界面与演示传输数据。</sub>

## 开始使用

Rivu 正在积极开发，目标平台为 **macOS 12 及以上版本**。请按照[构建与运行指南](docs/zh/development.md)体验当前应用。正式签名分发与自动更新仍在准备中；Windows 和 Linux 暂不提供受支持的发布版本。

启动 Rivu 后：

1. 选择 **添加下载**，打开 `.torrent` 文件或粘贴磁力链接。
2. 检查下载信息、选择保存位置，并在元数据就绪后选择文件。
3. 开始传输，通过主窗口或菜单栏查看进度。

磁力链接要求、目录处理与恢复流程见[使用指南](docs/zh/usage.md)。欢迎通过 [GitHub Issues](https://github.com/ntnyq/rivu/issues)报告问题或提出建议。

## 了解项目

[中文文档首页](docs/zh/index.md) · [English documentation](docs/index.md) · [开发指南](docs/zh/development.md) · [架构](docs/zh/architecture.md) · [实现状态](docs/zh/implementation-status.md) · [构建与发布](docs/zh/releasing.md) · [第三方声明](THIRD_PARTY_NOTICES.md)

## 文档站点

中英文文档站点基于 VitePress。安装 Node.js 22.12 或更高版本，以及 `package.json` 声明的 pnpm 版本后，在仓库根目录运行：

```sh
pnpm install
pnpm docs:dev
```

运行 `pnpm docs:build` 生成 `docs/.vitepress/dist`，使用 `pnpm docs:preview` 预览构建结果。英文文档位于 `docs/`，简体中文文档位于 `docs/zh/`。子路径部署方式见[文档开发流程](docs/zh/development.md#文档站点)。原生应用独立使用 Go 工具链构建。
