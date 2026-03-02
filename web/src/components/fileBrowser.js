import { apiUrl } from '../utils.js'

export function fileBrowser() {
  return {
    files: [],
    currentPath: '.',
    selectedFile: null,
    preview: null,
    loading: false,
    uploadProgress: null,
    uploadSuccess: false,
    downloadProgress: null,
    sortBy: 'name', // 'name' or 'date'
    sortAsc: true,
    initialized: false,
    
    init() {
      // Wait for server discovery before loading files
      const checkReady = () => {
        if (window.API_BASE) {
          this.initialized = true
          this.loadFiles()
        } else {
          setTimeout(checkReady, 100)
        }
      }
      checkReady()
    },
    
    get breadcrumbPath() {
      if (this.currentPath === '.') return [{ name: 'Home', path: '.' }]
      const parts = this.currentPath.split('/')
      const crumbs = [{ name: 'Home', path: '.' }]
      let accum = ''
      for (const part of parts) {
        accum = accum ? accum + '/' + part : part
        crumbs.push({ name: part, path: accum })
      }
      return crumbs
    },
    
    get sortedFiles() {
      const sorted = [...this.files]
      // Always show directories first
      sorted.sort((a, b) => {
        if (a.isDir && !b.isDir) return -1
        if (!a.isDir && b.isDir) return 1
        if (this.sortBy === 'name') {
          return this.sortAsc 
            ? a.name.localeCompare(b.name)
            : b.name.localeCompare(a.name)
        } else if (this.sortBy === 'date') {
          const aTime = new Date(a.modified || 0).getTime()
          const bTime = new Date(b.modified || 0).getTime()
          return this.sortAsc ? aTime - bTime : bTime - aTime
        }
        return 0
      })
      return sorted
    },
    
    async loadFiles(path = '.') {
      this.loading = true
      this.currentPath = path
      this.selectedFile = null
      this.preview = null
      this.uploadSuccess = false
      try {
        const res = await fetch(apiUrl(`/api/files?path=${encodeURIComponent(path)}`))
        const data = await res.json()
        this.files = data.files || []
      } catch (err) {
        console.error('Failed to load files:', err)
        this.files = []
      }
      this.loading = false
    },
    
    navigateTo(path) {
      this.loadFiles(path)
    },
    
    navigateUp() {
      const parts = this.currentPath.split('/')
      parts.pop()
      const newPath = parts.join('/') || '.'
      this.navigateTo(newPath)
    },
    
    async selectFile(file) {
      if (file.isDir) {
        this.navigateTo(file.path)
        return
      }
      this.selectedFile = file
      
      // Server determines if file is binary - trust it completely
      if (file.isBinary) {
        this.preview = { type: 'binary', reason: 'Binary file (detected by server)' }
        return
      }
      
      // Not binary - try to show as text
      this.preview = { type: 'text', loading: true }
      try {
        const res = await fetch(apiUrl(`/api/download?path=${encodeURIComponent(file.path)}`))
        const text = await res.text()
        this.preview.content = text
        this.preview.loading = false
      } catch (err) {
        console.error('Failed to load text file:', err)
        this.preview = { type: 'error', message: 'Failed to load file' }
      }
    },
    
    async downloadFile(event) {
      // Prevent any default behavior or navigation
      if (event) {
        event.preventDefault()
        event.stopPropagation()
      }
      
      if (!this.selectedFile) return
      
      this.downloadProgress = 0
      try {
        const res = await fetch(apiUrl(`/api/download?path=${encodeURIComponent(this.selectedFile.path)}`))
        const contentLength = res.headers.get('content-length')
        const total = contentLength ? parseInt(contentLength, 10) : null
        
        const reader = res.body.getReader()
        const chunks = []
        let loaded = 0
        
        while (true) {
          const { done, value } = await reader.read()
          if (done) break
          chunks.push(value)
          loaded += value.length
          if (total) {
            this.downloadProgress = Math.round((loaded / total) * 100)
          }
        }
        
        const blob = new Blob(chunks)
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = this.selectedFile.name
        // Use programmatic click - don't append to DOM to avoid navigation issues
        a.click()
        URL.revokeObjectURL(url)
        
        this.downloadProgress = null
      } catch (err) {
        console.error('Download failed:', err)
        this.downloadProgress = null
      }
    },
    
    async uploadFile(event) {
      const file = event.target.files[0]
      if (!file) return
      
      this.uploadProgress = 0
      this.uploadSuccess = false
      
      const formData = new FormData()
      formData.append('file', file)
      formData.append('path', this.currentPath)
      
      try {
        const xhr = new XMLHttpRequest()
        
        await new Promise((resolve, reject) => {
          xhr.upload.addEventListener('progress', (e) => {
            if (e.lengthComputable) {
              this.uploadProgress = Math.round((e.loaded / e.total) * 100)
            }
          })
          
          xhr.addEventListener('load', () => {
            if (xhr.status === 200) {
              resolve()
            } else {
              reject(new Error('Upload failed'))
            }
          })
          
          xhr.addEventListener('error', reject)
          
          xhr.open('POST', apiUrl('/api/upload'))
          xhr.send(formData)
        })
        
        this.uploadProgress = null
        this.uploadSuccess = true
        this.loadFiles(this.currentPath)
        
        // Hide success message after 2 seconds
        setTimeout(() => {
          this.uploadSuccess = false
        }, 2000)
      } catch (err) {
        console.error('Upload failed:', err)
        this.uploadProgress = null
        this.uploadSuccess = false
      }
      
      event.target.value = ''
    },
    
    toggleSort() {
      if (this.sortBy === 'name') {
        this.sortBy = 'date'
        this.sortAsc = false
      } else {
        this.sortBy = 'name'
        this.sortAsc = true
      }
    },
    
    formatSize(bytes) {
      if (bytes == null) return ''
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    },
    
    formatDate(dateStr) {
      if (!dateStr) return ''
      const date = new Date(dateStr)
      const now = new Date()
      const diff = now - date
      const days = Math.floor(diff / (1000 * 60 * 60 * 24))
      
      if (days === 0) {
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      } else if (days === 1) {
        return 'Yesterday'
      } else if (days < 7) {
        return days + 'd ago'
      } else {
        return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
      }
    },
  }
}