import test from 'node:test'
import assert from 'node:assert/strict'

import { discovery } from '../src/components/discovery.js'
import { clipboardPanel } from '../src/components/clipboard.js'
import { fileBrowser } from '../src/components/fileBrowser.js'

function setGlobals(overrides) {
  const previous = {}
  for (const [key, value] of Object.entries(overrides)) {
    previous[key] = globalThis[key]
    globalThis[key] = value
  }
  return () => {
    for (const [key, value] of Object.entries(previous)) {
      if (value === undefined) {
        delete globalThis[key]
      } else {
        globalThis[key] = value
      }
    }
  }
}

test('discovery pingServer uses HTTPS and returns null on failure', async () => {
  const restore = setGlobals({
    window: { location: { protocol: 'https:' } },
    fetch: async (url) => {
      assert.equal(url, 'https://example.local:41111/ping')
      throw new Error('offline')
    },
  })

  try {
    const d = discovery()
    assert.equal(await d.pingServer('example.local'), null)
  } finally {
    restore()
  }
})

test('discovery findServer surfaces final network failure and hint', async () => {
  const restore = setGlobals({
    window: { location: { protocol: 'file:' } },
    localStorage: {
      getItem: () => null,
      setItem: () => {},
    },
    console: { log: () => {}, error: () => {} },
  })

  try {
    const d = discovery()
    d.getCachedIP = () => null
    d.saveCachedIP = () => {}
    d.pingServer = async () => null
    d.pingServerWithIP = async () => null

    assert.equal(await d.findServer(), null)
    assert.equal(d.status, 'failed')
    assert.match(d.log.join('\n'), /ERROR: No server found on network/)
    assert.match(d.log.join('\n'), /open the shared link directly instead of looty.html/)
  } finally {
    restore()
  }
})

test('fileBrowser loadFiles clears state when fetch fails', async () => {
  const restore = setGlobals({
    window: { location: { protocol: 'http:', host: '127.0.0.1:41111' } },
    fetch: async () => {
      throw new Error('network down')
    },
    console: { log: () => {}, error: () => {} },
  })

  try {
    const browser = fileBrowser()
    browser.files = [{ name: 'stale' }]
    await browser.loadFiles('docs')

    assert.equal(browser.currentPath, 'docs')
    assert.deepEqual(browser.files, [])
    assert.equal(browser.loading, false)
    assert.equal(browser.refreshing, false)
  } finally {
    restore()
  }
})

test('clipboard connectWebSocket exposes connection failure', () => {
  const restore = setGlobals({
    window: { location: { protocol: 'http:', host: '127.0.0.1:41111' } },
    WebSocket: class {
      constructor() {
        throw new Error('socket refused')
      }
    },
    console: { log: () => {}, error: () => {} },
  })

  try {
    const panel = clipboardPanel()
    panel.connectWebSocket()
    assert.equal(panel.wsError, 'Cannot connect to server')
    assert.equal(panel.wsConnected, false)
  } finally {
    restore()
  }
})
