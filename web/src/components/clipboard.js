import { wsUrl } from '../utils.js'

export function clipboardPanel() {
  return {
    text: '',
    history: [],
    ws: null,
    
    init() {
      this.connectWebSocket()
    },
    
    connectWebSocket() {
      const wsUrlStr = wsUrl()
      
      this.ws = new WebSocket(wsUrlStr)
      
      this.ws.onopen = () => {
        console.log('WebSocket connected')
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
        setTimeout(() => this.connectWebSocket(), 3000)
      }
      
      this.ws.onerror = (err) => {
        console.error('WebSocket error:', err)
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
    
    copyToClipboard(text) {
      navigator.clipboard.writeText(text)
    },
    
    useHistoryItem(item) {
      this.text = item.text
    },
  }
}