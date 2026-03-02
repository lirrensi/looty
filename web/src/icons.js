// Extension to color mapping - distinct colors for quick navigation
const extensionColors = {
  // === OFFICE ===
  doc: '#2b579a', docx: '#2b579a',
  xls: '#217346', xlsx: '#217346',
  ppt: '#d24726', pptx: '#d24726',
  pdf: '#f40f02',
  
  // === ADOBE SUITE ===
  psd: '#001d34', ai: '#330000',
  ae: '#9999ff', pr: '#9933ff',
  indd: '#ff3366', xd: '#ff61f6',
  fig: '#f24e1e', sketch: '#fdad00',
  
  // === CODE - JAVASCRIPT/WEB ===
  js: '#f7df1e', mjs: '#f7df1e', cjs: '#f7df1e',
  ts: '#3178c6', tsx: '#3178c6',
  jsx: '#61dafb',
  vue: '#42b883', svelte: '#ff3e00',
  angular: '#dd0031',
  
  // === CODE - BACKEND ===
  go: '#00add8',
  rs: '#dea584',
  py: '#3776ab',
  rb: '#cc342d',
  php: '#777bb4',
  java: '#b07219',
  kt: '#7f52ff',
  swift: '#f05138',
  c: '#555555', h: '#555555',
  cpp: '#f34b7d', hpp: '#f34b7d', cc: '#f34b7d',
  cs: '#178600',
  lua: '#000080',
  scala: '#dc322f',
  clj: '#5881d8',
  ex: '#6e4a7e', exs: '#6e4a7e',
  hs: '#5e5086',
  el: '#4053a5',
  nim: '#ffc200',
  dart: '#0175c2',
  
  // === CODE - SYSTEM ===
  sh: '#4eaa25', bash: '#4eaa25', zsh: '#4eaa25',
  fish: '#34b000',
  ps1: '#012456', psm1: '#012456',
  bat: '#c1f12e', cmd: '#c1f12e',
  
  // === WEB ===
  html: '#e34c26', htm: '#e34c26',
  css: '#264de4',
  scss: '#cd6799', sass: '#cd6799', less: '#1d365d',
  
  // === DATA/CONFIG ===
  json: '#cbcb41',
  yaml: '#cb171e', yml: '#cb171e',
  xml: '#0060ac',
  toml: '#9c4121',
  sql: '#e38c00',
  graphql: '#e535ab', gql: '#e535ab',
  proto: '#c128c9',
  
  // === CONFIG ===
  env: '#ecd53f', gitignore: '#f14e32', dockerignore: '#2496ed',
  ini: '#6d8086', cfg: '#6d8086', conf: '#6d8086',
  
  // === MARKUP/DOCS ===
  md: '#083fa1', markdown: '#083fa1',
  rst: '#141414',
  tex: '#3d6117',
  rtf: '#7a8b8b',
  
  // === IMAGES ===
  png: '#a074c4', jpg: '#a074c4', jpeg: '#a074c4',
  gif: '#a074c4', webp: '#a074c4', bmp: '#a074c4',
  ico: '#a074c4', svg: '#ffb13b', avif: '#a074c4',
  heic: '#a074c4', heif: '#a074c4',
  tiff: '#a074c4', tif: '#a074c4',
  
  // === VIDEO ===
  mp4: '#ff5c5c', mov: '#ff5c5c', avi: '#ff5c5c',
  mkv: '#ff5c5c', wmv: '#ff5c5c', flv: '#ff5c5c',
  webm: '#ff5c5c', m4v: '#ff5c5c', mpeg: '#ff5c5c',
  
  // === AUDIO ===
  mp3: '#a855f7', wav: '#a855f7', ogg: '#a855f7',
  flac: '#a855f7', aac: '#a855f7', m4a: '#a855f7',
  wma: '#a855f7', aiff: '#a855f7',
  
  // === ARCHIVES ===
  zip: '#f97316', tar: '#f97316', gz: '#f97316',
  bz2: '#f97316', xz: '#f97316', '7z': '#f97316',
  rar: '#f97316', tgz: '#f97316',
  
  // === FONTS ===
  ttf: '#3b82f6', otf: '#3b82f6', woff: '#3b82f6',
  woff2: '#3b82f6', eot: '#3b82f6',
  
  // === DATABASE ===
  db: '#003b57', sqlite: '#003b57', sqlite3: '#003b57',
  prisma: '#2d3748',
  
  // === DEVOPS ===
  dockerfile: '#2496ed', docker: '#2496ed',
  kube: '#326ce5', kubernetes: '#326ce5',
  tf: '#7b42bc', hcl: '#7b42bc',
  
  // === EXECUTABLES ===
  exe: '#6b7280', msi: '#6b7280', dmg: '#6b7280',
  app: '#6b7280', deb: '#6b7280', rpm: '#6b7280',
  apk: '#3ddb93', ipa: '#007aff',
  
  // === OTHER COMMON ===
  torrent: '#567a46', iso: '#6b7280', img: '#6b7280',
  bin: '#6b7280', log: '#6b7280', txt: '#6b7280',
  csv: '#217346',
  lock: '#8b8b8b',
  wasm: '#654ff0',
  sol: '#363636',
}

