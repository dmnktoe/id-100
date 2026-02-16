/**
 * Tests for admin-dashboard module
 */
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { 
  initAdminDashboard,
  resetToken,
  deactivateToken,
  updateQuota,
  downloadQR,
  copyUploadURL,
  markBagRequestDone,
  deleteContribution
} from '../lib/admin-dashboard'

describe('initAdminDashboard', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    global.fetch = vi.fn()
    window.alert = vi.fn()
    window.confirm = vi.fn(() => true)
    window.location.reload = vi.fn() as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should initialize without errors when elements are missing', () => {
    document.body.innerHTML = '<div></div>'
    expect(() => initAdminDashboard()).not.toThrow()
  })

  it('should setup form submission handler', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        token_id: 'test-123',
        qr_url: '/qr/test-123',
        upload_url: 'http://example.com/upload?token=abc'
      })
    })
    global.fetch = mockFetch

    document.body.innerHTML = `
      <form id="createTokenForm">
        <input id="bagName" value="Werkzeug #1" />
        <input id="maxUploads" value="100" />
        <button type="submit">Submit</button>
      </form>
      <div id="createResult"></div>
    `

    initAdminDashboard()

    const form = document.getElementById('createTokenForm') as HTMLFormElement
    const event = new Event('submit', { bubbles: true, cancelable: true })
    
    form.dispatchEvent(event)

    await vi.waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('/admin/tokens', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ bag_name: 'Werkzeug #1', max_uploads: 100 })
      })
    })
  })

  it('should show success message after creating token', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        token_id: 'test-123',
        qr_url: '/qr/test-123',
        upload_url: 'http://example.com/upload?token=abc'
      })
    })
    global.fetch = mockFetch

    document.body.innerHTML = `
      <form id="createTokenForm">
        <input id="bagName" value="Werkzeug #1" />
        <input id="maxUploads" value="100" />
        <button type="submit">Submit</button>
      </form>
      <div id="createResult"></div>
    `

    initAdminDashboard()

    const form = document.getElementById('createTokenForm') as HTMLFormElement
    const resultDiv = document.getElementById('createResult') as HTMLDivElement
    const event = new Event('submit', { bubbles: true, cancelable: true })
    
    form.dispatchEvent(event)

    await vi.waitFor(() => {
      expect(resultDiv.style.display).toBe('block')
      expect(resultDiv.innerHTML).toContain('Token erstellt!')
      expect(resultDiv.innerHTML).toContain('test-123')
    })
  })

  it('should handle error response when creating token', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: 'Invalid request' })
    })
    global.fetch = mockFetch

    document.body.innerHTML = `
      <form id="createTokenForm">
        <input id="bagName" value="Werkzeug #1" />
        <input id="maxUploads" value="100" />
        <button type="submit">Submit</button>
      </form>
      <div id="createResult"></div>
    `

    initAdminDashboard()

    const form = document.getElementById('createTokenForm') as HTMLFormElement
    const event = new Event('submit', { bubbles: true, cancelable: true })
    
    form.dispatchEvent(event)

    await vi.waitFor(() => {
      expect(window.alert).toHaveBeenCalledWith('Fehler: Invalid request')
    })
  })
})

describe('resetToken', () => {
  beforeEach(() => {
    global.fetch = vi.fn()
    window.alert = vi.fn()
    window.confirm = vi.fn(() => true)
    window.location.reload = vi.fn() as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should reset token when confirmed', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ message: 'Token reset' })
    })
    global.fetch = mockFetch

    await resetToken(1, 'Test Token')

    expect(window.confirm).toHaveBeenCalledWith('Test Token wirklich zurücksetzen? Der Upload-Counter wird auf 0 gesetzt.')
    expect(mockFetch).toHaveBeenCalledWith('/admin/tokens/1/reset', { method: 'POST' })
    expect(window.alert).toHaveBeenCalledWith('Token reset')
    expect(window.location.reload).toHaveBeenCalled()
  })

  it('should not reset token when cancelled', async () => {
    window.confirm = vi.fn(() => false)
    const mockFetch = vi.fn()
    global.fetch = mockFetch

    await resetToken(1, 'Test Token')

    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('should handle error when resetting token', async () => {
    const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'))
    global.fetch = mockFetch

    await resetToken(1, 'Test Token')

    expect(window.alert).toHaveBeenCalledWith('Fehler: Network error')
  })
})

describe('deactivateToken', () => {
  beforeEach(() => {
    global.fetch = vi.fn()
    window.alert = vi.fn()
    window.confirm = vi.fn(() => true)
    window.location.reload = vi.fn() as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should deactivate token when confirmed', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ status: 'deactivated' })
    })
    global.fetch = mockFetch

    await deactivateToken(1, 'Test Token')

    expect(window.confirm).toHaveBeenCalledWith('Test Token wirklich deaktivieren?')
    expect(mockFetch).toHaveBeenCalledWith('/admin/tokens/1/deactivate', { method: 'POST' })
    expect(window.alert).toHaveBeenCalledWith('deactivated')
    expect(window.location.reload).toHaveBeenCalled()
  })
})

