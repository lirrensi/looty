// Get the API base URL - uses discovered server IP if available, otherwise uses current host
export function getApiBase() {
  if (window.API_BASE) {
    return window.API_BASE
  }
  const protocol = window.location.protocol === 'https:' ? 'https:' : 'http:'
  return `${protocol}//${window.location.host}`
}

// Helper to build full API URL
export function apiUrl(path) {
  return `${getApiBase()}${path}`
}

// Helper to get WebSocket URL
export function wsUrl() {
  const base = getApiBase()
  const wsProtocol = base.startsWith('https') ? 'wss:' : 'ws:'
  // Extract host from base URL
  const host = base.replace(/^https?:\/\//, '')
  return `${wsProtocol}//${host}/ws`
}