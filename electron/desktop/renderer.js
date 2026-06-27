const api = window.newApiDesktop

let config = null
let selectedId = ''

const dict = {
  en: {
    title: 'New API Desktop',
    subtitle: 'Connect the bundled frontends to any backend.',
    instances: 'Instances',
    new: 'New',
    instanceSettings: 'Instance settings',
    name: 'Name',
    backendUrl: 'Backend URL',
    loginMode: 'Login mode',
    interactive: 'Interactive login',
    accessToken: 'Access token',
    userId: 'User ID',
    token: 'Access token',
    frontend: 'Frontend',
    defaultFrontend: 'Default frontend',
    classicFrontend: 'Classic frontend',
    validate: 'Validate',
    save: 'Save',
    delete: 'Delete instance',
    launch: 'Launch',
    openDefault: 'Open default frontend',
    openClassic: 'Open classic frontend',
    refresh: 'Refresh status',
    checkUpdates: 'Check updates',
    saved: 'Saved',
    validated: 'Token validated',
    noInstances: 'No instances yet.',
  },
  zh: {
    title: 'New API 桌面端',
    subtitle: '使用内置前端连接任意后端。',
    instances: '实例',
    new: '新建',
    instanceSettings: '实例设置',
    name: '名称',
    backendUrl: '后端地址',
    loginMode: '登录方式',
    interactive: '交互式登录',
    accessToken: 'Access token',
    userId: '用户 ID',
    token: 'Access token',
    frontend: '前端',
    defaultFrontend: '新版前端',
    classicFrontend: '经典前端',
    validate: '验证',
    save: '保存',
    delete: '删除实例',
    launch: '启动',
    openDefault: '打开新版前端',
    openClassic: '打开经典前端',
    refresh: '刷新状态',
    checkUpdates: '检查更新',
    saved: '已保存',
    validated: 'Token 验证通过',
    noInstances: '还没有实例。',
  },
  fr: {},
  ja: {},
  ru: {},
  vi: {},
}

for (const lang of ['fr', 'ja', 'ru', 'vi']) dict[lang] = dict.en

const $ = (id) => document.getElementById(id)

function lang() {
  const value = config?.desktopLanguage || 'auto'
  if (value !== 'auto') return value
  const locale = (config?.appLocale || navigator.language || 'en').toLowerCase()
  if (locale.startsWith('zh')) return 'zh'
  const short = locale.slice(0, 2)
  return dict[short] ? short : 'en'
}

function tr(key) {
  return dict[lang()]?.[key] || dict.en[key] || key
}

function applyI18n() {
  document.querySelectorAll('[data-i18n]').forEach((node) => {
    node.textContent = tr(node.dataset.i18n)
  })
}

function renderLanguage() {
  const language = $('language')
  language.innerHTML = ''
  const options = [
    ['auto', 'Auto'],
    ['en', 'English'],
    ['zh', '中文'],
    ['fr', 'Français'],
    ['ja', '日本語'],
    ['ru', 'Русский'],
    ['vi', 'Tiếng Việt'],
  ]
  for (const [value, label] of options) {
    const option = document.createElement('option')
    option.value = value
    option.textContent = label
    language.appendChild(option)
  }
  language.value = config?.desktopLanguage || 'auto'
}

function renderInstances() {
  const container = $('instances')
  container.innerHTML = ''
  if (!config.instances.length) {
    const empty = document.createElement('p')
    empty.textContent = tr('noInstances')
    container.appendChild(empty)
    return
  }
  for (const item of config.instances) {
    const button = document.createElement('button')
    button.type = 'button'
    button.className = `instance-item${item.id === selectedId ? ' active' : ''}`
    button.innerHTML = `<strong></strong><small></small>`
    button.querySelector('strong').textContent = item.name
    button.querySelector('small').textContent = `${item.baseUrl} · ${item.authMode === 'accessToken' ? 'Token' : 'Interactive'}`
    button.addEventListener('click', () => selectInstance(item.id))
    container.appendChild(button)
  }
}

