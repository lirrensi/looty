import Alpine from 'alpinejs'
import './style.css'
import { fileBrowser } from './components/fileBrowser.js'
import { clipboardPanel } from './components/clipboard.js'
import { discovery } from './components/discovery.js'

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
      const protocol = window.location.protocol === 'https:' ? 'https' : 'http'
      window.API_BASE = `${protocol}://${ip}:41111`
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
      const protocol = window.location.protocol === 'https:' ? 'https' : 'http'
      window.API_BASE = `${protocol}://${this.manualIP}:41111`
    }
  },
}))

Alpine.data('fileBrowser', fileBrowser)
Alpine.data('clipboardPanel', clipboardPanel)

Alpine.start()