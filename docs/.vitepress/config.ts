import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Rivu',
  description: 'A native BitTorrent downloader for macOS.',
  lastUpdated: true,
  head: [
    [
      'link',
      {
        rel: 'icon',
        type: 'image/png',
        href: '/icon.png',
      },
    ],
  ],
  locales: {
    root: {
      label: 'English',
      lang: 'en',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/usage' },
          { text: 'Development', link: '/development' },
          { text: 'Release status', link: '/implementation-status' },
        ],
        sidebar: [
          {
            text: 'Get started',
            items: [
              { text: 'Using Rivu', link: '/usage' },
              { text: 'Build and run', link: '/development' },
            ],
          },
          {
            text: 'Engineering',
            items: [
              { text: 'Architecture', link: '/architecture' },
              { text: 'Source structure', link: '/implementation-structure' },
              { text: 'Feature validation', link: '/refactor-validation' },
              { text: 'Desktop acceptance', link: '/desktop' },
              { text: 'Implementation status', link: '/implementation-status' },
              { text: 'Releasing', link: '/releasing' },
            ],
          },
          {
            text: 'Historical design',
            collapsed: true,
            items: [
              { text: 'Product specification', link: '/product-design' },
              { text: 'Implementation plan', link: '/implementation-plan' },
              {
                text: 'Proposed source tree',
                link: '/history/implementation-structure-2026-10-12',
              },
              { text: 'Visual prototype', link: '/design/README' },
            ],
          },
        ],
        editLink: {
          pattern: 'https://github.com/ntnyq/rivu/edit/main/docs/:path',
          text: 'Edit this page on GitHub',
        },
        footer: {
          message:
            'Built with Go and MyGo. Documentation powered by VitePress.',
        },
      },
    },
    zh: {
      label: '简体中文',
      lang: 'zh-Hans',
      description: '为 macOS 打造的原生 BitTorrent 下载器。',
      markdown: {
        container: {
          tipLabel: '提示',
          infoLabel: '说明',
          warningLabel: '注意',
          dangerLabel: '危险',
          detailsLabel: '详细信息',
        },
        codeCopyButton: { tooltipText: '复制代码', copiedText: '已复制' },
      },
      themeConfig: {
        nav: [
          { text: '使用指南', link: '/zh/usage' },
          { text: '开发指南', link: '/zh/development' },
          { text: '发布状态', link: '/zh/implementation-status' },
        ],
        sidebar: [
          {
            text: '开始使用',
            items: [
              { text: '使用 Rivu', link: '/zh/usage' },
              { text: '构建与运行', link: '/zh/development' },
            ],
          },
          {
            text: '工程文档',
            items: [
              { text: '架构与状态归属', link: '/zh/architecture' },
              { text: '源码结构', link: '/zh/implementation-structure' },
              { text: '功能与验证报告', link: '/zh/refactor-validation' },
              { text: '桌面行为与验收', link: '/zh/desktop' },
              { text: '实现状态', link: '/zh/implementation-status' },
              { text: '构建与发布', link: '/zh/releasing' },
            ],
          },
          {
            text: '历史设计',
            collapsed: true,
            items: [
              { text: '产品设计', link: '/zh/product-design' },
              { text: '实施计划', link: '/zh/implementation-plan' },
              {
                text: '早期目录规划',
                link: '/zh/history/implementation-structure-2026-10-12',
              },
              { text: '视觉原型', link: '/zh/design/README' },
            ],
          },
        ],
        editLink: {
          pattern: 'https://github.com/ntnyq/rivu/edit/main/docs/:path',
          text: '在 GitHub 上编辑此页',
        },
        outline: { label: '本页目录', level: [2, 3] },
        docFooter: { prev: '上一页', next: '下一页' },
        lastUpdated: {
          text: '最后更新',
          formatOptions: { dateStyle: 'medium' },
        },
        langMenuLabel: '切换语言',
        darkModeSwitchLabel: '外观',
        lightModeSwitchTitle: '切换到浅色模式',
        darkModeSwitchTitle: '切换到深色模式',
        sidebarMenuLabel: '目录',
        skipToContentLabel: '跳转到内容',
        returnToTopLabel: '返回顶部',
        navMenuLabel: '主导航',
        mobileMenuLabel: '菜单',
        extraMenuLabel: '更多选项',
        notFound: {
          title: '页面未找到',
          quote: '这个页面可能已移动，请返回首页继续浏览。',
          linkLabel: '返回首页',
          linkText: '返回首页',
        },
        footer: { message: '使用 Go 和 MyGo 构建，文档由 VitePress 驱动。' },
      },
    },
  },
  themeConfig: {
    i18nRouting(data, route, targetLocale) {
      const sourcePrefix =
        data.localeIndex.value === 'root' ? '' : `${data.localeIndex.value}/`
      const pagePath = route.data.relativePath
        .slice(sourcePrefix.length)
        .replace(/(^|\/)index\.md$/, '$1')
        .replace(/\.md$/, '.html')
      const targetPrefix = targetLocale === 'root' ? '/' : `/${targetLocale}/`

      // Translated headings have different anchors; switch to the corresponding page.
      return `${targetPrefix}${pagePath}${route.query}`
    },
    logo: { src: '/icon.png', alt: 'Rivu' },
    socialLinks: [
      {
        icon: 'github',
        link: 'https://github.com/ntnyq/rivu',
      },
    ],
    outline: { level: [2, 3] },
    search: {
      provider: 'local',
      options: {
        locales: {
          zh: {
            translations: {
              button: { buttonText: '搜索文档', buttonAriaLabel: '搜索文档' },
              modal: {
                displayDetails: '显示详细列表',
                resetButtonTitle: '清除搜索',
                backButtonTitle: '返回',
                noResultsText: '未找到相关结果',
                footer: {
                  selectText: '选择',
                  navigateText: '切换',
                  closeText: '关闭',
                },
              },
            },
          },
        },
      },
    },
  },
})
