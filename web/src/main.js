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
  
  async discoverServer() {
    const disco = discovery()
    const ip = await disco.findServer()
    
    if (ip) {
      this.serverIP = ip
      this.connected = true
      this.status = 'connected'
      // Redirect or set API base URL
      window.API_BASE = `http://${ip}:8080`
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
      window.API_BASE = `http://${this.manualIP}:8080`
    }
  },
}))

Alpine.data('fileBrowser', fileBrowser)
Alpine.data('clipboardPanel', clipboardPanel)

Alpine.start()