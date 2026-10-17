---
layout: home
title: 让下载安静流动
titleTemplate: Rivu

hero:
  name: Rivu
  text: 让下载安静流动。
  tagline: 为 macOS 打造的原生 BitTorrent 下载器。选择需要的文件，整理下载队列，通过菜单栏随时掌握传输进度。
  image:
    src: /icon.png
    alt: Rivu 应用图标
  actions:
    - theme: brand
      text: 开始使用
      link: /zh/usage
    - theme: alt
      text: 构建与运行
      link: /zh/development
    - theme: alt
      text: 在 GitHub 上查看
      link: https://github.com/ntnyq/rivu

features:
  - title: 融入你的 Mac
    details: 使用 Go 和 MyGo 构建，提供原生窗口、浅色与深色外观、快捷键和菜单栏面板。应用运行时无需 WebView 或 JavaScript。
    link: /zh/desktop
  - title: 只下载所需文件
    details: 打开种子或磁力链接，预览元数据，选择文件与优先级。支持 v1、v2 和混合种子。
    link: /zh/usage#添加下载
  - title: 记住你的下载安排
    details: 控制带宽与并发数，批量暂停，并在重启后保留手动暂停。恢复传输前重新校验数据。
    link: /zh/usage#管理传输
  - title: 自主设定做种策略
    details: 下载完成后停止、达到目标分享率后停止，或不限分享率持续做种；还可限制任务累计有效做种时长。
    link: /zh/usage#管理传输
  - title: 让整理更省心
    details: 按每日时段限速，使用标签整理任务，或启用文件夹监视，自动导入新种子。
    link: /zh/usage#标签、定时限速与监视文件夹
  - title: 始终掌控下载文件
    details: 接管并校验已有数据，修改任务位置。移除任务默认保留文件，移到废纸篓需单独选择。
    link: /zh/usage#存储与恢复
---

## 下载状态，一目了然

![Rivu 浅色主窗口](../images/main-light.png)

<sub>截图由应用界面渲染器生成，展示简体中文界面与演示传输数据。</sub>

## 体验开发版本

Rivu 正在积极开发，目标平台为 **macOS 12 及以上版本**。请按照[开发指南](development.md)在本地构建与运行。正式签名、公证和可信更新源仍在准备中；Windows 与 Linux 构建目前仅用于编译检查。

## 浏览文档

- [使用指南](usage.md)：导入、队列控制、自动化与恢复。
- [架构](architecture.md)与[源码结构](implementation-structure.md)：状态归属、模块和持久化。
- [功能验证](refactor-validation.md)与[实现状态](implementation-status.md)：已验证行为和待完成的发布工作。
- [桌面验收](desktop.md)与[发布指南](releasing.md)：原生验证、打包与分发。
- [产品设计](product-design.md)、[实施计划](implementation-plan.md)、[早期目录规划](history/implementation-structure-2026-10-12.md)与[视觉原型](design/README.md)：历史参考；当前代码与实现文档优先。
