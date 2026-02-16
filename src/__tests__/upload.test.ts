/**
 * Tests for upload module
 */
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { initUpload, deleteSessionUpload, endSession } from '../lib/upload'

describe('initUpload', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    global.fetch = vi.fn()
    window.alert = vi.fn()
    // Mock location
    delete (window as any).location
    window.location = {
      ...window.location,
      search: '',
      pathname: '/upload',
      href: '',
      reload: vi.fn()
    } as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should initialize without errors when elements are missing', () => {
    document.body.innerHTML = '<div></div>'
    expect(() => initUpload()).not.toThrow()
  })

  it('should setup drag and drop functionality', () => {
    document.body.innerHTML = `
      <div id="drop-zone">
        <span id="drop-text">Drop here</span>
      </div>
      <input type="file" id="fileInput" />
      <img id="preview" />
    `

    initUpload()

    const dropZone = document.getElementById('drop-zone') as HTMLElement
    
    // Test dragover
    const dragoverEvent = new DragEvent('dragover', { bubbles: true, cancelable: true })
    dropZone.dispatchEvent(dragoverEvent)
    expect(dropZone.classList.contains('active')).toBe(true)

    // Test dragleave
    const dragleaveEvent = new DragEvent('dragleave', { bubbles: true })
    dropZone.dispatchEvent(dragleaveEvent)
    expect(dropZone.classList.contains('active')).toBe(false)
  })

  it('should update preview when file is selected', () => {
    document.body.innerHTML = `
      <div id="drop-zone">
        <span id="drop-text">Drop here</span>
      </div>
      <input type="file" id="fileInput" />
      <img id="preview" style="display: none;" />
    `

    initUpload()

    const fileInput = document.getElementById('fileInput') as HTMLInputElement
    const preview = document.getElementById('preview') as HTMLImageElement
    const dropText = document.getElementById('drop-text') as HTMLElement

    // Create a mock file
    const file = new File(['test'], 'test.png', { type: 'image/png' })
    const dataTransfer = new DataTransfer()
    dataTransfer.items.add(file)
    fileInput.files = dataTransfer.files

    // Trigger change event
    const changeEvent = new Event('change', { bubbles: true })
    fileInput.dispatchEvent(changeEvent)

    // FileReader is async, need to wait
    setTimeout(() => {
      expect(preview.style.display).toBe('block')
      expect(dropText.style.display).toBe('none')
    }, 100)
  })

  it('should update character count for comment input', () => {
    document.body.innerHTML = `
      <div id="drop-zone">
        <span id="drop-text">Drop here</span>
      </div>
      <input type="file" id="fileInput" />
      <img id="preview" />
      <input type="text" id="commentInput" />
      <span id="charCount">0</span>
    `

    initUpload()

    const commentInput = document.getElementById('commentInput') as HTMLInputElement
    const charCount = document.getElementById('charCount') as HTMLElement

    commentInput.value = 'Test comment'
    const inputEvent = new Event('input', { bubbles: true })
    commentInput.dispatchEvent(inputEvent)

    expect(charCount.textContent).toBe('12')
  })

  it('should disable submit button on form submission', () => {
    document.body.innerHTML = `
      <div id="drop-zone">
        <span id="drop-text">Drop here</span>
      </div>
      <input type="file" id="fileInput" />
      <img id="preview" />
      <form id="uploadForm">
        <button type="submit" id="submitBtn">Hochladen</button>
      </form>
    `

    initUpload()

    const form = document.getElementById('uploadForm') as HTMLFormElement
    const submitBtn = document.getElementById('submitBtn') as HTMLButtonElement

    const submitEvent = new Event('submit', { bubbles: true, cancelable: true })
    submitEvent.preventDefault() // Prevent actual form submission
    form.dispatchEvent(submitEvent)

    expect(submitBtn.disabled).toBe(true)
    expect(submitBtn.innerText).toBe('optimiere...')
  })

  it('should prefill derive number from URL parameter', () => {
    window.location.search = '?number=42&token=abc'

    document.body.innerHTML = `
      <div id="drop-zone">
        <span id="drop-text">Drop here</span>
      </div>
      <input type="file" id="fileInput" />
      <img id="preview" />
      <select id="deriveInput">
        <option value="">Select</option>
        <option value="42">ID 42</option>
      </select>
    `

    initUpload()

    const deriveInput = document.getElementById('deriveInput') as HTMLSelectElement
    expect(deriveInput.value).toBe('42')
  })

  it('should not prefill if option is disabled', () => {
    window.location.search = '?number=42&token=abc'

    document.body.innerHTML = `
      <div id="drop-zone">
        <span id="drop-text">Drop here</span>
      </div>
      <input type="file" id="fileInput" />
      <img id="preview" />
      <select id="deriveInput">
        <option value="">Select</option>
        <option value="42" disabled>ID 42 (already uploaded)</option>
      </select>
    `

    initUpload()

    const deriveInput = document.getElementById('deriveInput') as HTMLSelectElement
    expect(deriveInput.value).toBe('')
  })

  it('should handle post-upload success state', () => {
    window.location.search = '?uploaded=1&token=abc'

    document.body.innerHTML = `
      <div class="container upload">
        <div id="drop-zone">
          <span id="drop-text" style="display: none;">Drop here</span>
        </div>
        <input type="file" id="fileInput" />
        <img id="preview" style="display: block;" src="data:image/png;base64,test" />
        <select id="deriveInput">
          <option value="42">ID 42</option>
        </select>
        <button id="submitBtn" disabled>optimiere...</button>
        <div class="session-uploads"></div>
      </div>
    `

    initUpload()

    const deriveInput = document.getElementById('deriveInput') as HTMLSelectElement
    const preview = document.getElementById('preview') as HTMLImageElement
    const dropText = document.getElementById('drop-text') as HTMLElement
    const submitBtn = document.getElementById('submitBtn') as HTMLButtonElement

    expect(deriveInput.value).toBe('')
    expect(preview.src).toContain('#') // URL will be something like http://localhost:3000/upload#
    expect(preview.style.display).toBe('none')
    expect(dropText.style.display).toBe('block')
    expect(submitBtn.disabled).toBe(false)
    expect(submitBtn.innerText).toBe('Hochladen')

    const uploadResult = document.getElementById('uploadResult')
    expect(uploadResult).toBeTruthy()
    expect(uploadResult?.innerText).toBe('Upload erfolgreich!')
    expect(uploadResult?.style.color).toBe('#2e7d32')
  })

  it('should click file input when drop zone is clicked', () => {
    document.body.innerHTML = `
      <div id="drop-zone">
        <span id="drop-text">Drop here</span>
      </div>
      <input type="file" id="fileInput" />
      <img id="preview" />
    `

    initUpload()

    const dropZone = document.getElementById('drop-zone') as HTMLElement
    const fileInput = document.getElementById('fileInput') as HTMLInputElement

    const clickSpy = vi.spyOn(fileInput, 'click')

    dropZone.click()

    expect(clickSpy).toHaveBeenCalled()
  })
})

