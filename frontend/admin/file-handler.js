// File loading and management
import { setStatus, output, tabs, btnShow, btnStations, btnEvaluation, btnOrtsverband, btnPDF, btnCertificates } from '../shared/dom.js';

export async function openFileDialog() {
    try {
        // Check if database has data
        const result = await window.go.main.App.CheckDB();
        
        if (result.hasData) {
            const confirmed = confirm(`The database contains ${result.count} participants. Do you want to discard this data and load a new file?`);
            if (!confirmed) {
                return;
            }
        }
        
        // Open file dialog and load file
        const uploadResult = await window.go.main.App.LoadFile();
        
        if (uploadResult.status === 'error') {
            setStatus('ERROR: ' + uploadResult.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Failed to load file. Please check the error message above.';
        } else {
            setStatus(uploadResult.message, 'success');
            btnShow.disabled = false;
            btnStations.disabled = false;
            btnEvaluation.disabled = false;
            btnOrtsverband.disabled = false;
            btnPDF.disabled = false;
            btnCertificates.disabled = false;
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = `✔ Successfully loaded ${uploadResult.count} participants and created balanced groups!\n\nNext steps:\n• Click "Show Groups" to view the groups\n• Click "Auswertung nach Gruppen" for group evaluation\n• Click "Auswertung nach Ortsverband" for location-based evaluation\n• Click "Generate PDF" to export groups to PDF\n• Click "Teilnehmer-Zertifikate" to generate participant certificates`;
            
            // Collapse Admin and expand other categories
            const adminDropdown = document.querySelector('.button-section:nth-child(1) .category-dropdown');
            const datenDropdown = document.querySelector('.button-section:nth-child(2) .category-dropdown');
            const ausgabeDropdown = document.querySelector('.button-section:nth-child(3) .category-dropdown');
            if (adminDropdown) adminDropdown.removeAttribute('open');
            if (datenDropdown) datenDropdown.setAttribute('open', 'open');
            if (ausgabeDropdown) ausgabeDropdown.setAttribute('open', 'open');
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        output.textContent = 'Error: ' + err;
    }
}

export async function handleBackupDatabase() {
    setStatus('Creating database backup...', 'info');
    
    try {
        const result = await window.go.main.App.BackupDatabase();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
        } else {
            setStatus('✅ ' + result.message, 'success');
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
    }
}

export async function handleRestoreDatabase() {
    setStatus('Loading available backups...', 'info');
    
    try {
        // Get list of backups
        const listResult = await window.go.main.App.ListBackups();
        
        if (listResult.status === 'error') {
            setStatus('ERROR: ' + listResult.message, 'error');
            alert('Failed to load backups: ' + listResult.message);
            return;
        }
        
        if (listResult.count === 0) {
            setStatus('No backups available', 'info');
            alert('No database backups found. Please create a backup first.');
            return;
        }
        
        // Sort backups by modified date (newest first)
        const backups = listResult.backups.sort((a, b) => {
            // Parse dates and compare (newer dates should come first)
            return new Date(b.modified) - new Date(a.modified);
        });
        
        let dialogHTML = '<div style="position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); z-index: 1000; display: flex; justify-content: center; align-items: center;" id="restore-dialog">';
        dialogHTML += '<div style="background: white; border-radius: 12px; padding: 30px; max-width: 600px; max-height: 80vh; overflow-y: auto; box-shadow: 0 20px 60px rgba(0,0,0,0.3);">';
        dialogHTML += '<h2 style="margin: 0 0 20px 0; color: #333;">🔄 Restore Database</h2>';
        dialogHTML += '<p style="margin-bottom: 20px; color: #666;">Select a backup to restore. <strong>Warning:</strong> This will replace your current database!</p>';
        
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
        dialogHTML += '<button onclick="window.closeRestoreDialog()" style="padding: 10px 20px; background: #ccc; color: #333; border: none; border-radius: 6px; cursor: pointer; font-weight: 600;">Cancel</button>';
        dialogHTML += '</div>';
        
        dialogHTML += '</div></div>';
        
        // Add dialog to page
        const dialogElement = document.createElement('div');
        dialogElement.innerHTML = dialogHTML;
        document.body.appendChild(dialogElement);
        
        setStatus('Select a backup to restore', 'info');
        
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Error: ' + err);
    }
}

// Helper function to close restore dialog
window.closeRestoreDialog = function() {
    const dialog = document.getElementById('restore-dialog');
    if (dialog && dialog.parentElement) {
        dialog.parentElement.remove();
    }
    setStatus('Restore cancelled', 'info');
};

// Helper function to confirm and perform restore
window.confirmRestore = async function(backupFilename) {
    const confirmed = confirm(
        '⚠️ WARNING: This will replace your current database with the backup!\n\n' +
        'Backup: ' + backupFilename + '\n\n' +
        'Are you sure you want to continue?'
    );
    
    if (!confirmed) {
        return;
    }
    
    // Close dialog
    window.closeRestoreDialog();
    
    setStatus('Restoring database from backup...', 'info');
    
    try {
        const result = await window.go.main.App.RestoreDatabase(backupFilename);
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            alert('Failed to restore database: ' + result.message);
        } else {
            setStatus('✅ ' + result.message, 'success');
            alert('Database restored successfully!\n\nThe application will now refresh.');
            
            // Enable all buttons since we now have data
            btnShow.disabled = false;
            btnStations.disabled = false;
            btnEvaluation.disabled = false;
            btnOrtsverband.disabled = false;
            btnPDF.disabled = false;
            btnCertificates.disabled = false;
            
            // Refresh the view
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = '✔ Database restored successfully from backup!\n\nYou can now use all features.';
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Error restoring database: ' + err);
    }
};
