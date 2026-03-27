// Config editor — opens a modal for editing config.toml and certificate_layout.json in-app.
import { setStatus } from '../shared/dom.js';

export async function handleEditConfig() {
    setStatus('Konfiguration wird geladen...', 'info');

    try {
        const [configResult, layoutResult] = await Promise.all([
            window.go.main.App.GetConfigRaw(),
            window.go.main.App.GetCertLayoutRaw(),
        ]);

        if (configResult.status === 'error') {
            setStatus('FEHLER: ' + configResult.message, 'error');
            return;
        }
        if (layoutResult.status === 'error') {
            setStatus('FEHLER: ' + layoutResult.message, 'error');
            return;
        }

        _openModal(configResult.content, layoutResult.content);
        setStatus('Konfiguration geladen.', 'info');
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

function _openModal(configContent, layoutContent) {
    const overlay = document.createElement('div');
    overlay.id = 'config-editor-overlay';
    overlay.style.cssText = 'position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.5);z-index:1000;display:flex;justify-content:center;align-items:center;';

    overlay.innerHTML = `
        <div style="background:white;border-radius:12px;padding:30px;width:720px;max-width:95vw;max-height:90vh;display:flex;flex-direction:column;box-shadow:0 20px 60px rgba(0,0,0,0.3);">
            <h2 style="margin:0 0 16px 0;color:#333;">⚙️ Konfiguration bearbeiten</h2>

            <!-- Tab bar -->
            <div style="display:flex;gap:0;border-bottom:2px solid #e0e0e0;margin-bottom:16px;">
                <button id="cfg-tab-config"
                    onclick="window._switchConfigTab('config')"
                    style="padding:8px 20px;border:none;border-bottom:2px solid #667eea;margin-bottom:-2px;background:none;cursor:pointer;font-weight:600;font-size:13px;color:#667eea;">
                    config.toml
                </button>
                <button id="cfg-tab-layout"
                    onclick="window._switchConfigTab('layout')"
                    style="padding:8px 20px;border:none;border-bottom:2px solid transparent;margin-bottom:-2px;background:none;cursor:pointer;font-weight:400;font-size:13px;color:#888;">
                    certificate_layout.json
                </button>
            </div>

            <!-- Tab: config.toml -->
            <div id="cfg-panel-config" style="display:flex;flex-direction:column;flex:1;">
                <p style="margin:0 0 10px 0;color:#666;font-size:13px;">
                    Bearbeiten Sie die <code>config.toml</code>. Zeilen mit <code>#</code> sind Kommentare.
                    Ungültige TOML-Syntax wird vor dem Speichern abgewiesen.
                </p>
                <textarea id="config-editor-textarea"
                    spellcheck="false"
                    style="flex:1;min-height:320px;font-family:'Courier New',monospace;font-size:13px;line-height:1.6;
                           padding:12px;border:2px solid #ddd;border-radius:6px;resize:vertical;color:#333;
                           outline:none;transition:border-color 0.2s;"
                    onfocus="this.style.borderColor='#667eea'"
                    onblur="this.style.borderColor='#ddd'"
                ></textarea>
            </div>

            <!-- Tab: certificate_layout.json -->
            <div id="cfg-panel-layout" style="display:none;flex-direction:column;flex:1;">
                <p style="margin:0 0 10px 0;color:#666;font-size:13px;">
                    Bearbeiten Sie das <code>certificate_layout.json</code>. Dieses JSON steuert Position,
                    Schriftgröße und Farbe aller Elemente auf den Urkunden.
                    Ungültiges JSON wird abgewiesen.
                </p>
                <textarea id="layout-editor-textarea"
                    spellcheck="false"
                    style="flex:1;min-height:320px;font-family:'Courier New',monospace;font-size:12px;line-height:1.5;
                           padding:12px;border:2px solid #ddd;border-radius:6px;resize:vertical;color:#333;
                           outline:none;transition:border-color 0.2s;"
                    onfocus="this.style.borderColor='#667eea'"
                    onblur="this.style.borderColor='#ddd'"
                ></textarea>
            </div>

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
    document.getElementById('config-editor-textarea').value = configContent;
    document.getElementById('layout-editor-textarea').value = layoutContent;
    // Focus the active (first) tab's textarea
    document.getElementById('config-editor-textarea').focus();
}

window._switchConfigTab = function (tab) {
    const isConfig = tab === 'config';
    document.getElementById('cfg-panel-config').style.display = isConfig ? 'flex' : 'none';
    document.getElementById('cfg-panel-layout').style.display = isConfig ? 'none' : 'flex';

    const tabConfig = document.getElementById('cfg-tab-config');
    const tabLayout = document.getElementById('cfg-tab-layout');
    if (isConfig) {
        tabConfig.style.borderBottomColor = '#667eea';
        tabConfig.style.fontWeight = '600';
        tabConfig.style.color = '#667eea';
        tabLayout.style.borderBottomColor = 'transparent';
        tabLayout.style.fontWeight = '400';
        tabLayout.style.color = '#888';
    } else {
        tabLayout.style.borderBottomColor = '#667eea';
        tabLayout.style.fontWeight = '600';
        tabLayout.style.color = '#667eea';
        tabConfig.style.borderBottomColor = 'transparent';
        tabConfig.style.fontWeight = '400';
        tabConfig.style.color = '#888';
    }
    // Hide any previous error when switching tabs
    const errorDiv = document.getElementById('config-editor-error');
    if (errorDiv) errorDiv.style.display = 'none';
};

window._closeConfigEditor = function () {
    const overlay = document.getElementById('config-editor-overlay');
    if (overlay) overlay.remove();
    setStatus('Bearbeitung abgebrochen.', 'info');
};

window._saveConfig = async function () {
    const errorDiv = document.getElementById('config-editor-error');
    errorDiv.style.display = 'none';

    // Determine which tab is active
    const configPanel = document.getElementById('cfg-panel-config');
    const isConfigTab = configPanel && configPanel.style.display !== 'none';

    try {
        let result;
        if (isConfigTab) {
            const ta = document.getElementById('config-editor-textarea');
            if (!ta) return;
            result = await window.go.main.App.SaveConfigRaw(ta.value);
            if (result.status === 'ok') {
                // Refresh the in-memory config so score bounds, etc. stay current
                try { window.appConfig = await window.go.main.App.GetConfig(); } catch (_) { /* non-fatal */ }
            }
        } else {
            const ta = document.getElementById('layout-editor-textarea');
            if (!ta) return;
            result = await window.go.main.App.SaveCertLayoutRaw(ta.value);
        }

        if (result.status === 'error') {
            errorDiv.textContent = '⚠️ ' + result.message;
            errorDiv.style.display = 'block';
            return;
        }

        const overlay = document.getElementById('config-editor-overlay');
        if (overlay) overlay.remove();
        setStatus('✅ ' + result.message, 'success');

    } catch (err) {
        errorDiv.textContent = '⚠️ ' + err;
        errorDiv.style.display = 'block';
    }
};
