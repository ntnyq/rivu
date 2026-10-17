# 常见问题

## 支持哪些平台？

Rivu 支持 **macOS 12 及以上版本、Windows 10 / 11 和 Linux**。发布产物包括 macOS 通用 DMG、Windows x64 安装程序，以及 Linux x64 DEB 和压缩包。Linux 需要 GTK 3，托盘还需要 AppIndicator。详见[安装说明](getting-started.md#获取-rivu)。

## 在哪里下载？

从项目的 [GitHub Releases](https://github.com/ntnyq/rivu/releases) 下载，也可以[从源码构建](getting-started.md#从源码构建)。当前 macOS 下载包尚未使用 Apple Developer ID 签名，也未经过 Apple 公证；首次打开被拦截时，请参考[被 macOS 拦截时打开 Rivu](getting-started.md#被-macos-拦截时打开-rivu)。新版发行构建支持从 GitHub Releases 获取签名更新，使用方式见[更新设置](settings.md#更新)。文档站地址为 [rivu.ntnyq.dev](https://rivu.ntnyq.dev/zh/)。

## 可以下载哪些内容？

支持 v1、纯 v2 和混合 BitTorrent 种子的 `.torrent` 文件与磁力链接。当前功能不包括 HTTP 直链下载、RSS 订阅、远程控制和流媒体播放。

## 关闭窗口会停止下载吗？

不会。只要 Rivu 仍在运行，传输就会继续。可以从菜单栏查看进度或重新打开主窗口；退出应用则会保存会话并停止传输。

## 移除任务会删除文件吗？

默认保留文件。移到废纸篓需要单独选择，且仅适用于该任务拥有的文件。接管的已有文件在移除任务时始终保留。

## 为什么重启后要检查文件？

Rivu 会先校验实际数据再继续。保存的百分比无法证明文件完整且未被修改，因此大体积下载可能需要一段校验时间才能恢复传输。

## 为什么跳过的文件仍出现在磁盘上？

种子分片可能横跨多个文件。完成相邻的选中文件时，可能需要写入跳过文件中的部分数据。

## 支持哪些语言？

应用支持英语、简体中文、繁体中文、日语和韩语，可在 **设置 → 通用 → 语言** 中切换。文档提供英语与简体中文版本。

## 如何反馈问题？

先阅读[故障排查](troubleshooting.md)并搜索已有 [GitHub Issues](https://github.com/ntnyq/rivu/issues)，再附上 操作系统与版本、Rivu 版本、预期行为与复现步骤。提交诊断信息前，请去除 Tracker 凭据与私有路径。
