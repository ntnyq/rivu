# 快速开始

从安装 Rivu 到第一次下载，按下面的步骤即可开始体验。

## 获取 Rivu

Rivu 支持 **macOS 12 及以上版本、Windows 10 / 11 和 Linux**。从项目的 [GitHub Releases](https://github.com/ntnyq/rivu/releases) 下载对应系统的安装包。

| 系统                     | 安装包                                                  | 安装方式                                                                           |
| ------------------------ | ------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| macOS                    | Apple Silicon 与 Intel 通用 `.dmg`                      | 打开磁盘映像，将 **Rivu.app** 拖入**应用程序**文件夹，再从该文件夹打开。           |
| Windows                  | x64 `Rivu Setup <version> amd64.exe`                    | 运行安装程序，随后从开始菜单启动 Rivu。                                            |
| Linux（Debian / Ubuntu） | x64 `rivu_<version>_amd64.deb`                          | 使用下载文件的实际名称，运行 `sudo apt install ./rivu_<version>_amd64.deb`。       |
| 其他 Linux 发行版        | x64 `rivu-<version>-linux-amd64.tar.gz` 和 `install.sh` | 将两者下载到同一目录，运行 `sh install.sh`，为当前用户安装到 `~/.local/rivu.app`。 |

Linux 原生窗口需要 GTK 3；系统托盘还需要 `libayatana-appindicator3`（Debian / Ubuntu 可安装 `libayatana-appindicator3-1`）。Rivu 使用原生 UI，窗口不需要 WebView2 或 WebKitGTK，但当前 MyGo 生成的 DEB 仍声明了 WebKitGTK 依赖。

### macOS 安装说明

当前 macOS 下载包尚未使用 Apple Developer ID 签名，也未经过 Apple 公证，因此首次打开可能被 macOS 拦截。请按下方[被 macOS 拦截时打开 Rivu](#被-macos-拦截时打开-rivu)的步骤操作。新版发行构建支持在[设置 → 更新与关于](settings.md#更新)中获取 GitHub Releases 签名更新。如果旧版本显示更新不可用，请先手动安装一次支持更新的发行版。

### 被 macOS 拦截时打开 Rivu

如果 macOS 提示无法验证开发者，或无法检查 Rivu 是否包含恶意软件，请先确认安装包来自项目的 [GitHub Releases](https://github.com/ntnyq/rivu/releases) 且未经修改，再按以下步骤操作。

弹窗可能提示 **Apple 无法检查 Rivu 是否包含恶意软件**，只提供**移到废纸篓**和**完成**，没有**打开**按钮。这表示 Apple 无法完成验证，不等于已经检测到恶意软件。请点击**完成**保留应用，再到「隐私与安全性」中放行：

1. 在**应用程序**中尝试打开 **Rivu.app**。如果弹窗显示**移到废纸篓**和**完成**，请点击**完成**，不要点击**移到废纸篓**；如果对应按钮显示为**取消**，则点击**取消**。
2. 打开 **Apple 菜单 → 系统设置 → 隐私与安全性**，向下滚动到**安全性**区域。
3. 找到 Rivu 被阻止打开的提示，点击**仍要打开**；如有要求，按提示验证身份。
4. 在再次出现的确认框中点击**打开**。macOS 会为这份 Rivu 保存例外，之后可以正常打开。

在 **macOS 12 Monterey** 上，入口为 **系统偏好设置 → 安全性与隐私 → 通用**；如有需要，先解锁面板，再点击**仍要打开**。

如果没有看到**仍要打开**，请再次尝试启动 Rivu，然后返回该面板。受管理的 Mac 可能需要管理员协助。如果提示应用**将损坏你的电脑**或**已损坏**，请停止操作并[反馈完整提示](https://github.com/ntnyq/rivu/issues)。

更多说明见 [Apple 官方的下载应用打开方法](https://support.apple.com/zh-cn/102445)。

## 从源码构建

如果希望自行构建，请安装 **Go 1.27.2 或更高版本**，然后运行：

```sh
git clone https://github.com/ntnyq/rivu.git
cd rivu
go mod download
go run .
```

需要生成当前系统的安装包时运行：

```sh
go tool mygo build
```

产物位于 `build/<os>-<arch>/`。Apple Silicon 上的应用为 `build/darwin-arm64/Rivu.app`；添加 `-skip-dmg` 可仅生成应用包。本地 macOS 构建使用临时签名。其他构建目标与打包细节见[开发指南](../development/index.md)。

## 添加第一个下载

下文快捷键使用 macOS 按键符号，Windows 和 Linux 请将 **⌘** 替换为 **Ctrl**。

1. 选择 **添加下载**，或按 **⌘N**。
2. 选择 `.torrent` 文件或粘贴磁力链接，也可将种子文件拖入主窗口。
3. 选择保存位置，点击 **读取并预览**。
4. 检查名称、大小和文件列表，至少保留一个选中文件，然后添加任务。

默认下载目录是 `~/Downloads/Rivu`。新下载请选择没有冲突文件的位置；复用已有数据请参考[接管已有文件](adding-downloads.md#复用已有文件)。

磁力链接需要明确的 Tracker 或对等节点地址才能开始解析。如果链接被拒绝或解析超时，请阅读[磁力链接预览](adding-downloads.md#预览磁力链接)。

## 查看下载进度

选择任务可查看进度、文件与连接信息。通过侧边栏筛选活动、已完成或已暂停的传输；关闭主窗口后，可在菜单栏面板查看简要信息。

关闭窗口会继续下载；**⌘Q** 会在保存会话后退出应用并停止传输。

## 调整常用设置

按 **⌘,** 打开设置，选择语言、默认保存位置与带宽限制。接下来可以了解[队列与做种管理](managing-transfers.md)，或设置[自动导入](automation.md)。
