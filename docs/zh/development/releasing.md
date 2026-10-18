# 构建与发布 Rivu

## 产品标识

- 应用：Rivu
- Bundle ID：`com.ntnyq.rivu`
- 仓库：`https://github.com/ntnyq/rivu`
- 网站：`https://rivu.ntnyq.dev`
- 平台：macOS、Windows、Linux
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

各发布目标的打包命令：

```sh
go tool mygo build -platform darwin/universal
go tool mygo build -platform windows/amd64
go tool mygo build -platform linux/amd64
```

macOS 签名和 DMG 创建需要 Mac。Windows 安装程序在 Windows 上构建，MyGo 自动查找或下载 NSIS。Linux 打包生成 DEB、`.tar.gz` 压缩包和 `install.sh`。Windows 与 Linux 发布包目前面向 x64；源码构建可选择 MyGo 支持的其他架构。Windows 未安装 Make 时，直接运行对应的 `go` 命令。

图标源文件为 `resources/icon.svg`。`make icon` 生成全部 PNG、`resources/rivu.icns`、`resources/rivu.iconset`、`resources/rivu.ico` 及单色模板托盘图标。不要直接修改生成的位图。

## 发布工作流

在默认分支手动触发 Release 工作流，输入 `MAJOR.MINOR.PATCH` 格式的稳定版本，不带 `v`。工作流要求默认分支已有提交，且具备仓库写权限。

1. **准备：** 校验版本和分支，更新 `internal/app/run.go`、`mygo.json` 与 `CHANGELOG.md`，运行竞态测试、词条检查、本地 swarm 集成测试、Go vet 与 MyGo vet。将准备后的文件和 `RELEASE_NOTES.md` 传给后续任务。
2. **构建：** 在 macOS、Windows 和 Linux runner 上并行打包同一个版本，通过 `MYGO_UPDATER_PRIVATE_KEY` 签名更新包，验证并收集各平台的安装包、更新包与清单。缺少产物，或签名、大小、版本、下载地址不匹配，都会使构建失败。
3. **发布：** 所有构建成功后再次校验汇总的更新产物，生成 `SHA256SUMS`，原子推送发布提交和标签，并上传 **草稿** GitHub Release。Release Notes 与新增变更日志使用同一份文本。

| Runner                | 目标               | 发布产物                                                                      |
| --------------------- | ------------------ | ----------------------------------------------------------------------------- |
| macOS                 | `darwin/universal` | Apple Silicon 与 Intel 通用 `.dmg`                                            |
| Windows               | `windows/amd64`    | `Rivu Setup <version> amd64.exe`                                              |
| Linux（Ubuntu 24.04） | `linux/amd64`      | `rivu_<version>_amd64.deb`、`rivu-<version>-linux-amd64.tar.gz`、`install.sh` |

各平台还包含 `rivu-<version>-<target>.tar.gz` 和 `update-<target>.json`，Linux 压缩包同时用于安装与更新。汇总后的工作流产物和草稿还包括 `RELEASE_NOTES.md` 与 `SHA256SUMS`。公开草稿前请检查[原生验收矩阵与已知平台限制](desktop.md)，然后将其公开为稳定版并标为 Latest，客户端才能检查到更新。

### 面向用户的变更日志

`CHANGELOG.md` 和 GitHub Release Notes 仅收录上一个稳定标签之后的 `feat`、`fix` 和 `refactor` 提交；首次发布检查全部历史。支持 `fix(storage): ...`、`feat!: ...` 等带 scope 或破坏性变更标记的形式。合并提交以及 `chore`、`docs`、`test`、`ci`、`build`、`style`、`perf` 等其他类型均不展示。用户可感知的变更应使用这三种类型之一。

提交区间非空但没有匹配项时仍可发布，说明显示“No user-facing changes in this release.”。存在上一个稳定标签时，通过 Full Changelog 对比链接仍可查看全部提交。新增版本时保留旧版本的变更日志。

仅获取本地开发包时不要运行该工作流，因为它会修改仓库并创建标签。源代码中不包含签名密钥。

## 签名更新