// Shape types - determines which base SVG to use
const imageExtensions = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'ico', 'svg', 'avif', 'heic', 'heif', 'tiff', 'tif']
const videoExtensions = ['mp4', 'mov', 'avi', 'mkv', 'wmv', 'flv', 'webm', 'm4v', 'mpeg']
const audioExtensions = ['mp3', 'wav', 'ogg', 'flac', 'aac', 'm4a', 'wma', 'aiff']
const archiveExtensions = ['zip', 'tar', 'gz', 'bz2', 'xz', '7z', 'rar', 'tgz']
const fontExtensions = ['ttf', 'otf', 'woff', 'woff2', 'eot']

function getShapeType(ext) {
  if (imageExtensions.includes(ext)) return 'image'
  if (videoExtensions.includes(ext)) return 'video'
  if (audioExtensions.includes(ext)) return 'audio'
  if (archiveExtensions.includes(ext)) return 'archive'
  if (fontExtensions.includes(ext)) return 'font'
  return 'document'
}

// Base SVG shapes
const shapes = {
  folder: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" fill="${color}" fill-opacity="0.15"/>
      <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
    </svg>`,
  
  document: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" fill="${color}" fill-opacity="0.1"/>
      <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
      <polyline points="14 2 14 8 20 8"/>
      <line x1="16" y1="13" x2="8" y2="13"/>
      <line x1="16" y1="17" x2="8" y2="17"/>
      <line x1="10" y1="9" x2="8" y2="9"/>
    </svg>`,
  
  image: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <rect x="3" y="3" width="18" height="18" rx="2" ry="2" fill="${color}" fill-opacity="0.1"/>
      <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
      <circle cx="8.5" cy="8.5" r="1.5"/>
      <polyline points="21 15 16 10 5 21"/>
    </svg>`,
  
  video: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <rect x="2" y="4" width="20" height="16" rx="2" fill="${color}" fill-opacity="0.1"/>
      <rect x="2" y="4" width="20" height="16" rx="2"/>
      <polygon points="10 9 15 12 10 15 10 9" fill="${color}"/>
    </svg>`,
  
  audio: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M9 18V5l12-2v13" fill="${color}" fill-opacity="0.1"/>
      <circle cx="6" cy="18" r="3" fill="${color}" fill-opacity="0.2"/>
      <circle cx="18" cy="16" r="3" fill="${color}" fill-opacity="0.2"/>
      <path d="M9 18V5l12-2v13"/>
      <circle cx="6" cy="18" r="3"/>
      <circle cx="18" cy="16" r="3"/>
    </svg>`,
  
  archive: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <path d="M21 8v13H3V8" fill="${color}" fill-opacity="0.1"/>
      <path d="M1 3h22v5H1z"/>
      <path d="M10 12h4"/>
      <path d="M21 8v13H3V8"/>
    </svg>`,
  
  font: (color) => `
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="${color}" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
      <polyline points="4 7 4 4 20 4 20 7" fill="${color}" fill-opacity="0.1"/>
      <line x1="9" y1="20" x2="15" y2="20"/>
      <line x1="12" y1="4" x2="12" y2="20"/>
      <polyline points="4 7 4 4 20 4 20 7"/>
    </svg>`,
}

// Create badge overlay for extension
function createBadge(ext, color, size = 24) {
  const badgeWidth = size * 0.5
  const badgeHeight = size * 0.28
  const x = size - badgeWidth - 1
  const y = size - badgeHeight - 2
  const fontSize = size * 0.17
  const text = ext.toUpperCase().slice(0, 4)
  
  return `
    <g transform="translate(${x}, ${y})">
      <rect width="${badgeWidth}" height="${badgeHeight}" rx="2" fill="${color}"/>
      <text x="${badgeWidth/2}" y="${badgeHeight * 0.7}" 
            text-anchor="middle" 
            fill="white" 
            font-family="system-ui, -apple-system, sans-serif" 
            font-size="${fontSize}" 
            font-weight="600"
            style="text-shadow: 0 1px 1px rgba(0,0,0,0.3)">${text}</text>
    </g>`
}

// Main file icon function
export function fileIcon(filename, isDir = false, size = 24) {
  if (isDir) {
    return shapes.folder('#eab308') // yellow-500
  }
  
  const ext = filename.split('.').pop()?.toLowerCase() || ''
  const baseName = filename.toLowerCase()
  
  // Special case: hidden files start with dot, show dimmed
  if (filename.startsWith('.') && ext === filename.slice(1)) {
    const color = '#6b7280'
    return shapes.document(color)
  }
  
  // Get color for extension
  const color = extensionColors[ext] || '#6b7280'
  
  // Get shape type
  const shapeType = getShapeType(ext)
  
  // Generate SVG
  const baseShape = shapes[shapeType](color)
  
  // Add badge for documents (not for image/video/audio which have distinctive shapes)
  if (shapeType === 'document' && ext) {
    const badge = createBadge(ext, color, size)
    // Insert badge before closing SVG tag
    return baseShape.replace('</svg>', badge + '</svg>')
  }
  
  return baseShape
}

// UI icons for buttons (simple stroke icons)
export function icon(name, size = 20) {
  const icons = {
    upload: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>`,
    refresh: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>`,
    clipboard: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"/><rect x="8" y="2" width="8" height="4" rx="1" ry="1"/></svg>`,
    close: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`,
    chevronRight: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="9 18 15 12 9 6"/></svg>`,
    arrowUp: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="19" x2="12" y2="5"/><polyline points="5 12 12 5 19 12"/></svg>`,
    home: `<svg xmlns="http://www.w3.org/2000/svg" width="${size}" height="${size}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>`,
  }
  
  return icons[name] || ''
}

// Make available globally for Alpine templates
window.fileIcon = fileIcon
window.icon = icon