// FILE: web/src/main.js
// PURPOSE: Bootstrap the standalone Looty client and bind discovery results into API connection state.
// OWNS: App-level discovery startup, API base selection, manual connection flow.
// EXPORTS: Alpine app registration side effects
// DOCS: agent_chat/plan_qr-port-artifact_2026-05-17.md, docs/spec.md, docs/arch.md

import Alpine from 'alpinejs'
import './style.css'
import { fileBrowser } from './components/fileBrowser.js'
import { clipboardPanel } from './components/clipboard.js'
import { discovery, resolveLootyPort } from './components/discovery.js'

function buildApiBase(ip) {
  const protocol = window.location.protocol === 'https:' ? 'https' : 'http'

  if (!window.location.protocol.startsWith('file:') && window.location.hostname === ip) {
    return `${protocol}://${window.location.host}`
  }

  return `${protocol}://${ip}:${resolveLootyPort()}`
}

Alpine.data('app', () => ({
  connected: false,
  serverIP: '',
  status: 'searching',
  showClipboard: false,
  manualIP: '',
  
  init() {
    this.discoverServer()
    
    window.addEventListener('toggle-clipboard', () => {
      this.showClipboard = !this.showClipboard
    })
    
    window.addEventListener('files-refresh', () => {
      window.dispatchEvent(new CustomEvent('refresh-files'))
    })
  },
  
  disco: null,
  
  discoLog: [],
  
  async discoverServer() {
    const disco = discovery()
    this.discoLog = disco.log
    const ip = await disco.findServer()
    this.discoLog = [...disco.log]  // Force reactivity
    
    if (ip) {
      this.serverIP = ip
      this.connected = true
      this.status = 'connected'
      window.API_BASE = buildApiBase(ip)
    } else {
      this.status = 'failed'
    }
  },
  
  async connectManual() {
    if (!this.manualIP) return
    
    const disco = discovery()
    const found = await disco.pingServer(this.manualIP)
    
    if (found) {
      this.serverIP = this.manualIP
      this.connected = true
      this.status = 'connected'
      window.API_BASE = buildApiBase(this.manualIP)
    }
  },
}))

Alpine.data('fileBrowser', fileBrowser)
Alpine.data('clipboardPanel', clipboardPanel)

Alpine.start()
