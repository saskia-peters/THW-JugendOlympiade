// File loading and management
import { setStatus, output, tabs, btnShow, btnDistribute, btnStations, btnOverview, btnPDF, sectionAusgabe, ausgabeDropdown, btnBackup, btnConvert, setEvalButtonsEnabled } from '../shared/dom.js';

export async function openFileDialog() {
    try {
        // Check if database has data
        const result = await window.go.main.App.CheckDB();
        
        if (result.hasData) {
            const confirmed = confirm(`Die Datenbank enthält ${result.count} Teilnehmende. Möchten Sie diese Daten verwerfen und eine neue Datei laden?`);
            if (!confirmed) {
                return;
            }
        }
        
        // Open file dialog and load file
        const uploadResult = await window.go.main.App.LoadFile();
        
        if (uploadResult.status === 'error') {
            alert(uploadResult.message);
            setStatus('FEHLER: ' + uploadResult.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Datei konnte nicht geladen werden. Bitte prüfen Sie die Fehlermeldung.';
        } else {
            setStatus(uploadResult.message, 'success');
            // Only the distribute button is enabled until groups are created
            btnDistribute.disabled = false;
            btnShow.disabled = true;
            btnStations.disabled = true;
            if (btnOverview) btnOverview.disabled = true;
            if (btnConvert) btnConvert.disabled = true;
            setEvalButtonsEnabled(false);
            btnPDF.disabled = true;
            output.style.display = 'block';
            tabs.style.display = 'none';
            btnBackup.disabled = false;
            ausgabeDropdown.removeAttribute('open');
            output.textContent = `✔ ${uploadResult.count} Teilnehmende geladen.\n\nNächster Schritt:\n• Klicken Sie auf "Gruppen zusammenstellen" um ausgewogene Gruppen zu erstellen`;
            
            if (uploadResult.warning) {
                alert('⚠️ Warnung Stationen:\n\n' + uploadResult.warning);
            }

            // Collapse Admin and expand Daten
            const adminDropdown = document.querySelector('.button-section:nth-child(1) .category-dropdown');
            const datenDropdown = document.querySelector('.button-section:nth-child(2) .category-dropdown');
            if (adminDropdown) adminDropdown.removeAttribute('open');
            if (datenDropdown) datenDropdown.setAttribute('open', 'open');
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
        output.textContent = 'Fehler: ' + err;
    }
}

export async function handleBackupDatabase() {
    setStatus('Datenbank-Backup wird erstellt...', 'info');
    
    try {
        const result = await window.go.main.App.BackupDatabase();
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
        } else {
            setStatus('✅ ' + result.message, 'success');
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

export async function handleRestoreDatabase() {
    setStatus('Verfügbare Backups werden geladen...', 'info');
    
    try {
        // Get list of backups
        const listResult = await window.go.main.App.ListBackups();
        
        if (listResult.status === 'error') {
            setStatus('FEHLER: ' + listResult.message, 'error');
            alert('Backups konnten nicht geladen werden: ' + listResult.message);
            return;
        }
        
        if (listResult.count === 0) {
            setStatus('Keine Backups verfügbar', 'info');
            alert('Keine Datenbankbackups gefunden. Bitte zuerst ein Backup erstellen.');
            return;
        }
        
        // Sort backups by modified date (newest first)
        const backups = listResult.backups.sort((a, b) => {
            // Parse dates and compare (newer dates should come first)
            return new Date(b.modified) - new Date(a.modified);
        });
        
window._selectedBackup = null;

        let dialogHTML = '<div style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); z-index: 1000; display: flex; justify-content: center; align-items: center;" id="restore-dialog">';
        dialogHTML += '<div style="background: white; border-radius: 12px; padding: 30px; max-width: 600px; width: 90%; max-height: 80vh; display: flex; flex-direction: column; box-shadow: 0 20px 60px rgba(0,0,0,0.3);">';
        // Header: title + Abbrechen
        dialogHTML += '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">';
        dialogHTML += '<h2 style="margin: 0; color: #333;">🔄 Datenbank wiederherstellen</h2>';
        dialogHTML += '<button onclick="window.closeRestoreDialog()" style="padding: 8px 16px; background: #ccc; color: #333; border: none; border-radius: 6px; cursor: pointer; font-weight: 600;">Abbrechen</button>';
        dialogHTML += '</div>';
        dialogHTML += '<p style="margin-bottom: 16px; color: #666;">Backup auswählen. <strong>Warnung:</strong> Dies ersetzt die aktuelle Datenbank!</p>';
        // Scrollable backup list
        dialogHTML += '<div style="overflow-y: auto; flex: 1; margin-bottom: 16px;">';
        backups.forEach((backup, index) => {
            const sizeKB = (backup.size / 1024).toFixed(2);
            // No user data in the HTML string — text is set via textContent after DOM insertion
            dialogHTML += '<div style="border: 2px solid #ddd; border-radius: 8px; padding: 15px; margin-bottom: 10px; cursor: pointer; transition: all 0.3s;" ';
            dialogHTML += 'id="backup-item-' + index + '">';
            dialogHTML += '<div class="backup-item-name" style="font-weight: 600; font-size: 14px; margin-bottom: 5px;"></div>';
            dialogHTML += '<div style="font-size: 12px; color: #666;">';
            dialogHTML += '<span class="backup-item-date"></span>';
            dialogHTML += '<span style="margin-left: 15px;">💾 ' + sizeKB + ' KB</span>';
            dialogHTML += '</div>';
            dialogHTML += '</div>';
        });
        dialogHTML += '</div>';
        // Footer: Wiederherstellen button
        dialogHTML += '<div style="text-align: right;">';
        dialogHTML += '<button id="btn-do-restore" onclick="window.doRestoreSelected()" disabled ';
        dialogHTML += 'style="padding: 10px 20px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border: none; border-radius: 6px; font-weight: 600; opacity: 0.4; cursor: not-allowed;">Wiederherstellen</button>';
        dialogHTML += '</div>';
        dialogHTML += '</div></div>';
        
        // Add dialog to page
        const dialogElement = document.createElement('div');
        dialogElement.innerHTML = dialogHTML;
        document.body.appendChild(dialogElement);

        // Populate text and wire click handlers after DOM insertion (prevents XSS via backup.name)
        backups.forEach((backup, index) => {
            const item = dialogElement.querySelector('#backup-item-' + index);
            if (item) {
                item.querySelector('.backup-item-name').textContent = backup.name;
                item.querySelector('.backup-item-date').textContent = '📅 ' + backup.modified;
                item.addEventListener('click', () => window.selectBackup(index, backup.name));
            }
        });
        
        setStatus('Backup zum Wiederherstellen auswählen', 'info');
        
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
        alert('Fehler: ' + err);
    }
}

// Helper function to close restore dialog
window.closeRestoreDialog = function() {
    const dialog = document.getElementById('restore-dialog');
    if (dialog && dialog.parentElement) {
        dialog.parentElement.remove();
    }
    setStatus('Wiederherstellung abgebrochen', 'info');
};

// Select a backup in the restore dialog
window.selectBackup = function(index, name) {
    document.querySelectorAll('[id^="backup-item-"]').forEach(el => {
        el.style.borderColor = '#ddd';
        el.style.background = 'white';
    });
    const item = document.getElementById('backup-item-' + index);
    if (item) {
        item.style.borderColor = '#667eea';
        item.style.background = '#f0f8ff';
    }
    window._selectedBackup = name;
    const btn = document.getElementById('btn-do-restore');
    if (btn) {
        btn.disabled = false;
        btn.style.opacity = '1';
        btn.style.cursor = 'pointer';
    }
};

// Trigger restore for the selected backup
window.doRestoreSelected = function() {
    if (window._selectedBackup) window.confirmRestore(window._selectedBackup);
};

// Perform restore
window.confirmRestore = async function(backupFilename) {
    window.closeRestoreDialog();
    
    setStatus('Datenbank wird aus Backup wiederhergestellt...', 'info');
    
    try {
        const result = await window.go.main.App.RestoreDatabase(backupFilename);
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            alert('Datenbank konnte nicht wiederhergestellt werden: ' + result.message);
        } else {
            setStatus('✅ ' + result.message, 'success');
            alert('Datenbank erfolgreich wiederhergestellt!\n\nDie Anwendung wird jetzt aktualisiert.');
            
            // Enable core buttons
            btnShow.disabled = false;
            btnStations.disabled = false;
            if (btnOverview) btnOverview.disabled = false;
            btnPDF.disabled = false;
            // Only enable redistribution if no scores exist yet
            const hasScores = await window.go.main.App.HasScores();
            btnDistribute.disabled = hasScores;
            // Evaluation and certificates only available once scores exist
            setEvalButtonsEnabled(hasScores);
            
            // Refresh the view
            output.style.display = 'block';
            tabs.style.display = 'none';
            btnBackup.disabled = false;
            ausgabeDropdown.setAttribute('open', 'open');
            output.textContent = '✔ Datenbank erfolgreich aus Backup wiederhergestellt!\n\nAlle Funktionen stehen jetzt zur Verfügung.';
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Error restoring database: ' + err);
    }
};

export async function handleDistributeGroups() {
    setStatus('Gruppen werden erstellt...', 'info');
    try {
        const result = await window.go.main.App.DistributeGroups();
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            return;
        }
        setStatus('✅ ' + result.message, 'success');
        if (result.warning) {
            alert('⚠️ Warnung Betreuende:\n\n' + result.warning);
        }
        btnShow.disabled = false;
        btnStations.disabled = false;
        if (btnOverview) btnOverview.disabled = false;
        btnPDF.disabled = false;
        // Evaluation and certificates stay disabled until the first score is entered
        setEvalButtonsEnabled(false);
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = `✔ ${result.message}\n\nNächste Schritte:\n• Klicken Sie auf "Gruppen anzeigen" um die Gruppen anzuzeigen\n• Klicken Sie auf "Ergebniseingabe" um Ergebnisse einzugeben\n• Klicken Sie auf "Gruppen-PDF erstellen" um die Gruppen als PDF zu exportieren\n• Auswertung und Urkunden sind verfügbar sobald das erste Ergebnis gespeichert wurde`;

        const ausgabeDropdown = document.querySelector('.button-section:nth-child(3) .category-dropdown');
        if (ausgabeDropdown) ausgabeDropdown.setAttribute('open', 'open');
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

export async function handleConvertMasterExcel() {
    // Ask which event to extract before opening the file dialog
    const event = await askEventChoice();
    if (!event) {
        setStatus('Bereit.', 'info');
        return;
    }

    setStatus(`Master-Excel wird konvertiert (${event})...`, 'info');
    try {
        const result = await window.go.main.App.ConvertMasterExcel(event);

        if (result.status === 'cancelled') {
            setStatus('Bereit.', 'info');
            return;
        }
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            output.style.display = 'block';
            output.textContent = 'Konvertierung fehlgeschlagen:\n' + result.message;
            return;
        }

        setStatus('✅ ' + result.message, 'success');
        output.style.display = 'block';
        output.textContent = `✔ Master-Excel (${event}) erfolgreich konvertiert.\n\nGespeichert unter:\n${result.destPath}\n\nNächster Schritt:\n• Klicken Sie auf "Excel einlesen" und wählen Sie die soeben gespeicherte Datei.`;
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

// Show a modal dialog asking the user to choose between Jugend and Mini.
// Returns "Jugend", "Mini", or null if cancelled.
function askEventChoice() {
    return new Promise(resolve => {
        const overlay = document.createElement('div');
        overlay.style.cssText = 'position:fixed;inset:0;background:rgba(0,0,0,.55);display:flex;align-items:center;justify-content:center;z-index:9999';

        const box = document.createElement('div');
        box.style.cssText = 'background:#fff;border-radius:10px;padding:32px 36px;max-width:480px;width:90%;box-shadow:0 8px 32px rgba(0,0,0,.3);font-family:Arial,sans-serif';

        box.innerHTML = `
            <h2 style="margin:0 0 12px;font-size:1.2em;color:#333">Veranstaltung auswählen</h2>
            <p style="margin:0 0 24px;color:#555;line-height:1.5">
                Welche Veranstaltung soll aus dem Master-Excel extrahiert werden?
            </p>
            <div style="display:flex;gap:12px;justify-content:flex-end">
                <button id="_btnMini"   style="padding:10px 20px;background:#1976d2;color:#fff;border:none;border-radius:6px;cursor:pointer;font-weight:600">Mini</button>
                <button id="_btnJugend" style="padding:10px 20px;background:#2e7d32;color:#fff;border:none;border-radius:6px;cursor:pointer;font-weight:600">Jugend</button>
            </div>
            <div style="display:flex;justify-content:flex-end;margin-top:12px;padding-top:12px;border-top:1px solid #eee">
                <button id="_btnCancel" style="padding:10px 20px;background:#9e9e9e;color:#fff;border:none;border-radius:6px;cursor:pointer;font-weight:600">Abbrechen</button>
            </div>`;

        overlay.appendChild(box);
        document.body.appendChild(overlay);

        const close = (val) => { document.body.removeChild(overlay); resolve(val); };
        box.querySelector('#_btnJugend').addEventListener('click', () => close('Jugend'));
        box.querySelector('#_btnMini').addEventListener('click',   () => close('Mini'));
        box.querySelector('#_btnCancel').addEventListener('click', () => close(null));
    });
}
