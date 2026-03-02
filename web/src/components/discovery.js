export function discovery() {
  return {
    status: 'searching', // searching, connected, failed
    serverIP: '',
    log: [], // Track all attempts
    
    logMsg(msg) {
      this.log.push(`[${new Date().toLocaleTimeString()}] ${msg}`)
      console.log(`[Blip] ${msg}`)
    },
    
    async findServer() {
      this.log = []
      this.logMsg('Starting server discovery...')
      
      // If we're already being served from a server (not file://), ping current host!
      if (!window.location.protocol.startsWith('file:')) {
        const currentHost = window.location.hostname
        this.logMsg(`Checking current host: ${currentHost}:41111`)
        const found = await this.pingServer(currentHost)
        if (found) {
          this.serverIP = currentHost
          this.status = 'connected'
          this.logMsg(`SUCCESS: Found server at ${currentHost}!`)
          return currentHost
        }
        this.logMsg(`${currentHost} failed`)
      }
      
      // FAST PATH: Check localhost first - instant for local testing
      this.logMsg('Trying: localhost:41111')
      const localFound = await this.pingServer('localhost')
      if (localFound) {
        this.serverIP = 'localhost'
        this.status = 'connected'
        this.logMsg('SUCCESS: Found server at localhost!')
        return 'localhost'
      }
      this.logMsg('localhost failed')

      // Network scan as fallback
      const subnet = await this.detectSubnet()
      this.logMsg(`Detected subnet: ${subnet}`)
      
      if (!subnet) {
        this.status = 'failed'
        this.logMsg('ERROR: Could not detect subnet')
        return null
      }
      
      // Scan subnet
      this.logMsg(`Scanning subnet ${subnet}.1-254...`)
      const promises = []
      for (let i = 1; i <= 254; i++) {
        promises.push(this.pingServer(`${subnet}.${i}`))
      }
      
      const results = await Promise.allSettled(promises)
      const found = results.find(r => r.status === 'fulfilled' && r.value)
      
      if (found) {
        this.serverIP = found.value
        this.status = 'connected'
        this.logMsg(`SUCCESS: Found server at ${found.value}!`)
        return found.value
      }
      
      this.status = 'failed'
      this.logMsg('ERROR: No server found on network')
      return null
    },
    
    async detectSubnet() {
      // Try to detect by pinging common servers or using WebRTC
      // For simplicity, we'll try common subnets
      const commonSubnets = ['192.168.1', '192.168.0', '10.0.0', '192.168.2']
      
      for (const subnet of commonSubnets) {
        // Try a few IPs quickly
        for (let i = 1; i <= 5; i++) {
          const found = await this.pingServer(`${subnet}.${i}`)
          if (found) {
            return subnet
          }
        }
      }
      
      // Default to most common
      return '192.168.1'
    },
    
    async pingServer(ip) {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 500)
      
      try {
        const response = await fetch(`http://${ip}:41111/ping`, {
          method: 'GET',
          signal: controller.signal,
        })
        clearTimeout(timeout)
        
        if (response.ok) {
          return ip
        }
        return null
      } catch (err) {
        clearTimeout(timeout)
        // Log failure reason (but not every single one - too noisy)
        return null
      }
    },
    
    useManualIP(ip) {
      this.serverIP = ip
      this.status = 'connected'
    },
  }
}