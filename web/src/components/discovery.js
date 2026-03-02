export function discovery() {
  return {
    status: 'searching', // searching, connected, failed
    serverIP: '',
    
    async findServer() {
      // Get current IP to determine subnet
      const subnet = await this.detectSubnet()
      
      if (!subnet) {
        this.status = 'failed'
        return null
      }
      
      // Scan subnet
      const promises = []
      for (let i = 1; i <= 254; i++) {
        promises.push(this.pingServer(`${subnet}.${i}`))
      }
      
      const results = await Promise.allSettled(promises)
      const found = results.find(r => r.status === 'fulfilled' && r.value)
      
      if (found) {
        this.serverIP = found.value
        this.status = 'connected'
        return found.value
      }
      
      this.status = 'failed'
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
        const response = await fetch(`http://${ip}:8080/ping`, {
          method: 'GET',
          signal: controller.signal,
        })
        clearTimeout(timeout)
        
        if (response.ok) {
          return ip
        }
        return null
      } catch {
        clearTimeout(timeout)
        return null
      }
    },
    
    useManualIP(ip) {
      this.serverIP = ip
      this.status = 'connected'
    },
  }
}