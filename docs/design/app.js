/* Review-only prototype. No network requests, file writes, or system changes. */
;(() => {
  const $ = (selector, root = document) => root.querySelector(selector)
  const $$ = (selector, root = document) => [...root.querySelectorAll(selector)]
  const paths = {
    download:
      '<path d="M12 3v12m-5-5 5 5 5-5M4 15v4a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-4"/>',
    'arrow-down': '<path d="M12 4v16m-5-5 5 5 5-5"/>',
    'arrow-up': '<path d="M12 20V4m-5 5 5-5 5 5"/>',
    plus: '<path d="M12 5v14M5 12h14"/>',
    pause: '<path d="M8 5v14M16 5v14"/>',
    play: '<path d="m8 5 11 7-11 7Z"/>',
    search: '<circle cx="10.5" cy="10.5" r="6.5"/><path d="m16 16 5 5"/>',
    file: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Zm0 0v6h6M8 13h8M8 17h5"/>',
    folder:
      '<path d="M3 6a2 2 0 0 1 2-2h5l2 3h7a2 2 0 0 1 2 2v10a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2Z"/>',
    grid: '<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><rect x="14" y="14" width="7" height="7" rx="1.5"/>',
    check: '<circle cx="12" cy="12" r="9"/><path d="m8 12 3 3 5-6"/>',
    alert:
      '<path d="m10.3 4-8 14a2 2 0 0 0 1.7 3h16a2 2 0 0 0 1.7-3l-8-14a2 2 0 0 0-3.4 0ZM12 9v4m0 4h.01"/>',
    settings:
      '<path d="m9 3-1 3-3 1-2 3 2 2-1 3 3 2 3-1 2 3 3-1 1-3 3-1 2-3-2-2 1-3-3-2-3 1-2-3Z"/><circle cx="11.5" cy="11" r="3"/>',
    drive:
      '<path d="m5 4-3 12h20L19 4Z"/><rect x="2" y="16" width="20" height="5" rx="1"/><path d="M17 18.5h2"/>',
    sidebar:
      '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M9 4v16"/>',
    panel:
      '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M15 4v16"/>',
    trash: '<path d="M3 6h18M9 6V3h6v3M5 6l1 15h12l1-15M10 10v7m4-7v7"/>',
    movie:
      '<rect x="3" y="3" width="18" height="18" rx="2"/><path d="M7 3v18M17 3v18M3 8h4m-4 8h4M17 8h4m-4 8h4"/>',
    archive:
      '<rect x="3" y="4" width="18" height="5" rx="1"/><path d="M5 9v12h14V9M10 13h4"/>',
    gauge:
      '<path d="M4 18a9 9 0 1 1 16 0M12 13l5-6"/><circle cx="12" cy="14" r="2"/>',
    moon: '<path d="M20 15A9 9 0 0 1 9 3a9 9 0 1 0 11 12Z"/>',
    sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2m0 16v2M2 12h2m16 0h2M5 5l1.5 1.5m11 11L19 19M5 19l1.5-1.5m11-11L19 5"/>',
    wifi: '<path d="M2 8a17 17 0 0 1 20 0M5 12a11 11 0 0 1 14 0M8.5 16a5 5 0 0 1 7 0M12 20h.01"/>',
    battery:
      '<rect x="2" y="6" width="17" height="12" rx="2"/><path d="M22 10v4M5 9h9v6H5Z"/>',
    x: '<path d="m6 6 12 12M6 18 18 6"/>',
    external:
      '<path d="M14 3h7v7m0-7L10 14M10 3H4a1 1 0 0 0-1 1v16a1 1 0 0 0 1 1h16a1 1 0 0 0 1-1v-6"/>',
    bell: '<path d="M18 8a6 6 0 0 0-12 0c0 7-3 8-3 9h18c0-1-3-2-3-9M9 21h6"/>',
    refresh:
      '<path d="M20 8a9 9 0 0 0-15-3L2 8m0-6v6h6M4 16a9 9 0 0 0 15 3l3-3m0 6v-6h-6"/>',
    globe:
      '<circle cx="12" cy="12" r="9"/><ellipse cx="12" cy="12" rx="4" ry="9"/><path d="M3 12h18"/>',
    clock: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
  }
  const icon = name =>
    `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.65" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${paths[name] || paths.file}</svg>`
  const icons = () =>
    $$('[data-icon]').forEach(el => {
      el.innerHTML = icon(el.dataset.icon)
    })
  const escape = value =>
    String(value).replace(
      /[&<>"']/g,
      char =>
        ({
          '&': '&amp;',
          '<': '&lt;',
          '>': '&gt;',
          '"': '&quot;',
          "'": '&#39;',
        })[char],
    )
  const initial = [
    {
      id: 1,
      name: 'Ubuntu 24.04.1 Desktop',
      file: 'ubuntu-24.04.1-desktop-amd64.iso',
      type: 'file',
      state: 'downloading',
      progress: 62.4,
      total: 5.9,
      down: 8.4,
      up: 0.42,
      eta: '约 4 分钟',
      peers: '24 / 68',
      uploaded: '1.55 GB',
      ratio: '0.42',
      date: '今天 10:16',
      path: '~/Downloads/Rivu',
      order: 6,
    },
    {
      id: 2,
      name: 'Blender Open Movies',
      file: 'Blender Open Movies',
      type: 'movie',
      state: 'downloading',
      progress: 38.2,
      total: 12.4,
      down: 4.2,
      up: 0.33,
      eta: '约 30 分钟',
      peers: '18 / 42',
      uploaded: '320 MB',
      ratio: '0.07',
      date: '今天 09:52',
      path: '~/Movies/Open Movies',
      order: 5,
    },
    {
      id: 3,
      name: 'Debian 12.7.0 arm64',
      file: 'debian-12.7.0-arm64-netinst.iso',
      type: 'file',
      state: 'seeding',
      progress: 100,
      total: 0.65,
      down: 0,
      up: 0.45,
      eta: '分享率 1.24 / 2.0',
      peers: '8 / 21',
      uploaded: '806 MB',
      ratio: '1.24',
      date: '今天 08:41',
      path: '~/Downloads/Rivu',
      order: 4,
    },
    {
      id: 4,
      name: 'Sintel · Open Movie',
      file: 'Sintel_1080p.mp4',
      type: 'movie',
      state: 'completed',
      progress: 100,
      total: 1.28,
      down: 0,
      up: 0,
      eta: '已达到做种目标',
      peers: '0 / 0',
      uploaded: '2.56 GB',
      ratio: '2.00',
      date: '昨天 19:30',
      path: '~/Movies/Open Movies',
      order: 3,
    },
    {
      id: 5,
      name: 'LibreOffice 24.8 Archive',
      file: 'LibreOffice_24.8_archive.tar',
      type: 'archive',
      state: 'paused',
      resume: 'downloading',
      progress: 21.6,
      total: 0.98,
      down: 0,
      up: 0,
      baseDown: 2.3,
      eta: '等待继续',
      peers: '0 / 0',
      uploaded: '12 MB',
      ratio: '0.06',
      date: '昨天 18:12',
      path: '~/Downloads/Rivu',
      order: 2,
    },
    {
      id: 6,
      name: 'Fedora Workstation 40',
      file: 'Fedora-Workstation-Live-x86_64-40.iso',
      type: 'file',
      state: 'error',
      progress: 14.8,
      total: 2.14,
      down: 0,
      up: 0,
      eta: '需要释放 1.82 GB',
      peers: '0 / 0',
      uploaded: '18 MB',
      ratio: '0.06',
      date: '昨天 17:05',
      path: '/Volumes/Archive/Downloads',
      order: 1,
    },
  ]
  const filterDefs = [
    ['all', '全部任务', 'grid'],
    ['downloading', '下载中', 'arrow-down'],
    ['seeding', '做种中', 'arrow-up'],
    ['completed', '已完成', 'check'],
    ['paused', '已暂停', 'pause'],
    ['error', '需要处理', 'alert'],
  ]
  const stateNames = {
    downloading: '下载中',
    seeding: '做种中',
    completed: '已完成',
    paused: '已暂停',
    error: '空间不足',
    queued: '排队中',
    metadata: '获取元数据',
  }
  let tasks = structuredClone(initial)
  let selected = 1
  let filter = 'all'
  let detailTab = 'overview'
  let limited = false
  let batchPaused = []
  let emptyScene = false
  let settingsTab = 'general'
  let metadataReady = false
  let updateState = 'available'
  let toastTimer
  const prefs = {
    theme: 'light',
    menubar: true,
    login: false,
    background: true,
    resume: true,
    automatic: true,
    install: false,
    completeNotice: true,
    errorNotice: true,
    dht: true,
    pex: true,
    concurrent: '3',
    ratio: '2.0',
    port: '51413',
    directory: '~/Downloads/Rivu',
  }
  const sampleMagnet =
    'magnet:?xt=urn:btih:0000000000000000000000000000000000000001&dn=Big%20Buck%20Bunny'
  const visibleTasks = () => (emptyScene ? [] : tasks)
  const active = task =>
    ['downloading', 'seeding', 'metadata', 'queued'].includes(task.state)
  const matchFilter = (task, key) =>
    key === 'all'
    || (key === 'completed'
      ? ['completed', 'seeding'].includes(task.state)
      : task.state === key)
  const current = () => visibleTasks().find(task => task.id === selected)
  const effectiveRates = task => {
    const totalDown = visibleTasks().reduce(
      (sum, t) => sum + (t.state === 'downloading' ? t.down : 0),
      0,
    )
    const totalUp = visibleTasks().reduce(
      (sum, t) => sum + (active(t) ? t.up : 0),
      0,
    )
    return {
      down: task.down * (limited && totalDown > 2 ? 2 / totalDown : 1),
      up: task.up * (limited && totalUp > 0.512 ? 0.512 / totalUp : 1),
    }
  }
  const size = value =>
    value < 1 ? `${Math.round(value * 1000)} MB` : `${value.toFixed(2)} GB`
  const rate = value =>
    value >= 1
      ? `${value.toFixed(1)} MB/s`
      : value > 0
        ? `${Math.round(value * 1000)} KB/s`
        : '0 KB/s'
  const totals = () => {
    const list = visibleTasks()
    return {
      down: Math.min(
        list.reduce((v, t) => v + (t.state === 'downloading' ? t.down : 0), 0),
        limited ? 2 : Infinity,
      ),
      up: Math.min(
        list.reduce((v, t) => v + (active(t) ? t.up : 0), 0),
        limited ? 0.512 : Infinity,
      ),
    }
  }
  const toast = message => {
    clearTimeout(toastTimer)
    $('#toast').textContent = message
    $('#toast').hidden = false
    toastTimer = setTimeout(() => {
      $('#toast').hidden = true
    }, 3200)
  }
  const progress = (task, extra = '') =>
    `<div class="progress-track ${extra}" role="progressbar" aria-label="${escape(task.name)} 下载进度" aria-valuemin="0" aria-valuemax="100" aria-valuenow="${task.progress}"><i style="width:${task.progress}%"></i></div>`

  function renderFilters() {
    $('#filters').innerHTML = filterDefs
      .map(
        ([key, label, mark]) =>
          `<button class="filter ${key === filter ? 'active' : ''}" data-filter="${key}" ${key === filter ? 'aria-current="page"' : ''}>${icon(mark)}<span>${label}</span><span class="count">${visibleTasks().filter(t => matchFilter(t, key)).length}</span></button>`,
      )
      .join('')
  }
  function renderList() {
    const query = $('#search').value.trim().toLowerCase()
    const list = visibleTasks().filter(
      t => matchFilter(t, filter) && t.name.toLowerCase().includes(query),
    )
    const sort = $('#sort').value
    list.sort((a, b) =>
      sort === 'name'
        ? a.name.localeCompare(b.name)
        : sort === 'progress'
          ? b.progress - a.progress
          : b.order - a.order,
    )
    if (!list.some(t => t.id === selected)) selected = list[0]?.id ?? null
    $('#view-title').textContent = filterDefs.find(([key]) => key === filter)[1]
    $('#view-count').textContent = `${list.length} 个任务`
    $('#list-hint').hidden = !list.length
    $('#task-list').innerHTML = list.length
      ? list
          .map(
            task =>
              `<article class="task-row ${task.state} ${task.id === selected ? 'selected' : ''}"><button class="task-select" data-select="${task.id}" aria-label="查看 ${escape(task.name)}，${stateNames[task.state]}，${task.progress}%" aria-pressed="${task.id === selected}"></button><div class="file-symbol ${task.type}">${icon(task.type)}</div><div class="task-info"><div class="task-top"><strong title="${escape(task.name)}">${escape(task.name)}</strong><span class="percent">${task.progress.toFixed(task.progress === 100 ? 0 : 1)}%</span></div><div class="task-sub"><span>${size((task.total * task.progress) / 100)} / ${size(task.total)}</span><span class="state">${stateNames[task.state]}</span></div>${progress(task)}<div class="task-meta"><span class="${task.state === 'downloading' ? 'rate' : ''}">${task.state === 'downloading' ? `↓ ${rate(effectiveRates(task).down)}　↑ ${rate(effectiveRates(task).up)}` : task.state === 'seeding' ? `↑ ${rate(effectiveRates(task).up)}` : task.state === 'error' ? '下载磁盘空间不足' : task.state === 'paused' ? '任务已暂停' : task.state === 'queued' ? '等待空闲下载位置' : '文件已准备就绪'}</span><span>${task.eta}</span></div></div><button class="icon-button row-control" data-toggle="${task.id}" aria-label="${task.state === 'paused' ? '继续' : task.state === 'error' ? '处理' : task.state === 'completed' ? '查看' : '暂停'} ${escape(task.name)}">${icon(task.state === 'paused' ? 'play' : task.state === 'error' ? 'alert' : task.state === 'completed' ? 'folder' : 'pause')}</button></article>`,
          )
          .join('')
      : `<div class="empty-state">${icon(query ? 'search' : filter === 'error' ? 'check' : 'download')}<h2>${query ? '没有找到匹配的任务' : filter === 'all' ? '让下一个下载，从这里开始' : `暂无${filter === 'error' ? '需要处理的' : filterDefs.find(([key]) => key === filter)[1]}任务`}</h2><p>${query ? '试试其他名称，或清除搜索条件。' : '添加种子文件或磁力链接。<br>关上窗口，也能安心继续下载。'}</p><button class="primary" data-action="${query ? 'clear-search' : 'add'}">${query ? '清除搜索' : '添加下载'}</button></div>`
    const task = current()
    $('#pause-selected').disabled =
      !task || ['error', 'completed'].includes(task.state)
    $('#pause-selected').innerHTML = icon(
      task?.state === 'paused' ? 'play' : 'pause',
    )
    $('#pause-selected').setAttribute(
      'aria-label',
      task?.state === 'paused' ? '继续所选任务' : '暂停所选任务',
    )
    $('#remove-trigger').disabled = !task
  }
  function renderInspector() {
    const task = current()
    if (!task) {
      $('#inspector').innerHTML =
        `<div class="empty-state">${icon('panel')}<p>选择一个任务<br>查看下载详情</p></div>`
      return
    }
    let content
    if (detailTab === 'files') {
      content = `<p class="detail-heading">已选择全部文件</p><div class="detail-files">${(task.type
        === 'movie' && task.id === 2
        ? [
            ['Big Buck Bunny.mp4', '4.2 GB'],
            ['Sintel.mp4', '3.8 GB'],
            ['Tears of Steel.mp4', '4.4 GB'],
          ]
        : [[task.file, size(task.total)]]
      )
        .map(
          ([name, bytes]) =>
            `<label><input type="checkbox" checked data-demo-file="${escape(name)}"><span>${escape(name)}<small>${bytes} · ${task.progress}%</small></span></label>`,
        )
        .join(
          '',
        )}</div><p class="seeding-note">文件选择与优先级在此调整。此处演示文件列表，未连接下载引擎。</p>`
    } else if (detailTab === 'connections') {
      content = `<p class="detail-heading">已连接节点</p><p class="detail-percentage">${task.peers.split(' / ')[0]}<small>个</small></p><p class="seeding-note">Tracker 和节点状态均为示例。</p><hr style="border:0;border-top:1px solid var(--line);margin:20px 0"><div class="connection-row">tracker.example.org:6969<span>${active(task) ? '✓ 连接正常 · 18 分钟后通告' : '未连接'}</span></div><div class="connection-row">DHT / PEX<span>${active(task) ? '节点发现已启用' : '任务未运行'}</span></div>`
    } else {
      content = `${task.state === 'error' ? '<div class="error-box"><strong>下载磁盘空间不足</strong><br>Archive 还需要 1.82 GB。请释放空间，或更换保存位置。<button data-action="fix-path">更换保存位置…</button></div>' : ''}<p class="detail-heading">下载进度</p><p class="detail-percentage">${task.progress.toFixed(1)}<small>%</small></p>${progress(task, 'detail-progress')}<p class="detail-caption">${size((task.total * task.progress) / 100)} / ${size(task.total)}</p><div class="detail-speed"><div><span>↓ 下载速度</span><strong class="download">${rate(effectiveRates(task).down).split(' ')[0]} <small>${rate(effectiveRates(task).down).split(' ')[1]}</small></strong></div><div><span>↑ 上传速度</span><strong>${rate(effectiveRates(task).up).split(' ')[0]} <small>${rate(effectiveRates(task).up).split(' ')[1]}</small></strong></div></div><dl class="details"><div><dt>预计剩余</dt><dd>${task.state === 'downloading' ? task.eta : '—'}</dd></div><div><dt>已上传</dt><dd>${task.uploaded}</dd></div><div><dt>分享率</dt><dd>${task.ratio}</dd></div><div><dt>连接节点</dt><dd>${task.peers}</dd></div><div><dt>添加时间</dt><dd>${task.date}</dd></div></dl><div class="location"><p class="detail-heading">保存位置</p><button data-action="reveal">${icon('folder')}<span>${escape(task.path)}</span><span class="reveal-arrow">${icon('external')}</span></button></div><p class="seeding-note">下载完成后继续做种<br>分享率达到 2.0 时自动停止</p>`
    }
    $('#inspector').innerHTML =
      `<div class="inspector-file"><div class="file-symbol ${task.type}">${icon(task.type)}</div><span class="status-pill ${task.state}">${stateNames[task.state]}</span></div><h2>${escape(task.name)}</h2><p class="file-subtitle">${task.type === 'movie' ? '开放电影' : task.type === 'archive' ? '开源软件归档' : '系统镜像'} · ${size(task.total)}</p><div class="detail-tabs" role="tablist" aria-label="任务详情">${[
        ['overview', '概览'],
        ['files', '文件'],
        ['connections', '连接'],
      ]
        .map(
          ([key, label]) =>
            `<button role="tab" id="detail-${key}" aria-controls="detail-content" aria-selected="${detailTab === key}" data-detail="${key}">${label}</button>`,
        )
        .join(
          '',
        )}</div><div id="detail-content" role="tabpanel" aria-labelledby="detail-${detailTab}">${content}</div>`
  }
  function renderStatus() {
    const total = totals()
    $('#total-down').textContent = rate(total.down)
    $('#total-up').textContent = rate(total.up)
    $('#limit-trigger').setAttribute('aria-pressed', limited)
    $('#limit-label').textContent = limited ? '限速已开启' : '限速模式'
  }
  function renderTray() {
    if ($('#tray-panel').hidden) return
    const total = totals()
    const list = visibleTasks()
      .filter(t => active(t) || t.state === 'paused')
      .slice(0, 4)
    const errors = visibleTasks().filter(t => t.state === 'error').length
    $('#tray-panel').innerHTML =
      `<div class="tray-head"><strong>Rivu</strong><span>${visibleTasks().filter(active).length} 个任务进行中</span></div><div class="tray-speed"><div><span>${icon('arrow-down')}下载</span><strong>${rate(total.down).split(' ')[0]}<small>${rate(total.down).split(' ')[1]}</small></strong></div><div><span>${icon('arrow-up')}上传</span><strong>${rate(total.up).split(' ')[0]}<small>${rate(total.up).split(' ')[1]}</small></strong></div></div>${list.length ? list.map(t => `<div class="tray-task"><button class="tray-task-main" data-open-task="${t.id}"><strong>${escape(t.name)}</strong>${progress(t)}<p><span>${t.progress}% · ${stateNames[t.state]}</span><span>${t.state === 'downloading' ? rate(effectiveRates(t).down) : t.state === 'seeding' ? '↑ ' + rate(effectiveRates(t).up) : '—'}</span></p></button><button class="icon-button" data-toggle="${t.id}" aria-label="${t.state === 'paused' ? '继续' : '暂停'} ${escape(t.name)}">${icon(t.state === 'paused' ? 'play' : 'pause')}</button></div>`).join('') : '<div class="tray-empty">暂无活跃任务</div>'}${errors ? `<button class="tray-alert" data-action="errors">${icon('alert')}${errors} 个任务需要处理</button>` : ''}<div class="tray-controls"><button class="secondary" data-action="batch">${icon(batchPaused.length ? 'play' : 'pause')}${batchPaused.length ? '恢复刚才暂停的任务' : '全部暂停'}</button><button class="secondary" data-action="limit">${icon('gauge')}${limited ? '关闭限速' : '限速模式'}</button></div><div class="tray-bottom"><button data-action="open-main">打开主窗口 <span style="color:var(--muted)">⌘ 0</span></button><button class="icon-button" data-action="add" aria-label="添加下载">${icon('plus')}</button><button class="icon-button" data-action="settings" aria-label="打开设置">${icon('settings')}</button></div>`
  }
  function render() {
    renderFilters()
    renderList()
    renderInspector()
    renderStatus()
    renderTray()
  }
  function showMain() {
    $('#main-window').hidden = false
    $('#closed-window').hidden = true
  }
  function refreshSceneNav() {
    const name = $('#settings-dialog').open
      ? 'settings'
      : $('#add-dialog').open
        ? 'add'
        : !$('#tray-panel').hidden
          ? 'tray'
          : 'main'
    $$('.review-nav button').forEach(button =>
      button.classList.toggle('active', button.dataset.scene === name),
    )
  }
  function closeTray() {
    $('#tray-panel').hidden = true
    $('#tray-trigger').setAttribute('aria-expanded', 'false')
    refreshSceneNav()
  }
  function positionTray() {
    if ($('#tray-panel').hidden) return
    const desktop = $('#desktop').getBoundingClientRect()
    const trigger = $('#tray-trigger').getBoundingClientRect()
    const panel = $('#tray-panel')
    const center = trigger.left + trigger.width / 2 - desktop.left
    const left = Math.max(
      8,
      Math.min(
        center - panel.offsetWidth + 28,
        desktop.width - panel.offsetWidth - 8,
      ),
    )
    panel.style.left = `${left}px`
    panel.style.right = 'auto'
    panel.style.setProperty(
      '--arrow-right',
      `${panel.offsetWidth - (center - left) - 5}px`,
    )
  }
  function toggleTray(force) {
    if (force === true || $('#tray-panel').hidden) {
      $('#tray-panel').hidden = false
      $('#tray-trigger').setAttribute('aria-expanded', 'true')
      renderTray()
      positionTray()
      refreshSceneNav()
    } else closeTray()
  }
  function toggleTask(id) {
    const task = tasks.find(t => t.id === id)
    if (!task) return
    if (task.state === 'error') {
      selectTask(id)
      toast('请在任务详情中更换保存位置')
      return
    }
    if (task.state === 'completed') {
      toast('演示：在 Finder 中显示已下载文件')
      return
    }
    if (task.state === 'paused') {
      task.state = task.resume || 'downloading'
      task.down = task.state === 'downloading' ? task.baseDown || 2.3 : 0
      task.up = task.baseUp || 0
      task.eta =
        task.state === 'queued'
          ? '等待空闲下载位置'
          : task.state === 'seeding'
            ? `分享率 ${task.ratio} / 2.0`
            : task.baseEta || '正在估算…'
      batchPaused = batchPaused.filter(i => i !== id)
    } else {
      task.resume = task.state
      task.baseDown = task.down
      task.baseUp = task.up
      task.baseEta = task.eta
      task.state = 'paused'
      task.down = task.up = 0
      task.eta = '等待继续'
    }
    render()
  }
  function selectTask(id) {
    selected = id
    renderList()
    renderInspector()
  }
  function batch() {
    if (batchPaused.length) {
      const ids = [...batchPaused]
      batchPaused = []
      ids.forEach(id => {
        if (tasks.find(t => t.id === id)?.state === 'paused') toggleTask(id)
      })
      toast('已恢复刚才暂停的任务')
    } else {
      const ids = visibleTasks()
        .filter(active)
        .map(t => t.id)
      ids.forEach(toggleTask)
      batchPaused = ids
      toast(ids.length ? `已暂停 ${ids.length} 个任务` : '没有需要暂停的任务')
    }
    render()
  }
  function toggleLimit() {
    limited = !limited
    render()
    toast(
      limited ? '限速模式：下载 2 MB/s，上传 512 KB/s' : '已恢复常规带宽设置',
    )
  }
  function openDialog(id) {
    closeTray()
    $$('dialog[open]').forEach(d => d.close())
    $(id).showModal()
    refreshSceneNav()
  }
  function openAdd() {
    $('#add-error').hidden = true
    openDialog('#add-dialog')
  }
  function setMetadata(value) {
    metadataReady = value
    $('#metadata').hidden = !value
    updateSelectedSize()
  }
  function updateSelectedSize() {
    const files = $$('input[name=file]:checked')
    const total = files.reduce((sum, input) => sum + Number(input.value), 0)
    $('#selected-size').textContent = metadataReady
      ? `已选择 ${files.length} 个文件 · ${size(total)}　可用空间 128 GB`
      : '可用空间 128 GB'
  }

  function setTheme(theme) {
    prefs.theme = theme
    const dark =
      theme === 'dark'
      || (theme === 'system'
        && matchMedia('(prefers-color-scheme: dark)').matches)
    document.body.dataset.theme = dark ? 'dark' : 'light'
    $('#theme-label').textContent = dark ? '浅色' : '深色'
    $('#theme-toggle').innerHTML =
      icon(dark ? 'sun' : 'moon')
      + `<span id="theme-label">${dark ? '浅色' : '深色'}</span>`
    $('#theme-toggle').setAttribute(
      'aria-label',
      dark ? '切换浅色外观' : '切换深色外观',
    )
  }
  const settingSwitch = (key, title, description = '', disabled = false) =>
    `<label class="setting-row"><span>${title}${description ? `<small>${description}</small>` : ''}</span><input class="switch" type="checkbox" data-pref="${key}" aria-label="${title}" ${prefs[key] ? 'checked' : ''} ${disabled ? 'disabled' : ''}></label>`
  function renderSettings() {
    const tabs = [
      ['general', '通用', 'settings'],
      ['downloads', '下载与做种', 'download'],
      ['network', '连接', 'globe'],
      ['notifications', '通知', 'bell'],
      ['updates', '更新与关于', 'refresh'],
    ]
    $('#settings-nav').innerHTML = tabs
      .map(
        ([id, title, mark]) =>
          `<button class="${settingsTab === id ? 'active' : ''}" data-setting-tab="${id}" ${settingsTab === id ? 'aria-current="page"' : ''}>${icon(mark)}${title}</button>`,
      )
      .join('')
    let content
    if (settingsTab === 'general')
      content = `<h3>通用</h3><div class="setting-group"><label class="setting-row"><span>外观</span><select data-pref="theme" aria-label="外观">${[
        ['system', '跟随系统'],
        ['light', '浅色'],
        ['dark', '深色'],
      ]
        .map(
          ([v, l]) =>
            `<option value="${v}" ${prefs.theme === v ? 'selected' : ''}>${l}</option>`,
        )
        .join(
          '',
        )}</select></label>${settingSwitch('menubar', '在菜单栏显示 Rivu', '随时查看下载进度')}</div><div class="setting-group">${settingSwitch('login', '登录时启动 Rivu')}${settingSwitch('background', '仅在后台启动', '登录后继续下载，不打开主窗口', !prefs.login)}${settingSwitch('resume', '恢复上次运行的任务', '手动暂停的任务保持暂停')}</div><p class="settings-caption">关闭主窗口后，下载会继续。<br>从 Rivu 菜单选择“退出 Rivu”可结束应用。</p><p class="settings-caption">演示中的开关不修改系统登录项。</p>`
    else if (settingsTab === 'downloads')
      content = `<h3>下载与做种</h3><div class="setting-group"><label class="setting-row"><span>默认保存位置</span><select data-pref="directory" aria-label="默认保存位置"><option ${prefs.directory === '~/Downloads/Rivu' ? 'selected' : ''}>~/Downloads/Rivu</option><option ${prefs.directory === '~/Movies/Open Movies' ? 'selected' : ''}>~/Movies/Open Movies</option></select></label><label class="setting-row"><span>同时下载的任务数</span><input type="number" min="1" max="20" data-pref="concurrent" value="${prefs.concurrent}" aria-label="同时下载的任务数"></label><label class="setting-row"><span>完成后停止做种的分享率</span><select data-pref="ratio" aria-label="做种目标">${['0', '1.0', '2.0', '不限'].map(v => `<option ${prefs.ratio === v ? 'selected' : ''}>${v}</option>`).join('')}</select></label></div><h3 style="font-size:13px;margin:24px 0 13px">临时限速</h3><div class="setting-group"><div class="setting-row"><span>下载速度上限</span><span>2 MB/s</span></div><div class="setting-row"><span>上传速度上限</span><span>512 KB/s</span></div></div><p class="settings-caption">可从主窗口底部或菜单栏面板切换限速模式。带宽数值为本轮设计预设。</p>`
    else if (settingsTab === 'network')
      content = `<h3>连接</h3><div class="setting-group"><label class="setting-row"><span>监听端口</span><input type="number" min="1" max="65535" data-pref="port" value="${prefs.port}" aria-label="监听端口"></label>${settingSwitch('dht', 'DHT 节点发现')}${settingSwitch('pex', 'PEX 节点交换')}</div><p class="settings-caption">私有种子遵守 Tracker 的私有规则，不使用相应的公共节点发现方式。</p><button class="secondary" data-action="diagnose">检查连接</button><p class="settings-caption">连接信息为演示数据，原型不发起网络连接。</p>`
    else if (settingsTab === 'notifications')
      content = `<h3>通知</h3><div class="setting-group">${settingSwitch('completeNotice', '下载完成时通知')}${settingSwitch('errorNotice', '任务需要处理时通知', '空间不足、磁盘断开或保存失败')}</div><div class="update-state">系统通知权限：尚未请求<p>首次使用通知时请求 macOS 授权。拒绝后提供系统设置入口。</p><button class="secondary" data-action="permission">打开通知设置…</button></div><p class="settings-caption">专注模式可能延迟系统通知；任务状态仍在 Rivu 内显示。</p>`
    else {
      const states = {
        available:
          '<strong>Rivu 0.0.2 可供更新</strong><p>提升任务恢复稳定性，优化菜单栏进度显示。</p><p>示例版本 · 18.4 MB</p><button class="primary" data-action="install">下载并安装更新…</button>',
        checking:
          '<strong>正在检查更新…</strong><p>正在读取版本信息（演示）</p>',
        installing:
          '<strong>正在下载并验证更新…</strong><p>这是更新流程演示，不会下载安装包。</p>',
        ready:
          '<strong>更新已准备就绪（演示）</strong><p>新版本将在下次启动时生效。</p><button class="primary" data-action="restart">保存任务并重新启动</button>',
        latest:
          '<strong>已是最新版本（演示）</strong><p>当前运行 Rivu 0.0.2</p>',
        disabled:
          '<strong>当前构建未配置更新</strong><p>没有更新源或验证公钥，无法检查。正式版配置完成后可用。</p>',
        failed:
          '<strong>无法连接更新服务器</strong><p>请检查网络后重试。上次成功检查：今天 09:00（演示）。</p>',
      }
      content = `<h3>更新与关于</h3><div class="update-app"><div class="app-icon">${icon('download')}</div><div><strong>Rivu</strong><span>版本 ${updateState === 'latest' ? '0.0.2' : '0.0.1'} · 为你的 Mac 而设计</span></div></div><div class="update-state" role="status">${states[updateState]}</div><div class="setting-group">${settingSwitch('automatic', '自动检查更新', '获得同意后，每天检查一次')}${settingSwitch('install', '空闲时自动安装更新', '无下载、校验或做种任务时安装')}</div><button class="secondary" data-action="check-update" ${['checking', 'installing'].includes(updateState) ? 'disabled' : ''}>检查更新</button><label class="setting-row"><span>预览更新状态</span><select id="update-scenario" aria-label="预览更新状态"><option value="available" ${updateState === 'available' ? 'selected' : ''}>有新版本</option><option value="disabled" ${updateState === 'disabled' ? 'selected' : ''}>更新未配置</option><option value="failed" ${updateState === 'failed' ? 'selected' : ''}>网络失败</option></select></label><p class="settings-caption">自动安装默认关闭。传输期间不会自动重新启动。</p>`
    }
    $('#settings-content').innerHTML = content
  }
  function openSettings(tab = settingsTab) {
    settingsTab = tab
    renderSettings()
    openDialog('#settings-dialog')
  }
  function scenario(name) {
    $$('.review-nav button').forEach(b =>
      b.classList.toggle('active', b.dataset.scene === name),
    )
    $$('dialog[open]').forEach(d => d.close())
    if (name === 'main') {
      showMain()
      closeTray()
    }
    if (name === 'tray') toggleTray(true)
    if (name === 'add') {
      $('#magnet').value = sampleMagnet
      setMetadata(true)
      openAdd()
    }
    if (name === 'settings') openSettings()
  }
  const actions = {
    add: openAdd,
    settings: () => openSettings(),
    batch,
    limit: toggleLimit,
    'open-main': () => {
      showMain()
      closeTray()
    },
    errors: () => {
      filter = 'error'
      showMain()
      closeTray()
      render()
    },
    'clear-search': () => {
      $('#search').value = ''
      render()
    },
    reveal: () => toast('演示：在 Finder 中显示下载文件'),
    'fix-path': () => {
      const task = current()
      task.path = '~/Downloads/Rivu'
      task.state = 'downloading'
      task.down = 2.1
      task.up = 0.08
      task.eta = '约 15 分钟'
      render()
      toast('演示：已更换到 Macintosh HD，恢复下载')
    },
    diagnose: () => toast('演示：监听端口可用，DHT 与 Tracker 连接正常'),
    permission: () => toast('演示：打开 macOS 系统设置 → 通知 → Rivu'),
    'check-update': () => {
      if (updateState === 'disabled') {
        toast('当前构建未配置更新源和验证公钥')
        return
      }
      updateState = 'checking'
      renderSettings()
      setTimeout(() => {
        updateState = 'available'
        renderSettings()
      }, 850)
    },
    install: () => {
      updateState = 'installing'
      renderSettings()
      setTimeout(() => {
        updateState = 'ready'
        renderSettings()
      }, 1100)
    },
    restart: () => {
      updateState = 'latest'
      renderSettings()
      toast('演示：已保存任务并重新启动，恢复原本运行的任务')
    },
  }
  document.addEventListener('click', event => {
    const target = event.target.closest('button')
    if (
      !$('#tray-panel').hidden
      && !event.target.closest('#tray-panel,#tray-trigger,[data-scene=tray]')
    )
      closeTray()
    if (!target) return
    if (target.dataset.scene) scenario(target.dataset.scene)
    if (target.dataset.filter) {
      filter = target.dataset.filter
      render()
    }
    if (target.dataset.select) selectTask(Number(target.dataset.select))
    if (target.dataset.toggle) toggleTask(Number(target.dataset.toggle))
    if (target.dataset.openTask) {
      filter = 'all'
      $('#search').value = ''
      selected = Number(target.dataset.openTask)
      showMain()
      closeTray()
      render()
    }
    if (target.dataset.detail) {
      detailTab = target.dataset.detail
      renderInspector()
      $(`[data-detail="${detailTab}"]`).focus()
    }
    if (target.dataset.settingTab) {
      settingsTab = target.dataset.settingTab
      renderSettings()
    }
    if (target.dataset.action) actions[target.dataset.action]?.()
    if (target.classList.contains('close-dialog'))
      target.closest('dialog').close()
  })
  document.addEventListener('change', event => {
    const target = event.target
    if (target.matches('input[name=file]')) updateSelectedSize()
    if (target.dataset.demoFile)
      toast(target.checked ? '演示：已选择此文件' : '演示：此文件将不再下载')
    if (target.dataset.pref) {
      const key = target.dataset.pref
      prefs[key] = target.type === 'checkbox' ? target.checked : target.value
      if (key === 'theme') setTheme(prefs.theme)
      if (key === 'menubar') {
        $('#tray-trigger').hidden = !prefs.menubar
        closeTray()
      }
      if (key === 'directory') $('#destination').value = prefs.directory
      renderSettings()
    }
    if (target.id === 'update-scenario') {
      updateState = target.value
      renderSettings()
    }
  })
  $('#tray-trigger').addEventListener('click', () => toggleTray())
  $('#window-close').addEventListener('click', () => {
    $('#main-window').hidden = true
    $('#closed-window').hidden = false
    toast('主窗口已关闭，任务继续在后台运行')
  })
  $('#reopen-window').addEventListener('click', showMain)
  $('#settings-trigger').addEventListener('click', () => openSettings())
  $('#add-trigger').addEventListener('click', openAdd)
  $('#pause-selected').addEventListener('click', () => toggleTask(selected))
  $('#limit-trigger').addEventListener('click', toggleLimit)
  $('#search').addEventListener('input', render)
  $('#sort').addEventListener('change', render)
  $('#inspector-toggle').addEventListener('click', () => {
    const grid = $('#content-grid')
    if (matchMedia('(max-width: 1000px)').matches)
      grid.classList.toggle('show-inspector')
    else grid.classList.toggle('hide-inspector')
    $('#inspector-toggle').setAttribute(
      'aria-pressed',
      matchMedia('(max-width: 1000px)').matches
        ? grid.classList.contains('show-inspector')
        : !grid.classList.contains('hide-inspector'),
    )
  })
  $('#theme-toggle').addEventListener('click', () =>
    setTheme(document.body.dataset.theme === 'dark' ? 'light' : 'dark'),
  )
  $('#demo-magnet').addEventListener('click', () => {
    $('#magnet').value = sampleMagnet
    setMetadata(true)
    $('#add-error').hidden = true
  })
  $('#demo-torrent').addEventListener('click', () => {
    $('#magnet').value = sampleMagnet
    setMetadata(true)
    toast('演示：已选择 BigBuckBunny.torrent')
  })
  $('#magnet').addEventListener('input', () => {
    setMetadata($('#magnet').value === sampleMagnet)
    $('#add-error').hidden = true
  })
  $('#add-form').addEventListener('submit', event => {
    event.preventDefault()
    const error = message => {
      $('#add-error').textContent = message
      $('#add-error').hidden = false
    }
    if (!$('#magnet').value.trim().startsWith('magnet:?'))
      return error('请输入磁力链接，或选择种子文件。')
    if (!metadataReady)
      return error('此原型不连接网络。请点击“填入示例”查看文件选择与添加流程。')
    const files = $$('input[name=file]:checked')
    if (!files.length) return error('请至少选择一个文件。')
    if (tasks.some(t => t.id === 7)) {
      filter = 'all'
      emptyScene = false
      selected = 7
      $('#search').value = ''
      $('#add-dialog').close()
      render()
      toast('任务已存在，已为你定位')
      return
    }
    const total = files.reduce((sum, input) => sum + Number(input.value), 0)
    tasks.unshift({
      id: 7,
      name: 'Big Buck Bunny',
      file: 'BigBuckBunny_1080p.mp4',
      type: 'movie',
      state: $('#start-now').checked ? 'downloading' : 'queued',
      progress: 0,
      total,
      down: 0,
      up: 0,
      eta: $('#start-now').checked ? '正在连接节点…' : '等待空闲下载位置',
      peers: '0 / 0',
      uploaded: '0 MB',
      ratio: '0.00',
      date: '刚刚',
      path: $('#destination').value,
      order: 7,
    })
    emptyScene = false
    filter = 'all'
    selected = 7
    $('#search').value = ''
    $('#add-dialog').close()
    showMain()
    render()
    toast('已添加示例任务，未发起真实下载')
  })
  $('#remove-trigger').addEventListener('click', () => {
    if (!current()) return
    $('#remove-name').textContent = current().name
    $('#remove-files').checked = false
    openDialog('#remove-dialog')
  })
  $('#confirm-remove').addEventListener('click', () => {
    tasks = tasks.filter(t => t.id !== selected)
    batchPaused = batchPaused.filter(id => id !== selected)
    $('#remove-dialog').close()
    render()
    toast(
      $('#remove-files').checked
        ? '演示：任务已移除，文件移到废纸篓'
        : '任务已移除，保留下载文件',
    )
  })
  $('#empty-scene').addEventListener('click', () => {
    emptyScene = true
    filter = 'all'
    $('#search').value = ''
    showMain()
    render()
  })
  $('#error-scene').addEventListener('click', () => {
    emptyScene = false
    if (!tasks.some(t => t.state === 'error'))
      tasks.push({ ...structuredClone(initial[5]), id: 60, order: 8 })
    filter = 'error'
    $('#search').value = ''
    showMain()
    render()
  })
  $('#reset-scene').addEventListener('click', () => {
    tasks = structuredClone(initial)
    selected = 1
    filter = 'all'
    detailTab = 'overview'
    limited = false
    emptyScene = false
    batchPaused = []
    $('#search').value = ''
    $('#sort').value = 'added'
    showMain()
    closeTray()
    render()
    toast('任务演示已重置')
  })
  document.addEventListener('keydown', event => {
    if (event.key === 'Escape') closeTray()
    if (event.metaKey && event.key === 'n') {
      event.preventDefault()
      openAdd()
    }
    if (event.metaKey && event.key === ',') {
      event.preventDefault()
      openSettings()
    }
    if (event.metaKey && event.key === '0') {
      event.preventDefault()
      showMain()
      closeTray()
    }
    if (event.code === 'Space' && event.target.matches('.task-select')) {
      event.preventDefault()
      toggleTask(Number(event.target.dataset.select))
    }
  })
  matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    if (prefs.theme === 'system') setTheme('system')
  })
  $$('dialog').forEach(dialog =>
    dialog.addEventListener('close', refreshSceneNav),
  )
  window.addEventListener('resize', positionTray)
  icons()
  render()
})()
