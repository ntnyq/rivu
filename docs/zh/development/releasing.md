# 构建与发布 Rivu

## 产品标识

- 应用：Rivu
- Bundle ID：`com.ntnyq.rivu`
- 仓库：`https://github.com/ntnyq/rivu`
- 网站：`https://rivu.ntnyq.dev`
- 初始版本：`0.0.1`
- 最低 macOS 版本：12.0

这些值已在本地配置。DNS、网站托管和 GitHub 发布是独立部署步骤。

## 本地构建

```sh
make check
make race
make integration
go tool mygo vet
make snapshots
make icon
go tool mygo build -skip-dmg
```

Apple Silicon 上的应用位于 `build/darwin-arm64/Rivu.app`。默认采用 ad-hoc 签名，不是 Developer ID 签名或公证。使用 `go tool mygo build` 生成平台安装包，添加 `-platform darwin/universal` 构建 macOS 通用包。

图标源文件为 `resources/icon.svg`。`make icon` 生成全部 PNG、`resources/rivu.icns`、`resources/rivu.iconset`、`resources/rivu.ico` 及单色模板托盘图标。不要直接修改生成的位图。

## 发布工作流

手动触发的 Release 工作流验证稳定版本，更新 `internal/app/run.go`、`mygo.json` 和变更日志，运行测试并打包 macOS，然后创建发布提交、标签和 **草稿** GitHub Release。它要求默认分支已有提交，且具备仓库写权限。不发布 Windows/Linux 安装包。公开草稿前请检查原生验收矩阵。

仅获取本地开发包时不要运行该工作流，因为它会修改仓库并创建标签。源代码中不包含签名密钥。

## 签名更新

更新协调器已实现，但生产更新源和密钥尚未配置。使用固定版本的 MyGo 发布工具，提供可信 HTTPS 更新源、对应公钥、签名产物以及 Developer ID 和公证配置。私钥应放在发布密钥存储中。`Updater.Enabled() == false` 不能解释为已经是最新版本。

自动检查与自动安装分别由用户启用。自动安装要求会话空闲，并阻止新任务准入。手动安装先说明中断，保存并停止传输，再由 MyGo 下载、校验和替换，之后恢复允许执行的任务。重启使用正常的退出与保存协调流程。

向用户开放前，使用受控签名更新验证：有效更新、错误签名、错误架构、空间不足、替换中断，以及重启前保存失败。更新失败必须保留可运行的旧应用。其他原生验收要求见[桌面验证](desktop.md)。

## 许可证

仓库尚未为 Rivu 自有代码选定许可证。引擎固定为 anacrolix/torrent v1.61.1-0.20260927071845-d913b30f520e，采用 MPL-2.0；MyGo 和传递依赖各有许可证。[第三方声明](https://github.com/ntnyq/rivu/blob/main/THIRD_PARTY_NOTICES.md)记录主要组件。分发二进制前，需要准备完整依赖许可证清单并履行源码及声明义务。