`mygo.json` 已配置 `updates.github=ntnyq/rivu`、`tagPrefix=v` 与验证公钥。固定版本的 MyGo 构建会嵌入对应平台的 `https://github.com/ntnyq/rivu/releases/latest/download/update-<target>.json` 更新源。macOS 通用包在 Apple Silicon 与 Intel 上都使用 `darwin-universal`。当前采用完整签名更新包（`deltas=0`），构建时不需要下载历史版本。

仓库 Actions Secret **`MYGO_UPDATER_PRIVATE_KEY`** 必须包含匹配的私钥。工作流在准备发布前检查该密钥是否存在，仅向此检查和打包步骤提供私钥。`tools/releaseassets` 使用已提交的公钥验证更新包 SHA-256 摘要的 Ed25519 签名，校验通过才收集文件；汇总产物后、创建标签前会再次验证。缺少签名产物会阻止发布，`SHA256SUMS` 本身不能代替更新签名。

新 fork 应使用 `go tool mygo keygen` 生成专属密钥，修改 `updates.github`、`updates.publicKey` 和 `tools/releaseassets` 中的仓库地址校验，再将私钥放入该 fork 的 Actions Secret。请在 Git 外备份私钥；已发布的应用只信任内置公钥，不要随意重新生成或替换密钥。禁止将 `mygo-update.key` 提交到 Git 或放入工作流产物。

本地签名构建也通过 `MYGO_UPDATER_PRIVATE_KEY` 提供同一私钥。缺少私钥时 MyGo 仍可构建本地应用，但不生成签名更新产物，发布收集校验会失败。Apple Developer ID 签名与公证独立于更新签名，目前仍未配置。

草稿 Release 不会出现在客户端更新检查中。未配置更新源的旧版本需要先手动安装一次支持更新的发行版。开发模式与无写权限的安装会显示更新不可用；Linux DEB 安装应通过包管理器更新。`Updater.Enabled() == false` 不能解释为已经是最新版本。

自动检查与自动安装分别由用户启用。自动安装要求会话空闲，并阻止新任务准入。手动安装先说明中断，保存并停止传输，再由 MyGo 下载、校验和替换，之后恢复允许执行的任务。重启使用正常的退出与保存协调流程。

更新检查保留 45 秒的整体网络时限，最多尝试三次，每次 12 秒，两次重试分别退避 0.5 秒和 1 秒。连接与 TLS 握手各限制为 5 秒，响应头限制为 10 秒。仅对 EOF、响应截断、连接重置/中止和超时重试；取消、无效 JSON、证书错误和 404 等 HTTP 错误会停止检查。失败不会推进上次成功检查时间，也不会整体重试安装操作。

更新请求与更新包下载使用专用传输。显式 `HTTP_PROXY`/`HTTPS_PROXY` 优先，`NO_PROXY` 也可绕过 Windows 系统代理。否则 Windows 会为每个请求读取当前用户的系统代理，包括重定向后的 URL。手动代理优先于 WPAD 自动发现；显式 PAC 脚本优先于手动代理，且遵守 PAC 的 `DIRECT` 结果。PAC/WPAD 由 WinHTTP 解析，调用方最多等待 5 秒，同时最多保留一个尚未结束的原生查询。WPAD 不可用或超时时会回退直连，并暂停发现一分钟，避免家庭网络默认启用的自动检测阻止更新；手动代理或显式 PAC 的变更仍会立即生效。显式 PAC 失败会报错，不会悄悄绕过代理。MyGo v0.5.5 适配层仅为带有更新标记的上下文切换传输，保留其他应用 HTTP 流量的行为及 MyGo 原有的 HTTPS 和签名检查。CI 在现有 Windows AMD64 runner 上运行真实 MyGo EOF 回归测试，以及 Windows 系统代理/PAC 测试。

向用户开放前，使用受控签名更新验证：有效更新、错误签名、错误架构、空间不足、替换中断，以及重启前保存失败。更新失败必须保留可运行的旧应用。其他原生验收要求见[桌面验证](desktop.md)。

## 许可证

仓库尚未为 Rivu 自有代码选定许可证。引擎固定为 anacrolix/torrent v1.61.1-0.20260927071845-d913b30f520e，采用 MPL-2.0；MyGo 和传递依赖各有许可证。[第三方声明](https://github.com/ntnyq/rivu/blob/main/THIRD_PARTY_NOTICES.md)记录主要组件。分发二进制前，需要准备完整依赖许可证清单并履行源码及声明义务。
