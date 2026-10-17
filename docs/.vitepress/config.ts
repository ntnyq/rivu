import { defineConfig, type DefaultTheme } from 'vitepress'

const developmentSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: 'Developer documentation',
    items: [
      {
        text: 'Build and run',
        link: '/development/',
      },
      {
        text: 'Architecture',
        link: '/development/architecture',
      },
      {
        text: 'Source structure',
        link: '/development/implementation-structure',
      },
      {
        text: 'Feature validation',
        link: '/development/refactor-validation',
      },
      {
        text: 'Desktop acceptance',
        link: '/development/desktop',
      },
      {
        text: 'Release readiness',
        link: '/development/implementation-status',
      },
      {
        text: 'Packaging and releases',
        link: '/development/releasing',
      },
    ],
  },
  {
    text: 'Historical design',
    collapsed: true,
    items: [
      {
        text: 'About the archive',
        link: '/history/',
      },
      {
        text: 'Product specification',
        link: '/history/product-design',
      },
      {
        text: 'Implementation plan',
        link: '/history/implementation-plan',
      },
      {
        text: 'Proposed source tree',
        link: '/history/implementation-structure-2026-10-12',
      },
      {
        text: 'Visual prototype',
        link: '/design/README',
      },
    ],
  },
]

const zhDevelopmentSidebar: DefaultTheme.SidebarItem[] = [
  {
    text: '开发者文档',
    items: [
      {
        text: '构建与运行',
        link: '/zh/development/',
      },
      {
        text: '架构与状态归属',
        link: '/zh/development/architecture',
      },
      {
        text: '源码结构',
        link: '/zh/development/implementation-structure',
      },
      {
        text: '功能验证',
        link: '/zh/development/refactor-validation',
      },
      {
        text: '桌面验收',
        link: '/zh/development/desktop',
      },
      {
        text: '发布状态',
        link: '/zh/development/implementation-status',
      },
      {
        text: '构建与发布',
        link: '/zh/development/releasing',
      },
    ],
  },
  {
    text: '历史设计',
    collapsed: true,
    items: [
      {
        text: '归档说明',
        link: '/zh/history/',
      },
      {
        text: '产品设计',
        link: '/zh/history/product-design',
      },
      {
        text: '实施计划',
        link: '/zh/history/implementation-plan',
      },
      {
        text: '早期目录规划',
        link: '/zh/history/implementation-structure-2026-10-12',
      },
      {
        text: '视觉原型',
        link: '/zh/design/README',
      },
    ],
  },
]

export default defineConfig({
  title: 'Rivu',
  description:
    'A native BitTorrent app for macOS, Windows, and Linux. Choose your files, manage your queue, and keep downloading in the background.',
  lastUpdated: true,
  sitemap: { hostname: 'https://rivu.ntnyq.dev' },
  transformHead({ pageData }) {
    const pagePath = pageData.relativePath
      .replace(/index\.md$/, '')
      .replace(/\.md$/, '.html')
    const url = new URL(pagePath, 'https://rivu.ntnyq.dev/').href

    return [
      ['link', { rel: 'canonical', href: url }],
      ['meta', { property: 'og:url', content: url }],
      ['meta', { property: 'og:type', content: 'website' }],
      [
        'meta',
        {
          property: 'og:title',
          content: pageData.title ? `${pageData.title} | Rivu` : 'Rivu',
        },
      ],
      ['meta', { property: 'og:description', content: pageData.description }],
      [
        'meta',
        { property: 'og:image', content: 'https://rivu.ntnyq.dev/icon.png' },
      ],
    ]
  },
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
          {
            text: 'Guide',
            link: '/guide/',
            activeMatch: '^/guide/(?!faq)',
          },
          {
            text: 'FAQ',
            link: '/guide/faq',
          },
          {
            text: 'Developers',
            link: '/development/',
            activeMatch: '/(development|history|design)/',
          },
        ],
        sidebar: {
          '/guide/': [
            {
              text: 'Get started',
              items: [
                {
                  text: 'Meet Rivu',
                  link: '/guide/',
                },
                {
                  text: 'Your first download',
                  link: '/guide/getting-started',
                },
              ],
            },
            {
              text: 'Using Rivu',
              items: [
                {
                  text: 'Add downloads',
                  link: '/guide/adding-downloads',
                },
                {
                  text: 'Manage transfers',
                  link: '/guide/managing-transfers',
                },
                {
                  text: 'Automation',
                  link: '/guide/automation',
                },
                {
                  text: 'Settings & shortcuts',
                  link: '/guide/settings',
                },
              ],
            },
            {
              text: 'Help',
              items: [
                {
                  text: 'FAQ',
                  link: '/guide/faq',
                },
                {
                  text: 'Troubleshooting',
                  link: '/guide/troubleshooting',
                },
              ],
            },
          ],
          '/development/': developmentSidebar,
          '/history/': developmentSidebar,
          '/design/': developmentSidebar,
        },
        editLink: {
          pattern: 'https://github.com/ntnyq/rivu/edit/main/docs/:path',
          text: 'Edit this page on GitHub',
        },
        footer: {
          message: 'A native BitTorrent app for macOS, Windows, and Linux.',
        },
      },
    },
    zh: {
      label: '简体中文',
      lang: 'zh-Hans',
      description: '支持 macOS、Windows 和 Linux 的原生 BitTorrent 下载器。',
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
          {
            text: '使用指南',
            link: '/zh/guide/',
            activeMatch: '^/zh/guide/(?!faq)',
          },
          {
            text: '常见问题',
            link: '/zh/guide/faq',
          },
          {
            text: '开发者',
            link: '/zh/development/',
            activeMatch: '/zh/(development|history|design)/',
          },
        ],
        sidebar: {
          '/zh/guide/': [
            {
              text: '开始使用',
              items: [
                {
                  text: '认识 Rivu',
                  link: '/zh/guide/',
                },
                {
                  text: '快速开始',
                  link: '/zh/guide/getting-started',
                },
              ],
            },
            {
              text: '使用 Rivu',
              items: [
                {
                  text: '添加下载',
                  link: '/zh/guide/adding-downloads',
                },
                {
                  text: '管理传输',
                  link: '/zh/guide/managing-transfers',
                },
                {
                  text: '自动化',
                  link: '/zh/guide/automation',
                },
                {
                  text: '设置与快捷键',
                  link: '/zh/guide/settings',
                },
              ],
            },
            {
              text: '帮助',
              items: [
                {
                  text: '常见问题',
                  link: '/zh/guide/faq',
                },
                {
                  text: '故障排查',
                  link: '/zh/guide/troubleshooting',
                },
              ],
            },
          ],
          '/zh/development/': zhDevelopmentSidebar,
          '/zh/history/': zhDevelopmentSidebar,
          '/zh/design/': zhDevelopmentSidebar,
        },
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
        footer: {
          message: '支持 macOS、Windows 和 Linux 的原生 BitTorrent 应用。',
        },
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
