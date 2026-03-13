// Station management and display - Group-based results entry
import { setStatus, output, tabs, tabButtons, tabContents, clearAllTabs } from '../shared/dom.js';
import { escapeHtml } from '../shared/utils.js';

export async function handleShowStations() {
    await handleShowStationsForGroup(null);
}

export async function handleShowStationsForGroup(preselectedGroupID) {
    setStatus('Loading data for results entry...', 'info');
    
    try {
        // Get both stations and groups
        const [stationsResult, groupsResult] = await Promise.all([
            window.go.main.App.ShowStations(),
            window.go.main.App.GetAllGroups()
        ]);
        
        if (stationsResult.status === 'error') {
            setStatus('ERROR: ' + stationsResult.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Error loading stations: ' + stationsResult.message;
        } else if (groupsResult.status === 'error') {
            setStatus('ERROR: ' + groupsResult.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Error loading groups: ' + groupsResult.message;
        } else {
            setStatus('Ready for results entry', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            // Ensure complete cleanup before rendering
            clearAllTabs();
            renderGroupBasedEntry(stationsResult.stations, groupsResult.groups, preselectedGroupID);
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = 'Error: ' + err;
    }
}

function renderGroupBasedEntry(stations, groups, preselectedGroupID = null) {
    // Clear existing content
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!stations || stations.length === 0) {
        tabContents.innerHTML = '<div class="empty-message">No stations found.</div>';
        return;
    }
    
    if (!groups || groups.length === 0) {
        tabContents.innerHTML = '<div class="empty-message">No groups found.</div>';
        return;
    }
    
    // Store data globally
    window.currentStations = stations;
    window.currentGroups = groups;
    
    // Create main container
    const container = document.createElement('div');
    container.className = 'results-entry-container';
    
    // Group selector section
    let html = '<div class="group-selector-panel">';
    html += '<h2>📝 Ergebniseingabe</h2>';
    html += '<div class="group-selector-row">';
    html += '<label class="group-selector-label">Gruppe auswählen:</label>';
    html += '<select id="group-selector" class="group-selector">';
    html += '<option value="">Bitte wählen...</option>';
    groups.forEach(groupID => {
        const selected = groupID === preselectedGroupID ? ' selected' : '';
        html += '<option value="' + groupID + '"' + selected + '>Gruppe ' + groupID + '</option>';
    });
    html += '</select>';
    html += '</div>';
    html += '</div>';
    
    // Results table container (initially hidden)
    html += '<div id="results-table-container"></div>';
    
    container.innerHTML = html;
    tabContents.appendChild(container);
    
    // Add event listener to group selector
    setTimeout(() => {
        const groupSelector = document.getElementById('group-selector');
        if (groupSelector) {
            groupSelector.addEventListener('change', (e) => {
                const groupID = parseInt(e.target.value);
                const resultsContainer = document.getElementById('results-table-container');
                if (groupID && !isNaN(groupID)) {
                    renderStationTable(groupID, stations);
                } else {
                    resultsContainer.classList.remove('visible');
                }
            });
            // If a group was preselected, show its table immediately
            if (preselectedGroupID) {
                renderStationTable(preselectedGroupID, stations);
            }
        }
    }, 0);
}

function renderStationTable(groupID, stations) {
    const container = document.getElementById('results-table-container');
    if (!container) return;
    
    let html = '<h3 class="results-table-header">🏆 Ergebnisse für Gruppe ' + groupID + '</h3>';
    
    // Create table
    html += '<table class="group-table results-table">';
    html += '<thead><tr>';
    html += '<th class="col-station">Station</th>';
    html += '<th class="col-score">Ergebnis (100-1200)</th>';
    html += '<th class="col-action">Aktion</th>';
    html += '</tr></thead><tbody>';
    
    stations.forEach((station, index) => {
        // Find existing score for this group at this station
        let existingScore = '';
        if (station.GroupScores && station.GroupScores.length > 0) {
            const scoreEntry = station.GroupScores.find(gs => gs.GroupID === groupID);
            if (scoreEntry) {
                existingScore = scoreEntry.Score;
            }
        }
        
        html += '<tr id="row-' + station.StationID + '">';
        html += '<td class="station-name">' + escapeHtml(station.StationName) + '</td>';
        html += '<td>';
        html += '<input type="number" id="score-' + station.StationID + '" ';
        html += 'class="score-input" ';
        html += 'min="100" max="1200" step="50" ';
        html += 'value="' + existingScore + '" ';
        html += 'placeholder="100-1200">';
        html += '</td>';
        html += '<td>';
        html += '<button onclick="window.saveStationScore(' + groupID + ', ' + station.StationID + ')" ';
        html += 'class="btn-save-score">Speichern</button>';
        html += '</td>';
        html += '</tr>';
    });
    
    html += '</tbody></table>';
    
    // Save all button
    html += '<div class="save-all-container">';
    html += '<button onclick="window.saveAllScores(' + groupID + ')" ';
    html += 'class="btn-save-all">💾 Alle Ergebnisse speichern</button>';
    html += '</div>';
    
    container.innerHTML = html;
    container.classList.add('visible');
}

// Save single station score
window.saveStationScore = async function(groupID, stationID) {
    const scoreInput = document.getElementById('score-' + stationID);
    if (!scoreInput) return;
    
    const score = parseInt(scoreInput.value);
    
    if (!score || isNaN(score)) {
        alert('Bitte geben Sie ein gültiges Ergebnis ein.');
        return;
    }
    
    if (score < 100 || score > 1200) {
        alert('Das Ergebnis muss zwischen 100 und 1200 liegen.');
        return;
    }
    
    try {
        setStatus('Speichere Ergebnis...', 'info');
        const result = await window.go.main.App.AssignScore(groupID, stationID, score);
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            alert('Fehler beim Speichern: ' + result.message);
        } else {
            setStatus('✔ Ergebnis gespeichert', 'success');
            // Highlight the row briefly
            const row = document.getElementById('row-' + stationID);
            if (row) {
                row.classList.add('row-saved');
                setTimeout(() => { row.classList.remove('row-saved'); }, 1000);
            }
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Fehler: ' + err);
    }
};

// Save all scores for the selected group
window.saveAllScores = async function(groupID) {
    const stations = window.currentStations;
    if (!stations) return;
    
    const scoresToSave = [];
    let hasErrors = false;
    
    // Collect all scores
    stations.forEach(station => {
        const scoreInput = document.getElementById('score-' + station.StationID);
        if (scoreInput && scoreInput.value) {
            const score = parseInt(scoreInput.value);
            
            if (isNaN(score) || score < 100 || score > 1200) {
                alert('Ungültiges Ergebnis bei Station ' + station.StationName + '. Muss zwischen 100 und 1200 liegen.');
                hasErrors = true;
                return;
            }
            
            scoresToSave.push({
                stationID: station.StationID,
                stationName: station.StationName,
                score: score
            });
        }
    });
    
    if (hasErrors) return;
    
    if (scoresToSave.length === 0) {
        alert('Keine Ergebnisse zum Speichern eingegeben.');
        return;
    }
    
    // Confirm before saving all
    const confirmed = confirm(
        'Möchten Sie ' + scoresToSave.length + ' Ergebnis(se) für Gruppe ' + groupID + ' speichern?'
    );
    
    if (!confirmed) return;
    
    try {
        setStatus('Speichere alle Ergebnisse...', 'info');
        let savedCount = 0;
        let errorCount = 0;
        
        // Save each score
        for (const scoreData of scoresToSave) {
            const result = await window.go.main.App.AssignScore(groupID, scoreData.stationID, scoreData.score);
            
            if (result.status === 'error') {
                errorCount++;
                console.error('Error saving score for station ' + scoreData.stationName + ': ' + result.message);
            } else {
                savedCount++;
                // Highlight the row
                const row = document.getElementById('row-' + scoreData.stationID);
                if (row) {
                    row.classList.add('row-saved');
                }
            }
        }
        
        if (errorCount > 0) {
            setStatus('⚠ ' + savedCount + ' gespeichert, ' + errorCount + ' Fehler', 'error');
            alert('Es gab Fehler beim Speichern einiger Ergebnisse.\nGespeichert: ' + savedCount + '\nFehler: ' + errorCount);
        } else {
            setStatus('✔ Alle ' + savedCount + ' Ergebnisse gespeichert', 'success');
            alert('Alle Ergebnisse erfolgreich gespeichert!');
        }
        
        // Remove highlights after a delay
        setTimeout(() => {
            scoresToSave.forEach(scoreData => {
                const row = document.getElementById('row-' + scoreData.stationID);
                if (row) row.classList.remove('row-saved');
            });
        }, 2000);
        
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Fehler: ' + err);
    }
};
