// Main application orchestrator - imports and wires up all modules
import { openFileDialog, handleBackupDatabase, handleRestoreDatabase, handleDistributeGroups } from './admin/file-handler.js';
import { handleEditConfig } from './admin/config-editor.js';
import { handleShowGroups } from './groups/groups.js';
import { handleShowStations, handleShowStationsForGroup, handleShowInputOverview } from './stations/stations.js';
import { handleGroupEvaluation, handleOrtsverbandEvaluation, toggleOVScores } from './evaluations/evaluations.js';
import { 
    handleGeneratePDF, 
    handleGenerateGroupEvaluationPDF, 
    handleGenerateOrtsverbandEvaluationPDF, 
    handleGenerateCertificates,
    handleGenerateOVCertificates
} from './reports/pdf-handlers.js';
import { setStatus, output, tabs, btnShow, btnDistribute, btnStations, btnOverview, btnPDF, btnBackup, sectionAusgabe, ausgabeDropdown, setEvalButtonsEnabled } from './shared/dom.js';

// Load configuration and run startup DB check
(async () => {
    try {
        window.appConfig = await window.go.main.App.GetConfig();
    } catch (e) {
        window.appConfig = { scoreMin: 100, scoreMax: 1200, maxGroupSize: 8 };
    }

    try {
        const startup = await window.go.main.App.CheckStartup();
        if (startup.exists) {
            await _handleStartupDBChoice(startup.dbName);
        }
    } catch (e) {
        console.error('Startup check failed:', e);
    }
})();

async function _handleStartupDBChoice(dbName) {
    const choice = await new Promise(resolve => {
        const overlay = document.createElement('div');
        overlay.style.cssText = 'position:fixed;inset:0;background:rgba(0,0,0,.55);display:flex;align-items:center;justify-content:center;z-index:9999';

        const box = document.createElement('div');
        box.style.cssText = 'background:#fff;border-radius:10px;padding:32px 36px;max-width:460px;width:90%;box-shadow:0 8px 32px rgba(0,0,0,.3);font-family:Arial,sans-serif';

        box.innerHTML = `
            <h2 style="margin:0 0 12px;font-size:1.2em;color:#333">Datenbank gefunden</h2>
            <p style="margin:0 0 20px;color:#555;line-height:1.5">
                Die Datenbankdatei <strong>${dbName}</strong> ist bereits vorhanden.<br>
                Möchten Sie mit den vorhandenen Daten weiterarbeiten oder eine neue, leere Datenbank erstellen?
            </p>
            <p style="margin:0 0 24px;font-size:0.85em;color:#888">
                Bei Auswahl von <em>„Neu starten"</em> wird die bestehende Datenbank automatisch gesichert.
            </p>
            <div style="display:flex;gap:12px;justify-content:flex-end">
                <button id="_btnFresh" style="padding:10px 20px;background:#e53935;color:#fff;border:none;border-radius:6px;cursor:pointer;font-weight:600">Neu starten</button>
                <button id="_btnKeep"  style="padding:10px 20px;background:#1976d2;color:#fff;border:none;border-radius:6px;cursor:pointer;font-weight:600">Weiterarbeiten</button>
            </div>`;

        overlay.appendChild(box);
        document.body.appendChild(overlay);

        box.querySelector('#_btnKeep').addEventListener('click', () => { document.body.removeChild(overlay); resolve('keep'); });
        box.querySelector('#_btnFresh').addEventListener('click', () => { document.body.removeChild(overlay); resolve('fresh'); });
    });

    if (choice === 'keep') {
        const result = await window.go.main.App.UseExistingDB();
        if (result.status === 'ok') {
            setStatus(`Vorhandene Datenbank geladen (${result.count} Teilnehmende).`, 'success');
            // Re-enable buttons that require a loaded DB
            if (btnBackup) btnBackup.disabled = false;
            if (btnDistribute) btnDistribute.disabled = false;
            if (btnShow) btnShow.disabled = false;
            if (btnStations) btnStations.disabled = false;
            if (btnOverview) btnOverview.disabled = false;
            if (btnPDF) btnPDF.disabled = false;
            setEvalButtonsEnabled(true);
            output.style.display = 'block';
            output.textContent = `✔ Vorhandene Daten geladen (${result.count} Teilnehmende).`;
        } else {
            setStatus('FEHLER beim Öffnen der Datenbank: ' + result.message, 'error');
        }
    } else {
        const result = await window.go.main.App.ResetToFreshDB();
        if (result.status === 'ok') {
            const msg = result.backupPath
                ? `Neue Datenbank erstellt. Backup gespeichert unter: ${result.backupPath}`
                : 'Neue leere Datenbank erstellt.';
            setStatus(msg, 'success');
        } else {
            setStatus('FEHLER beim Zurücksetzen: ' + result.message, 'error');
        }
    }
}

// Expose functions to window object for onclick handlers
window.openFileDialog = openFileDialog;
window.handleBackupDatabase = handleBackupDatabase;
window.handleRestoreDatabase = handleRestoreDatabase;
window.handleDistributeGroups = handleDistributeGroups;
window.handleEditConfig = handleEditConfig;
window.handleShowGroups = handleShowGroups;
window.handleShowStations = handleShowStations;
window.handleShowStationsForGroup = handleShowStationsForGroup;
window.handleShowInputOverview = handleShowInputOverview;
window.handleEvaluation = handleGroupEvaluation;
window.handleOrtsverbandEvaluation = handleOrtsverbandEvaluation;
window.toggleOVScores = toggleOVScores;
window.handleGeneratePDF = handleGeneratePDF;
window.handleGenerateGroupEvaluationPDF = handleGenerateGroupEvaluationPDF;
window.handleGenerateOrtsverbandEvaluationPDF = handleGenerateOrtsverbandEvaluationPDF;
window.handleGenerateCertificates = handleGenerateCertificates;
window.handleGenerateOVCertificates = handleGenerateOVCertificates;
