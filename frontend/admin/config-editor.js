// Config editor — opens a modal for editing config.toml in-app
import { setStatus } from '../shared/dom.js';

export async function handleEditConfig() {
    setStatus('Konfiguration wird geladen...', 'info');

    try {
        const result = await window.go.main.App.GetConfigRaw();

        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            return;
        }

        _openModal(result.content);
        setStatus('Konfiguration geladen.', 'info');
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

function _openModal(content) {
    const overlay = document.createElement('div');
    overlay.id = 'config-editor-overlay';
    overlay.style.cssText = 'position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.5);z-index:1000;display:flex;justify-content:center;align-items:center;';

    overlay.innerHTML = `
        <div style="background:white;border-radius:12px;padding:30px;width:680px;max-width:95vw;max-height:85vh;display:flex;flex-direction:column;box-shadow:0 20px 60px rgba(0,0,0,0.3);">
            <h2 style="margin:0 0 8px 0;color:#333;">⚙️ Konfiguration bearbeiten</h2>
            <p style="margin:0 0 16px 0;color:#666;font-size:13px;">
                Bearbeiten Sie die <code>config.toml</code>. Zeilen mit <code>#</code> sind Kommentare.
                Ungültige TOML-Syntax wird vor dem Speichern abgewiesen.
            </p>
            <textarea id="config-editor-textarea"
                spellcheck="false"
                style="flex:1;min-height:340px;font-family:'Courier New',monospace;font-size:13px;line-height:1.6;
                       padding:12px;border:2px solid #ddd;border-radius:6px;resize:vertical;color:#333;
                       outline:none;transition:border-color 0.2s;"
                onfocus="this.style.borderColor='#667eea'"
                onblur="this.style.borderColor='#ddd'"
            ></textarea>
            <div id="config-editor-error" style="display:none;margin-top:10px;padding:10px 14px;background:#ffebee;border-left:4px solid #f44336;border-radius:4px;color:#c62828;font-size:13px;"></div>
            <div style="display:flex;justify-content:flex-end;gap:10px;margin-top:16px;">
                <button onclick="window._closeConfigEditor()"
                    style="padding:10px 22px;background:#e0e0e0;color:#333;border:none;border-radius:6px;cursor:pointer;font-weight:600;font-size:13px;">
                    Abbrechen
                </button>
                <button onclick="window._saveConfig()"
                    style="padding:10px 22px;background:linear-gradient(135deg,#667eea 0%,#764ba2 100%);color:white;border:none;border-radius:6px;cursor:pointer;font-weight:600;font-size:13px;">
                    💾 Speichern
                </button>
            </div>
        </div>
    `;

    document.body.appendChild(overlay);
    const ta = document.getElementById('config-editor-textarea');
    ta.value = content;
    ta.focus();
}

window._closeConfigEditor = function () {
    const overlay = document.getElementById('config-editor-overlay');
    if (overlay) overlay.remove();
    setStatus('Bearbeitung abgebrochen.', 'info');
};

window._saveConfig = async function () {
    const ta = document.getElementById('config-editor-textarea');
    const errorDiv = document.getElementById('config-editor-error');
    if (!ta) return;

    errorDiv.style.display = 'none';

    try {
        const result = await window.go.main.App.SaveConfigRaw(ta.value);

        if (result.status === 'error') {
            errorDiv.textContent = '⚠️ ' + result.message;
            errorDiv.style.display = 'block';
            return;
        }

        const overlay = document.getElementById('config-editor-overlay');
        if (overlay) overlay.remove();
        setStatus('✅ ' + result.message, 'success');

        // Refresh the in-memory config so score bounds, etc. stay current
        try {
            window.appConfig = await window.go.main.App.GetConfig();
        } catch (_) { /* non-fatal */ }

    } catch (err) {
        errorDiv.textContent = '⚠️ ' + err;
        errorDiv.style.display = 'block';
    }
};
