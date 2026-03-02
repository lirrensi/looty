import { apiUrl, wsUrl } from '../utils.js'

export function clipboardPanel() {
  return {
    content: '',
    history: [],
    ws: null,
    copySuccess: false,
    sendSuccess: false,
    wsConnected: false,
    wsError: '',
    initialized: false,
    syncTimeout: null,

    init() {
      const checkReady = () => {
        if (window.API_BASE) {
          this.loadScratchpad()
          this.connectWebSocket()
        } else {
          setTimeout(checkReady, 100)
        }
      }
      checkReady()
    },

    async loadScratchpad() {
      try {
        const res = await fetch(apiUrl('/api/scratchpad'))
        const data = await res.json()
        this.content = data.content || ''
        this.addToHistory(data.content || '')
      } catch (err) {
        console.error('Failed to load scratchpad:', err)
      }
    },

    connectWebSocket() {
      const wsUrlStr = wsUrl()
      this.wsError = ''

      try {
        this.ws = new WebSocket(wsUrlStr)

        this.ws.onopen = () => {
          console.log('WebSocket connected')
          this.wsConnected = true
          this.wsError = ''
        }

        this.ws.onmessage = (event) => {
          try {
            const msg = JSON.parse(event.data)
            if (msg.type === 'scratchpad') {
              // Update from another client
              this.content = msg.data
              this.addToHistory(msg.data)
            } else if (msg.type === 'clipboard') {
              this.addToHistory(msg.data)
            } else if (msg.type === 'refresh') {
              this.$dispatch('files-refresh')
            }
          } catch (err) {
            console.error('Failed to parse message:', err)
          }
        }

        this.ws.onclose = () => {
          console.log('WebSocket disconnected, reconnecting...')
          this.wsConnected = false
          setTimeout(() => this.connectWebSocket(), 3000)
        }

        this.ws.onerror = (err) => {
          console.error('WebSocket error:', err)
          this.wsConnected = false
          this.wsError = 'Connection failed - scratchpad sync unavailable'
        }
      } catch (err) {
        console.error('Failed to create WebSocket:', err)
        this.wsError = 'Cannot connect to server'
      }
    },

    onInput() {
      // Debounce sync to avoid spamming
      if (this.syncTimeout) clearTimeout(this.syncTimeout)
      this.syncTimeout = setTimeout(() => {
        this.syncToServer()
      }, 300)
    },

    async syncToServer() {
      try {
        await fetch(apiUrl('/api/scratchpad'), {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ content: this.content })
        })
      } catch (err) {
        console.error('Failed to sync scratchpad:', err)
      }
    },

    addToHistory(text) {
      // Skip if text is empty or same as last item
      if (!text || text === this.history[0]?.text) return

      this.history.unshift({
        text: text,
        time: new Date().toLocaleTimeString(),
      })
      if (this.history.length > 50) {
        this.history.pop()
      }
    },

    async copyToClipboard(text) {
      try {
        await navigator.clipboard.writeText(text)
        this.copySuccess = true
        setTimeout(() => { this.copySuccess = false }, 1500)
      } catch (err) {
        console.error('Failed to copy:', err)
        const textarea = document.createElement('textarea')
        textarea.value = text
        document.body.appendChild(textarea)
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
        this.copySuccess = true
        setTimeout(() => { this.copySuccess = false }, 1500)
      }
    },

    useHistoryItem(item) {
      this.content = item.text
      this.syncToServer()
    },

    closeScratchpad() {
      // Access root app's showClipboard
      const root = this.$data.$data  // Get root app data
      if (root) {
        root.showClipboard = false
      } else {
        // Fallback: try to find from window
        const app = document.querySelector('[x-data="app()"]')
        if (app && app.__x) {
          app.__x.showClipboard = false
        }
      }
    },
  }
}
