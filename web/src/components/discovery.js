export function discovery() {
  return {
    status: 'searching', // searching, connected, failed
    serverIP: '',
    log: [], // Track all attempts
    cachedIPKey: 'looty_cached_ip',
    
    logMsg(msg) {
      this.log.push(`[${new Date().toLocaleTimeString()}] ${msg}`)
      console.log(`[LOOTY] ${msg}`)
    },
    
    saveCachedIP(ip) {
      try {
        localStorage.setItem(this.cachedIPKey, ip)
      } catch (e) {
        // localStorage might be unavailable
      }
    },
    
    getCachedIP() {
      try {
        return localStorage.getItem(this.cachedIPKey)
      } catch (e) {
        return null
      }
    },
    
    async findServer() {
      this.log = []
      this.logMsg('Starting server discovery...')
      
      // If we're already being served from a server (not file://), ping current host first!
      if (!window.location.protocol.startsWith('file:')) {
        const currentHost = window.location.hostname
        this.logMsg(`Checking current host: ${currentHost}:41111`)
        const found = await this.pingServer(currentHost)
        if (found) {
          this.serverIP = currentHost
          this.status = 'connected'
          this.saveCachedIP(currentHost)
          this.logMsg(`SUCCESS: Found server at ${currentHost}!`)
          return currentHost
        }
        this.logMsg(`${currentHost} not responding`)
      }
      
      // 1. Try cached IP first (instant reconnect)
      const cachedIP = this.getCachedIP()
      if (cachedIP) {
        this.logMsg(`Trying cached IP: ${cachedIP}:41111`)
        const found = await this.pingServer(cachedIP)
        if (found) {
          this.serverIP = cachedIP
          this.status = 'connected'
          this.logMsg(`SUCCESS: Found server at cached IP ${cachedIP}!`)
          return cachedIP
        }
        this.logMsg(`Cached IP ${cachedIP} not responding`)
      }
      
      // 2. Try mDNS hostname (zero-config)
      this.logMsg('Trying looty.local:41111 (mDNS)')
      const localFound = await this.pingServer('looty.local')
      if (localFound) {
        this.serverIP = 'looty.local'
        this.status = 'connected'
        this.saveCachedIP('looty.local')
        this.logMsg('SUCCESS: Found server via mDNS!')
        return 'looty.local'
      }
      this.logMsg('looty.local not responding')
      
      // 3. Smart parallel scan - Tier 1: first 32 IPs of common subnets
      const subnets = ['192.168.1', '192.168.0', '10.0.0', '192.168.2']
      this.logMsg(`Tier 1: Scanning first 32 IPs of ${subnets.join(', ')} in parallel...`)
      
      const tier1Promises = []
      for (const subnet of subnets) {
        for (let i = 1; i <= 32; i++) {
          tier1Promises.push(this.pingServerWithIP(`${subnet}.${i}`))
        }
      }
      
      const tier1Results = await Promise.allSettled(tier1Promises)
      const tier1Found = tier1Results.find(r => r.status === 'fulfilled' && r.value)
      
      if (tier1Found) {
        const foundIP = tier1Found.value
        this.serverIP = foundIP
        this.status = 'connected'
        this.saveCachedIP(foundIP)
        this.logMsg(`SUCCESS: Found server at ${foundIP}!`)
        return foundIP
      }
      
      // 4. Tier 2: Expand to full subnet scan
      this.logMsg('Tier 2: Expanding to full subnet scan...')
      
      const tier2Promises = []
      for (const subnet of subnets) {
        for (let i = 33; i <= 254; i++) {
          tier2Promises.push(this.pingServerWithIP(`${subnet}.${i}`))
        }
      }
      
      const tier2Results = await Promise.allSettled(tier2Promises)
      const tier2Found = tier2Results.find(r => r.status === 'fulfilled' && r.value)
      
      if (tier2Found) {
        const foundIP = tier2Found.value
        this.serverIP = foundIP
        this.status = 'connected'
        this.saveCachedIP(foundIP)
        this.logMsg(`SUCCESS: Found server at ${foundIP}!`)
        return foundIP
      }
      
      this.status = 'failed'
      this.logMsg('ERROR: No server found on network')
      return null
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
        return null
      }
    },
    
    // Returns the IP if found, null otherwise (for Promise.allSettled usage)
    async pingServerWithIP(ip) {
      const found = await this.pingServer(ip)
      return found ? ip : null
    },
    
    useManualIP(ip) {
      this.serverIP = ip
      this.status = 'connected'
      this.saveCachedIP(ip)
    },
  }
}