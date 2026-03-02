import { apiUrl } from '../utils.js'

export function fileBrowser() {
  return {
    files: [],
    currentPath: '.',
    selectedFile: null,
    preview: null,
    loading: false,
    
    async loadFiles(path = '.') {
      this.loading = true
      this.currentPath = path
      try {
        const res = await fetch(apiUrl(`/api/files?path=${encodeURIComponent(path)}`))
        const data = await res.json()
        this.files = data.files
      } catch (err) {
        console.error('Failed to load files:', err)
      }
      this.loading = false
    },
    
    navigateTo(path) {
      this.loadFiles(path)
      this.selectedFile = null
      this.preview = null
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
      
      // Check if previewable
      const ext = file.name.split('.').pop().toLowerCase()
      const textExts = ['txt', 'md', 'json', 'log', 'js', 'ts', 'css', 'html', 'xml', 'yaml', 'yml']
      const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg']
      
      if (textExts.includes(ext)) {
        this.preview = { type: 'text', loading: true }
        const res = await fetch(apiUrl(`/api/download?path=${encodeURIComponent(file.path)}`))
        this.preview.content = await res.text()
        this.preview.loading = false
      } else if (imageExts.includes(ext)) {
        this.preview = { type: 'image', url: apiUrl(`/api/download?path=${encodeURIComponent(file.path)}`) }
      } else {
        this.preview = { type: 'binary' }
      }
    },
    
    downloadFile() {
      if (!this.selectedFile) return
      window.open(apiUrl(`/api/download?path=${encodeURIComponent(this.selectedFile.path)}`))
    },
    
    async uploadFile(event) {
      const file = event.target.files[0]
      if (!file) return
      
      const formData = new FormData()
      formData.append('file', file)
      formData.append('path', this.currentPath)
      
      try {
        await fetch(apiUrl('/api/upload'), {
          method: 'POST',
          body: formData,
        })
        this.loadFiles(this.currentPath)
      } catch (err) {
        console.error('Upload failed:', err)
      }
      
      event.target.value = ''
    },
    
    formatSize(bytes) {
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    },
  }
}