describe('updateQuota', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    global.fetch = vi.fn()
    window.alert = vi.fn()
    window.location.reload = vi.fn() as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should update quota with valid value', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      json: () => Promise.resolve({ max_uploads: 200 })
    })
    global.fetch = mockFetch

    document.body.innerHTML = '<input id="quota-1" value="200" />'

    await updateQuota(1)

    expect(mockFetch).toHaveBeenCalledWith('/admin/tokens/1/quota', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ max_uploads: 200 })
    })
    expect(window.alert).toHaveBeenCalledWith('✅ Kontingent auf 200 aktualisiert')
    expect(window.location.reload).toHaveBeenCalled()
  })

  it('should reject quota value of 0', async () => {
    const mockFetch = vi.fn()
    global.fetch = mockFetch

    document.body.innerHTML = '<input id="quota-1" value="0" />'

    await updateQuota(1)

    expect(window.alert).toHaveBeenCalledWith('Kontingent muss größer als 0 sein')
    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('should reject negative quota value', async () => {
    const mockFetch = vi.fn()
    global.fetch = mockFetch

    document.body.innerHTML = '<input id="quota-1" value="-5" />'

    await updateQuota(1)

    expect(window.alert).toHaveBeenCalledWith('Kontingent muss größer als 0 sein')
    expect(mockFetch).not.toHaveBeenCalled()
  })
})

describe('downloadQR', () => {
  beforeEach(() => {
    vi.spyOn(window, 'open').mockImplementation(() => null)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should open QR code in new window with SVG format', () => {
    downloadQR(1, 'Test Token', 'svg')
    expect(window.open).toHaveBeenCalledWith('/admin/tokens/1/qr?format=svg', '_blank')
  })

  it('should open QR code in new window with PNG format', () => {
    downloadQR(1, 'Test Token', 'png')
    expect(window.open).toHaveBeenCalledWith('/admin/tokens/1/qr?format=png', '_blank')
  })
})

describe('copyUploadURL', () => {
  beforeEach(() => {
    window.alert = vi.fn()
    window.prompt = vi.fn(() => null)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should copy URL to clipboard', async () => {
    const mockWriteText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: mockWriteText },
      writable: true,
      configurable: true
    })

    await copyUploadURL('abc123', 'Test Token')

    const expectedUrl = `${location.origin}/upload?token=abc123`
    expect(mockWriteText).toHaveBeenCalledWith(expectedUrl)
    expect(window.alert).toHaveBeenCalledWith('URL kopiert: ' + expectedUrl)
  })

  it('should fallback to prompt when clipboard fails', async () => {
    const mockWriteText = vi.fn().mockRejectedValue(new Error('Clipboard error'))
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: mockWriteText },
      writable: true,
      configurable: true
    })

    await copyUploadURL('abc123', 'Test Token')

    const expectedUrl = `${location.origin}/upload?token=abc123`
    expect(window.prompt).toHaveBeenCalledWith('URL kopieren:', expectedUrl)
  })

  it('should encode token in URL', async () => {
    const mockWriteText = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: mockWriteText },
      writable: true,
      configurable: true
    })

    await copyUploadURL('token with spaces', 'Test Token')

    const expectedUrl = `${location.origin}/upload?token=token%20with%20spaces`
    expect(mockWriteText).toHaveBeenCalledWith(expectedUrl)
  })
})

describe('markBagRequestDone', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    global.fetch = vi.fn()
    window.alert = vi.fn()
    window.confirm = vi.fn(() => true)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should mark request as done when confirmed', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ status: 'ok' })
    })
    global.fetch = mockFetch

    document.body.innerHTML = `
      <ul>
        <li id="testLi">
          <button id="testBtn">Mark Done</button>
        </li>
      </ul>
    `

    const btn = document.getElementById('testBtn') as HTMLElement

    await markBagRequestDone(1, btn)

    expect(window.confirm).toHaveBeenCalledWith('Als erledigt markieren?')
    expect(mockFetch).toHaveBeenCalledWith('/admin/werkzeug-anfragen/1/complete', { method: 'POST' })
    
    const li = document.getElementById('testLi') as HTMLLIElement
    expect(li.style.opacity).toBe('0.8')
    expect(li.textContent).toContain('✅ Erledigt')
  })

  it('should not mark request as done when cancelled', async () => {
    window.confirm = vi.fn(() => false)
    const mockFetch = vi.fn()
    global.fetch = mockFetch

    const btn = document.createElement('button')
    await markBagRequestDone(1, btn)

    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('should show error when request fails', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: 'Server error' })
    })
    global.fetch = mockFetch

    const btn = document.createElement('button')
    await markBagRequestDone(1, btn)

    expect(window.alert).toHaveBeenCalledWith('Fehler: Server error')
  })
})

describe('deleteContribution', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    global.fetch = vi.fn()
    window.alert = vi.fn()
    window.confirm = vi.fn(() => true)
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.restoreAllMocks()
    vi.useRealTimers()
  })

  it('should delete contribution when confirmed', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ status: 'deleted' })
    })
    global.fetch = mockFetch

    document.body.innerHTML = `
      <div id="contrib-1">
        <button id="deleteBtn">Delete</button>
      </div>
    `

    const btn = document.getElementById('deleteBtn') as HTMLElement

    await deleteContribution(1, btn)

    expect(window.confirm).toHaveBeenCalled()
    expect(mockFetch).toHaveBeenCalledWith('/admin/contributions/1/delete', { method: 'POST' })
    expect(window.alert).toHaveBeenCalledWith('Contribution erfolgreich gelöscht')

    const card = document.getElementById('contrib-1')
    expect(card?.style.opacity).toBe('0')
    expect(card?.style.transition).toBe('opacity 0.3s')

    // Fast-forward time
    vi.advanceTimersByTime(300)
    expect(document.getElementById('contrib-1')).toBeNull()
  })

  it('should not delete when cancelled', async () => {
    window.confirm = vi.fn(() => false)
    const mockFetch = vi.fn()
    global.fetch = mockFetch

    const btn = document.createElement('button')
    await deleteContribution(1, btn)

    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('should show error when deletion fails', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: 'Not found' })
    })
    global.fetch = mockFetch

    const btn = document.createElement('button')
    await deleteContribution(1, btn)

    expect(window.alert).toHaveBeenCalledWith('Fehler: Not found')
  })
})
