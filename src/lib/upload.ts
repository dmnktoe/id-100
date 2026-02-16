/**
 * Upload page functionality
 * Handles file upload, drag & drop, preview, and session management
 */

/**
 * Initialize upload page functionality
 */
export function initUpload(): void {
  const dz = document.getElementById('drop-zone') as HTMLElement | null;
  const fi = document.getElementById('fileInput') as HTMLInputElement | null;
  const pr = document.getElementById('preview') as HTMLImageElement | null;
  const dt = document.getElementById('drop-text') as HTMLElement | null;

  if (!dz || !fi || !pr || !dt) {
    return; // Elements not found, probably not on upload page
  }

  // Click to select file
  dz.addEventListener('click', () => fi.click());

  // Drag over
  dz.addEventListener('dragover', (e: DragEvent) => {
    e.preventDefault();
    dz.classList.add('active');
  });

  // Drag leave
  dz.addEventListener('dragleave', () => dz.classList.remove('active'));

  // Drop
  dz.addEventListener('drop', (e: DragEvent) => {
    e.preventDefault();
    dz.classList.remove('active');
    if (e.dataTransfer?.files.length) {
      fi.files = e.dataTransfer.files;
      updatePreview(e.dataTransfer.files[0], pr, dt);
    }
  });

  // File input change
  fi.addEventListener('change', (e: Event) => {
    const target = e.target as HTMLInputElement;
    if (target.files?.length) {
      updatePreview(target.files[0], pr, dt);
    }
  });

  // Character counter for comment
  const commentInput = document.getElementById('commentInput') as HTMLInputElement | null;
  const charCount = document.getElementById('charCount') as HTMLElement | null;
  if (commentInput && charCount) {
    commentInput.addEventListener('input', (e: Event) => {
      const target = e.target as HTMLInputElement;
      charCount.textContent = target.value.length.toString();
    });
  }

  // Form submission - disable button and show loading
  const uploadForm = document.getElementById('uploadForm') as HTMLFormElement | null;
  if (uploadForm) {
    uploadForm.onsubmit = function() {
      const btn = document.getElementById('submitBtn') as HTMLButtonElement | null;
      if (btn) {
        btn.disabled = true;
        btn.innerText = "optimiere...";
      }
    };
  }

  // Handle prefill and post-upload state
  handleUploadState();
}

/**
 * Update file preview
 */
function updatePreview(file: File, pr: HTMLImageElement, dt: HTMLElement): void {
  const reader = new FileReader();
  reader.onload = (e: ProgressEvent<FileReader>) => {
    if (e.target?.result) {
      pr.src = e.target.result as string;
      pr.style.display = 'block';
      dt.style.display = 'none';
    }
  }
  reader.readAsDataURL(file);
}

/**
 * Handle upload state including prefill and post-upload success message
 */
function handleUploadState(): void {
  try {
    const params = new URLSearchParams(location.search);
    const d = params.get('number');
    if (d) {
      const el = document.getElementById('deriveInput') as HTMLSelectElement | null;
      if (el) {
        el.value = d;
        // don't leave a disabled (already-uploaded) option selected
        const selectedOpt = el.querySelector(`option[value="${d}"]`) as HTMLOptionElement | null;
        if (selectedOpt && selectedOpt.disabled) {
          el.value = '';
        }
      }
    }

    // If redirected after successful upload, clear selection and reset form
    const uploaded = params.get('uploaded');
    if (uploaded === '1') {
      const el = document.getElementById('deriveInput') as HTMLSelectElement | null;
      const fileInput = document.getElementById('fileInput') as HTMLInputElement | null;
      const preview = document.getElementById('preview') as HTMLImageElement | null;
      const dropText = document.getElementById('drop-text') as HTMLElement | null;
      const submitBtn = document.getElementById('submitBtn') as HTMLButtonElement | null;
      
      // clear selection
      if (el) el.value = '';
      
      // reset file input & preview
      if (fileInput) {
        try { 
          fileInput.value = ''; 
        } catch(e) { 
          /* ignore */ 
        }
      }
      if (preview) {
        preview.src = '#';
        preview.style.display = 'none';
      }
      if (dropText) dropText.style.display = 'block';
      if (submitBtn) {
        submitBtn.disabled = false;
        submitBtn.innerText = 'Hochladen';
      }
      
      // show temporary success message
      let result = document.getElementById('uploadResult');
      if (!result) {
        result = document.createElement('div');
        result.id = 'uploadResult';
        result.style.marginTop = '1rem';
        result.style.fontWeight = '500';
        const container = document.querySelector('.container.upload');
        const sessionUploads = document.querySelector('.session-uploads');
        if (container && sessionUploads) {
          container.insertBefore(result, sessionUploads);
        }
      }
      result.innerText = 'Upload erfolgreich!';
      result.style.color = '#2e7d32';
      result.style.display = 'block';
      setTimeout(() => {
        if (result) result.style.display = 'none';
      }, 4000);

      // Remove uploaded param from URL to avoid repeated behavior on refresh
      try {
        params.delete('uploaded');
        const newUrl = location.pathname + (params.toString() ? '?' + params.toString() : '');
        history.replaceState(null, '', newUrl);
      } catch(e) { 
        /* ignore */ 
      }
    }
  } catch (e) { 
    /* ignore */ 
  }
}

/**
 * Delete a session upload
 */
export async function deleteSessionUpload(id: number, _btn: HTMLElement): Promise<void> {
  if (!confirm('Upload wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.')) return;

  try {
    const params = new URLSearchParams(location.search);
    const token = params.get('token') || '';
    const response = await fetch(`/upload/contributions/${id}/delete?token=${encodeURIComponent(token)}`, { 
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    });
    const data = await response.json();

    if (response.ok) {
      // Reload page to update uploaded numbers and points
      location.reload();
    } else {
      alert('Fehler: ' + (data.error || 'Unbekannter Fehler'));
    }
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'Unbekannter Fehler';
    alert('Fehler: ' + errorMessage);
  }
}

/**
 * End session - return bag for next player
 */
export async function endSession(): Promise<void> {
  if (!confirm('Session wirklich beenden? Das Werkzeug wird zurückgesetzt und kann an den nächsten Spieler weitergegeben werden.')) return;

  try {
    const params = new URLSearchParams(location.search);
    const token = params.get('token') || '';
    const response = await fetch(`/upload/end-session?token=${encodeURIComponent(token)}`, { 
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    });
    const data = await response.json();

    if (response.ok) {
      alert(data.message || 'Session beendet');
      // Redirect to upload page without token to force re-entry
      location.href = '/upload?token=' + encodeURIComponent(token);
    } else {
      alert('Fehler: ' + (data.error || 'Unbekannter Fehler'));
    }
  } catch (err) {
    const errorMessage = err instanceof Error ? err.message : 'Unbekannter Fehler';
    alert('Fehler: ' + errorMessage);
  }
}

// Export functions to global window object for inline onclick handlers
if (typeof window !== 'undefined') {
  (window as any).deleteSessionUpload = deleteSessionUpload;
  (window as any).endSession = endSession;
}
