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
    backendUrlPlaceholder: 'https://example.com',
    loginMode: 'Login mode',
    interactive: 'Interactive login',
    accessToken: 'Access token',
    userId: 'User ID',
    token: 'Access token',
    tokenPlaceholder: 'Enter a new token to replace the saved one.',
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
    tokenAuth: 'Token',
    interactiveAuth: 'Interactive',
    interactiveValidation: 'Interactive login will be checked after opening the frontend.',
    ok: 'OK',
    error: 'Error',
  },
  zh: {
    title: 'New API 桌面端',
    subtitle: '使用内置前端连接任意后端。',
    instances: '实例',
    new: '新建',
    instanceSettings: '实例设置',
    name: '名称',
    backendUrl: '后端地址',
    backendUrlPlaceholder: 'https://example.com',
    loginMode: '登录方式',
    interactive: '交互式登录',
    accessToken: 'Access token',
    userId: '用户 ID',
    token: 'Access token',
    tokenPlaceholder: '输入新 token 可替换已保存的 token。',
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
    tokenAuth: 'Token',
    interactiveAuth: '交互式',
    interactiveValidation: '交互式登录会在打开前端后检查。',
    ok: '正常',
    error: '错误',
  },
  fr: {
    title: 'New API Desktop',
    subtitle: 'Connectez les frontends integres a n’importe quel backend.',
    instances: 'Instances',
    new: 'Nouveau',
    instanceSettings: 'Parametres de l’instance',
    name: 'Nom',
    backendUrl: 'URL du backend',
    backendUrlPlaceholder: 'https://example.com',
    loginMode: 'Mode de connexion',
    interactive: 'Connexion interactive',
    accessToken: 'Access token',
    userId: 'ID utilisateur',
    token: 'Access token',
    tokenPlaceholder: 'Saisissez un nouveau token pour remplacer celui enregistre.',
    frontend: 'Frontend',
    defaultFrontend: 'Frontend par defaut',
    classicFrontend: 'Frontend classique',
    validate: 'Verifier',
    save: 'Enregistrer',
    delete: 'Supprimer l’instance',
    launch: 'Lancer',
    openDefault: 'Ouvrir le frontend par defaut',
    openClassic: 'Ouvrir le frontend classique',
    refresh: 'Actualiser',
    checkUpdates: 'Mises a jour',
    saved: 'Enregistre',
    validated: 'Token valide',
    noInstances: 'Aucune instance.',
    tokenAuth: 'Token',
    interactiveAuth: 'Interactif',
    interactiveValidation: 'La connexion interactive sera verifiee apres ouverture du frontend.',
    ok: 'OK',
    error: 'Erreur',
  },
  ja: {
    title: 'New API Desktop',
    subtitle: '内蔵フロントエンドを任意のバックエンドへ接続します。',
    instances: 'インスタンス',
    new: '新規',
    instanceSettings: 'インスタンス設定',
    name: '名前',
    backendUrl: 'バックエンド URL',
    backendUrlPlaceholder: 'https://example.com',
    loginMode: 'ログイン方式',
    interactive: '対話式ログイン',
    accessToken: 'Access token',
    userId: 'ユーザー ID',
    token: 'Access token',
    tokenPlaceholder: '保存済み token を置き換える場合のみ入力します。',
    frontend: 'フロントエンド',
    defaultFrontend: 'デフォルト版',
    classicFrontend: 'クラシック版',
    validate: '検証',
    save: '保存',
    delete: 'インスタンスを削除',
    launch: '起動',
    openDefault: 'デフォルト版を開く',
    openClassic: 'クラシック版を開く',
    refresh: '状態更新',
    checkUpdates: '更新確認',
    saved: '保存しました',
    validated: 'Token を検証しました',
    noInstances: 'インスタンスはありません。',
    tokenAuth: 'Token',
    interactiveAuth: '対話式',
    interactiveValidation: '対話式ログインはフロントエンドを開いた後に確認されます。',
    ok: 'OK',
    error: 'エラー',
  },
  ru: {
    title: 'New API Desktop',
    subtitle: 'Подключайте встроенные фронтенды к любому backend.',
    instances: 'Экземпляры',
    new: 'Создать',
    instanceSettings: 'Настройки экземпляра',
    name: 'Название',
    backendUrl: 'URL backend',
    backendUrlPlaceholder: 'https://example.com',
    loginMode: 'Способ входа',
    interactive: 'Интерактивный вход',
    accessToken: 'Access token',
    userId: 'ID пользователя',
    token: 'Access token',
    tokenPlaceholder: 'Введите новый token, чтобы заменить сохраненный.',
    frontend: 'Frontend',
    defaultFrontend: 'Новый frontend',
    classicFrontend: 'Классический frontend',
    validate: 'Проверить',
    save: 'Сохранить',
    delete: 'Удалить экземпляр',
    launch: 'Запуск',
    openDefault: 'Открыть новый frontend',
    openClassic: 'Открыть классический frontend',
    refresh: 'Обновить',
    checkUpdates: 'Проверить обновления',
    saved: 'Сохранено',
    validated: 'Token проверен',
    noInstances: 'Экземпляров пока нет.',
    tokenAuth: 'Token',
    interactiveAuth: 'Интерактивно',
    interactiveValidation: 'Интерактивный вход будет проверен после открытия frontend.',
    ok: 'OK',
    error: 'Ошибка',
  },
  vi: {
    title: 'New API Desktop',
    subtitle: 'Ket noi frontend tich hop voi bat ky backend nao.',
    instances: 'Phien ban',
    new: 'Moi',
    instanceSettings: 'Cai dat phien ban',
    name: 'Ten',
    backendUrl: 'URL backend',
    backendUrlPlaceholder: 'https://example.com',
    loginMode: 'Che do dang nhap',
    interactive: 'Dang nhap tuong tac',
    accessToken: 'Access token',
    userId: 'ID nguoi dung',
    token: 'Access token',
    tokenPlaceholder: 'Nhap token moi de thay the token da luu.',
    frontend: 'Frontend',
    defaultFrontend: 'Frontend moi',
    classicFrontend: 'Frontend co dien',
    validate: 'Kiem tra',
    save: 'Luu',
    delete: 'Xoa phien ban',
    launch: 'Mo',
    openDefault: 'Mo frontend moi',
    openClassic: 'Mo frontend co dien',
    refresh: 'Lam moi',
    checkUpdates: 'Kiem tra cap nhat',
    saved: 'Da luu',
    validated: 'Token hop le',
    noInstances: 'Chua co phien ban.',
    tokenAuth: 'Token',
    interactiveAuth: 'Tuong tac',
    interactiveValidation: 'Dang nhap tuong tac se duoc kiem tra sau khi mo frontend.',
    ok: 'OK',
    error: 'Loi',
  },
}

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
  document.querySelectorAll('[data-i18n-placeholder]').forEach((node) => {
    node.placeholder = tr(node.dataset.i18nPlaceholder)
  })
  document.documentElement.lang = lang()
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
    button.querySelector('small').textContent = `${item.baseUrl} · ${item.authMode === 'accessToken' ? tr('tokenAuth') : tr('interactiveAuth')}`
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
    $('accessToken').placeholder = item.hasAccessToken ? tr('tokenPlaceholder') : ''
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
  if (config.status?.message) setStatus(`${config.status.ok ? tr('ok') : tr('error')}: ${config.status.message}`)
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
    setStatus(tr('interactiveValidation'))
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
  setStatus(`${status.ok ? tr('ok') : tr('error')}: ${status.message}`)
})

$('checkUpdates').addEventListener('click', async () => {
  const status = await api.checkForUpdates()
  setStatus(`${status.ok ? tr('ok') : tr('error')}: ${status.message}`)
})

api.onConfigChanged((next) => {
  config = next
  applyI18n()
  renderInstances()
})

load().catch((err) => setStatus(err.message))