describe('deleteSessionUpload', () => {
  beforeEach(() => {
    global.fetch = vi.fn()
    window.alert = vi.fn()
    window.confirm = vi.fn(() => true)
    delete (window as any).location
    window.location = {
      ...window.location,
      search: '?token=test123',
      reload: vi.fn()
    } as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should delete session upload when confirmed', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ status: 'deleted' })
    })
    global.fetch = mockFetch

    const btn = document.createElement('button')

    await deleteSessionUpload(1, btn)

    expect(window.confirm).toHaveBeenCalledWith('Upload wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.')
    expect(mockFetch).toHaveBeenCalledWith('/upload/contributions/1/delete?token=test123', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    })
    expect(window.location.reload).toHaveBeenCalled()
  })

  it('should not delete when cancelled', async () => {
    window.confirm = vi.fn(() => false)
    const mockFetch = vi.fn()
    global.fetch = mockFetch

    const btn = document.createElement('button')

    await deleteSessionUpload(1, btn)

    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('should show error when deletion fails', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: 'Not authorized' })
    })
    global.fetch = mockFetch

    const btn = document.createElement('button')

    await deleteSessionUpload(1, btn)

    expect(window.alert).toHaveBeenCalledWith('Fehler: Not authorized')
  })

  it('should handle network error', async () => {
    const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'))
    global.fetch = mockFetch

    const btn = document.createElement('button')

    await deleteSessionUpload(1, btn)

    expect(window.alert).toHaveBeenCalledWith('Fehler: Network error')
  })
})

describe('endSession', () => {
  beforeEach(() => {
    global.fetch = vi.fn()
    window.alert = vi.fn()
    window.confirm = vi.fn(() => true)
    delete (window as any).location
    window.location = {
      ...window.location,
      search: '?token=test123',
      href: ''
    } as any
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('should end session when confirmed', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ message: 'Session ended' })
    })
    global.fetch = mockFetch

    await endSession()

    expect(window.confirm).toHaveBeenCalledWith('Session wirklich beenden? Das Werkzeug wird zurückgesetzt und kann an den nächsten Spieler weitergegeben werden.')
    expect(mockFetch).toHaveBeenCalledWith('/upload/end-session?token=test123', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' }
    })
    expect(window.alert).toHaveBeenCalledWith('Session ended')
    expect(window.location.href).toBe('/upload?token=test123')
  })

  it('should not end session when cancelled', async () => {
    window.confirm = vi.fn(() => false)
    const mockFetch = vi.fn()
    global.fetch = mockFetch

    await endSession()

    expect(mockFetch).not.toHaveBeenCalled()
  })

  it('should show error when ending session fails', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: 'Session not found' })
    })
    global.fetch = mockFetch

    await endSession()

    expect(window.alert).toHaveBeenCalledWith('Fehler: Session not found')
  })

  it('should handle network error', async () => {
    const mockFetch = vi.fn().mockRejectedValue(new Error('Network error'))
    global.fetch = mockFetch

    await endSession()

    expect(window.alert).toHaveBeenCalledWith('Fehler: Network error')
  })
})
