// File loading and management
import { setStatus, output, tabs, btnShow, btnDistribute, btnStations, btnEvaluation, btnOrtsverband, btnPDF, btnCertificates, btnOVCertificates } from '../shared/dom.js';

export async function openFileDialog() {
    try {
        // Check if database has data
        const result = await window.go.main.App.CheckDB();
        
        if (result.hasData) {
            const confirmed = confirm(`Die Datenbank enthält ${result.count} Teilnehmer. Möchten Sie diese Daten verwerfen und eine neue Datei laden?`);
            if (!confirmed) {
                return;
            }
        }
        
        // Open file dialog and load file
        const uploadResult = await window.go.main.App.LoadFile();
        
        if (uploadResult.status === 'error') {
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
            btnEvaluation.disabled = true;
            btnOrtsverband.disabled = true;
            btnPDF.disabled = true;
            btnCertificates.disabled = true;
            btnOVCertificates.disabled = true;
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = `✔ ${uploadResult.count} Teilnehmer geladen.\n\nNächster Schritt:\n• Klicken Sie auf "Teilnehmer zu Gruppen" um ausgewogene Gruppen zu erstellen`;
            
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
        
        let dialogHTML = '<div style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); z-index: 1000; display: flex; justify-content: center; align-items: center;" id="restore-dialog">';
        dialogHTML += '<div style="background: white; border-radius: 12px; padding: 30px; max-width: 600px; max-height: 80vh; overflow-y: auto; box-shadow: 0 20px 60px rgba(0,0,0,0.3);">';
        dialogHTML += '<h2 style="margin: 0 0 20px 0; color: #333;">🔄 Datenbank wiederherstellen</h2>';
        dialogHTML += '<p style="margin-bottom: 20px; color: #666;">Backup auswählen. <strong>Warnung:</strong> Dies ersetzt die aktuelle Datenbank!</p>';
        
        dialogHTML += '<div style="margin-bottom: 20px;">';
        backups.forEach((backup, index) => {
            const sizeKB = (backup.size / 1024).toFixed(2);
            dialogHTML += '<div style="border: 2px solid #ddd; border-radius: 8px; padding: 15px; margin-bottom: 10px; cursor: pointer; transition: all 0.3s;" ';
            dialogHTML += 'onmouseover="this.style.borderColor=\'#667eea\'; this.style.background=\'#f0f8ff\';" ';
            dialogHTML += 'onmouseout="this.style.borderColor=\'#ddd\'; this.style.background=\'white\';" ';
            dialogHTML += 'onclick="window.confirmRestore(\'' + backup.name + '\')">';
            dialogHTML += '<div style="font-weight: 600; font-size: 14px; margin-bottom: 5px;">' + backup.name + '</div>';
            dialogHTML += '<div style="font-size: 12px; color: #666;">';
            dialogHTML += '<span>📅 ' + backup.modified + '</span>';
            dialogHTML += '<span style="margin-left: 15px;">💾 ' + sizeKB + ' KB</span>';
            dialogHTML += '</div>';
            dialogHTML += '</div>';
        });
        dialogHTML += '</div>';
        
        dialogHTML += '<div style="text-align: right;">';
        dialogHTML += '<button onclick="window.closeRestoreDialog()" style="padding: 10px 20px; background: #ccc; color: #333; border: none; border-radius: 6px; cursor: pointer; font-weight: 600;">Abbrechen</button>';
        dialogHTML += '</div>';
        
        dialogHTML += '</div></div>';
        
        // Add dialog to page
        const dialogElement = document.createElement('div');
        dialogElement.innerHTML = dialogHTML;
        document.body.appendChild(dialogElement);
        
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

// Helper function to confirm and perform restore
window.confirmRestore = async function(backupFilename) {
    const confirmed = confirm(
        '⚠️ WARNUNG: Die aktuelle Datenbank wird durch das Backup ersetzt!\n\n' +
        'Backup: ' + backupFilename + '\n\n' +
        'Möchten Sie wirklich fortfahren?'
    );
    
    if (!confirmed) {
        return;
    }
    
    // Close dialog
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
            
            // Enable all buttons since we now have data
            btnShow.disabled = false;
            btnStations.disabled = false;
            btnEvaluation.disabled = false;
            btnOrtsverband.disabled = false;
            btnPDF.disabled = false;
            btnCertificates.disabled = false;
            btnOVCertificates.disabled = false;
            // Only enable redistribution if no scores exist yet
            const hasScores = await window.go.main.App.HasScores();
            btnDistribute.disabled = hasScores;
            
            // Refresh the view
            output.style.display = 'block';
            tabs.style.display = 'none';
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
        btnShow.disabled = false;
        btnStations.disabled = false;
        btnEvaluation.disabled = false;
        btnOrtsverband.disabled = false;
        btnPDF.disabled = false;
        btnCertificates.disabled = false;
        btnOVCertificates.disabled = false;
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = `✔ ${result.message}\n\nNächste Schritte:\n• Klicken Sie auf "Gruppen anzeigen" um die Gruppen anzuzeigen\n• Klicken Sie auf "Auswertung nach Gruppen" für die Gruppenauswertung\n• Klicken Sie auf "Auswertung nach Ortsverband" für die ortsverbandsbasierte Auswertung\n• Klicken Sie auf "Gruppen-PDF erstellen" um die Gruppen als PDF zu exportieren\n• Klicken Sie auf "Teilnehmer-Zertifikate" um Zertifikate zu erstellen`;

        const ausgabeDropdown = document.querySelector('.button-section:nth-child(3) .category-dropdown');
        if (ausgabeDropdown) ausgabeDropdown.setAttribute('open', 'open');
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}