function emptyForm() {
  $('id').value = ''
  $('name').value = ''
  $('baseUrl').value = ''
  $('authMode').value = 'interactive'
  $('userId').value = ''
  $('accessToken').value = ''
  $('flavor').value = 'default'
}

function selectInstance(id) {
  selectedId = id
  const item = config.instances.find((instance) => instance.id === id)
  if (!item) {
    emptyForm()
  } else {
    $('id').value = item.id
    $('name').value = item.name
    $('baseUrl').value = item.baseUrl
    $('authMode').value = item.authMode
    $('userId').value = item.userId || ''
    $('accessToken').value = ''
    $('accessToken').placeholder = item.hasAccessToken ? 'Saved token is hidden. Enter a new one to replace it.' : ''
    $('flavor').value = item.flavor || config.activeFlavor || 'default'
  }
  updateMode()
  renderInstances()
}

function updateMode() {
  const tokenMode = $('authMode').value === 'accessToken'
  document.querySelectorAll('.token-only').forEach((node) => {
    node.style.display = tokenMode ? 'grid' : 'none'
  })
}

function formValue() {
  return {
    id: $('id').value || undefined,
    name: $('name').value,
    baseUrl: $('baseUrl').value,
    authMode: $('authMode').value,
    userId: $('userId').value,
    accessToken: $('accessToken').value,
    flavor: $('flavor').value,
  }
}

function setStatus(message) {
  $('status').textContent = message || ''
}

async function load() {
  config = await api.getConfig()
  selectedId = selectedId || config.activeInstanceId || config.instances[0]?.id || ''
  renderLanguage()
  applyI18n()
  renderInstances()
  selectInstance(selectedId)
  if (config.status?.message) setStatus(`${config.status.ok ? 'OK' : 'ERR'}: ${config.status.message}`)
}

$('language').addEventListener('change', async () => {
  config = await api.setLanguage($('language').value)
  applyI18n()
  renderInstances()
})

$('newInstance').addEventListener('click', () => {
  selectedId = ''
  emptyForm()
  updateMode()
  renderInstances()
})

$('authMode').addEventListener('change', updateMode)

$('instanceForm').addEventListener('submit', async (event) => {
  event.preventDefault()
  try {
    config = await api.saveInstance(formValue())
    selectedId = config.activeInstanceId
    setStatus(tr('saved'))
    await load()
  } catch (err) {
    setStatus(err.message)
  }
})

$('validate').addEventListener('click', async () => {
  try {
    const item = formValue()
    if (item.authMode === 'accessToken') {
      await api.validateAccessToken(item)
      setStatus(tr('validated'))
      return
    }
    setStatus('Interactive login will be validated after opening the frontend.')
  } catch (err) {
    setStatus(err.message)
  }
})

$('deleteInstance').addEventListener('click', async () => {
  if (!$('id').value) return
  config = await api.deleteInstance($('id').value)
  selectedId = config.activeInstanceId || config.instances[0]?.id || ''
  await load()
})

$('openDefault').addEventListener('click', () => {
  if (selectedId) api.openWindow({ instanceId: selectedId, flavor: 'default' })
})

$('openClassic').addEventListener('click', () => {
  if (selectedId) api.openWindow({ instanceId: selectedId, flavor: 'classic' })
})

$('refreshStatus').addEventListener('click', async () => {
  const status = await api.refreshStatus()
  setStatus(`${status.ok ? 'OK' : 'ERR'}: ${status.message}`)
})

$('checkUpdates').addEventListener('click', async () => {
  const status = await api.checkForUpdates()
  setStatus(`${status.ok ? 'OK' : 'ERR'}: ${status.message}`)
})

api.onConfigChanged((next) => {
  config = next
  renderInstances()
})

load().catch((err) => setStatus(err.message))
