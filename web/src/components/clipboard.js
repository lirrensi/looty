import { wsUrl } from '../utils.js'

export function clipboardPanel() {
  return {
    text: '',
    history: [],
    ws: null,
    copySuccess: false,
    sendSuccess: false,
    wsConnected: false,
    wsError: '',
    
    init() {
      // Wait for server discovery before connecting WebSocket
      const checkReady = () => {
        if (window.API_BASE) {
          this.connectWebSocket()
        } else {
          setTimeout(checkReady, 100)
        }
      }
      checkReady()
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
            if (msg.type === 'clipboard') {
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
          this.wsError = 'Connection failed - clipboard sync unavailable'
        }
      } catch (err) {
        console.error('Failed to create WebSocket:', err)
        this.wsError = 'Cannot connect to server'
      }
    },
    
    send() {
      if (!this.text.trim()) return
      
      const msg = JSON.stringify({
        type: 'clipboard',
        data: this.text,
      })
      
      this.ws.send(msg)
      this.addToHistory(this.text)
      this.text = ''
      this.sendSuccess = true
      setTimeout(() => {
        this.sendSuccess = false
      }, 1500)
    },
    
    addToHistory(text) {
      // Add to beginning
      this.history.unshift({
        text: text,
        time: new Date().toLocaleTimeString(),
      })
      
      // Keep only last 10
      if (this.history.length > 10) {
        this.history.pop()
      }
    },
    
    async copyToClipboard(text) {
      try {
        await navigator.clipboard.writeText(text)
        this.copySuccess = true
        setTimeout(() => {
          this.copySuccess = false
        }, 1500)
      } catch (err) {
        console.error('Failed to copy:', err)
        // Fallback for older browsers
        const textarea = document.createElement('textarea')
        textarea.value = text
        document.body.appendChild(textarea)
        textarea.select()
        document.execCommand('copy')
        document.body.removeChild(textarea)
        this.copySuccess = true
        setTimeout(() => {
          this.copySuccess = false
        }, 1500)
      }
    },
    
    useHistoryItem(item) {
      this.text = item.text
    },
  }
}