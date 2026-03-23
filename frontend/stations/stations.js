// Station management and display - Group-based results entry
import { setStatus, output, tabs, tabButtons, tabContents, clearAllTabs, btnDistribute, setEvalButtonsEnabled } from '../shared/dom.js';
import { escapeHtml } from '../shared/utils.js';

// Score bounds are read from the backend config (window.appConfig) at runtime.
// These fallbacks are used if the config has not loaded yet.
function scoreMin() { return (window.appConfig && window.appConfig.scoreMin) || 100; }
function scoreMax() { return (window.appConfig && window.appConfig.scoreMax) || 1200; }

// Dirty-tracking state – reset each time a new group table is rendered
let savedScoreMap = {};   // { stationID: number|'' } last persisted value per station
let currentGroupID = null;
let pendingGroupID = null;

// Centralised score validation. Returns an error string or null if valid.
function validateScore(value, stationName) {
    const score = parseInt(value, 10);
    if (isNaN(score)) return 'Bitte geben Sie ein gültiges Ergebnis ein' + (stationName ? ' für Station ' + stationName : '') + '.';
    if (score < scoreMin() || score > scoreMax()) return 'Das Ergebnis bei ' + (stationName || 'dieser Station') + ' muss zwischen ' + scoreMin() + ' und ' + scoreMax() + ' liegen.';
    return null;
}

// Keep window.currentStations in sync after any successful DB save so that
// switching between groups always pre-fills with the latest persisted values.
function updateStationCache(stationID, groupID, score) {
    if (!window.currentStations) return;
    const station = window.currentStations.find(s => s.StationID === stationID);
    if (!station) return;
    if (!station.GroupScores) station.GroupScores = [];
    const existing = station.GroupScores.find(gs => gs.GroupID === groupID);
    if (existing) {
        existing.Score = score;
    } else {
        station.GroupScores.push({ GroupID: groupID, Score: score });
    }
}

export async function handleShowStations() {
    await handleShowStationsForGroup(null);
}

export async function handleShowStationsForGroup(preselectedGroupID) {
    setStatus('Daten für Ergebniseingabe werden geladen...', 'info');
    
    try {
        // Get both stations and groups
        const [stationsResult, groupsResult] = await Promise.all([
            window.go.main.App.ShowStations(),
            window.go.main.App.GetAllGroups()
        ]);
        
        if (stationsResult.status === 'error') {
            setStatus('FEHLER: ' + stationsResult.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Fehler beim Laden der Stationen: ' + stationsResult.message;
        } else if (groupsResult.status === 'error') {
            setStatus('FEHLER: ' + groupsResult.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Fehler beim Laden der Gruppen: ' + groupsResult.message;
        } else {
            setStatus('Bereit zur Ergebniseingabe', 'success');
            document.querySelectorAll('.category-dropdown').forEach(d => d.removeAttribute('open'));
            output.style.display = 'none';
            tabs.style.display = 'block';
            // Ensure complete cleanup before rendering
            clearAllTabs();
            renderGroupBasedEntry(stationsResult.stations, groupsResult.groups, preselectedGroupID);
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = 'Fehler: ' + err;
    }
}

function renderGroupBasedEntry(stations, groups, preselectedGroupID = null) {
    // Clear existing content
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!stations || stations.length === 0) {
        tabContents.innerHTML = '<div class="empty-message">Keine Stationen gefunden.</div>';
        return;
    }
    
    if (!groups || groups.length === 0) {
        tabContents.innerHTML = '<div class="empty-message">Keine Gruppen gefunden.</div>';
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
    
    // Attach event listener synchronously — the DOM node exists after appendChild
    const groupSelector = document.getElementById('group-selector');
    if (groupSelector) {
        groupSelector.addEventListener('change', (e) => {
            const newGroupID = parseInt(e.target.value, 10);
            if (!newGroupID || isNaN(newGroupID)) {
                document.getElementById('results-table-container').classList.remove('visible');
                return;
            }
            if (hasDirtyScores()) {
                groupSelector.value = currentGroupID || '';
                pendingGroupID = newGroupID;
                showUnsavedWarning(groupSelector, stations);
            } else {
                renderStationTable(newGroupID, stations);
            }
        });
        // If a group was preselected, show its table immediately
        if (preselectedGroupID) {
            renderStationTable(preselectedGroupID, stations);
        }
    }
}

function renderStationTable(groupID, stations) {
    currentGroupID = groupID;
    savedScoreMap = {};
    setStatus('Gruppe ' + groupID + ' – Ergebnisse eingeben', 'info');
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
        savedScoreMap[station.StationID] = existingScore;
        
        html += '<tr id="row-' + station.StationID + '">';
        html += '<td class="station-name">' + escapeHtml(station.StationName) + '</td>';
        html += '<td>';
        html += '<input type="text" inputmode="numeric" id="score-' + station.StationID + '" ';
        html += 'class="score-input" ';
        html += 'value="' + existingScore + '" ';
        html += 'placeholder="' + scoreMin() + '-' + scoreMax() + '" ';
        html += 'oninput="window.onScoreInput(this)">';
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

    // Apply saved (light blue) highlight to all pre-filled inputs
    stations.forEach((station) => {
        if (savedScoreMap[station.StationID] !== '') {
            const inp = document.getElementById('score-' + station.StationID);
            if (inp) inp.classList.add('score-input--saved');
        }
    });
}

// Score input: digits only + live range validation highlight
window.onScoreInput = function(input) {
    const stripped = input.value.replace(/\D/g, '');
    if (input.value !== stripped) input.value = stripped;
    input.classList.remove('score-input--invalid', 'score-input--valid', 'score-input--saved');
    if (stripped === '') return;
    const score = parseInt(stripped, 10);
    if (score < scoreMin() || score > scoreMax()) {
        input.classList.add('score-input--invalid');
    } else {
        input.classList.add('score-input--valid');
    }
};

// Save single station score
window.saveStationScore = async function(groupID, stationID) {
    const scoreInput = document.getElementById('score-' + stationID);
    if (!scoreInput) return;

    const error = validateScore(scoreInput.value, null);
    if (error) { alert(error); return; }

    const score = parseInt(scoreInput.value, 10);
    try {
        setStatus('Speichere Ergebnis...', 'info');
        const result = await window.go.main.App.AssignScore(groupID, stationID, score);
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            alert('Fehler beim Speichern: ' + result.message);
        } else {
            setStatus('✔ Ergebnis gespeichert', 'success');
            savedScoreMap[stationID] = score;
            updateStationCache(stationID, groupID, score);
            if (btnDistribute) btnDistribute.disabled = true;
            setEvalButtonsEnabled(true);
            // Switch input to saved (light blue) state
            const scoreInput2 = document.getElementById('score-' + stationID);
            if (scoreInput2) {
                scoreInput2.classList.remove('score-input--valid', 'score-input--invalid');
                scoreInput2.classList.add('score-input--saved');
            }
            const row = document.getElementById('row-' + stationID);
            if (row) {
                row.classList.add('row-saved');
                setTimeout(() => { row.classList.remove('row-saved'); }, 1000);
            }
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
        alert('Fehler: ' + err);
    }
};

// Core save logic: collects, validates and persists all filled-in scores for groupID.
// Updates savedScoreMap and row highlights. Returns { saved, errors }.
async function doSaveAll(groupID) {
    const stations = window.currentStations;
    if (!stations) return { saved: 0, errors: 0 };

    const scoresToSave = [];
    for (const station of stations) {
        const scoreInput = document.getElementById('score-' + station.StationID);
        if (!scoreInput || scoreInput.value.trim() === '') continue;
        const error = validateScore(scoreInput.value, station.StationName);
        if (error) {
            alert(error);
            return { saved: 0, errors: 1 };
        }
        scoresToSave.push({ stationID: station.StationID, stationName: station.StationName, score: parseInt(scoreInput.value, 10) });
    }

    if (scoresToSave.length === 0) return { saved: 0, errors: 0 };

    let savedCount = 0, errorCount = 0;
    for (const scoreData of scoresToSave) {
        try {
            const result = await window.go.main.App.AssignScore(groupID, scoreData.stationID, scoreData.score);
            if (result.status === 'error') {
                errorCount++;
                console.error('Error saving score for station ' + scoreData.stationName + ': ' + result.message);
            } else {
                savedCount++;
                savedScoreMap[scoreData.stationID] = scoreData.score;
                updateStationCache(scoreData.stationID, groupID, scoreData.score);
                if (btnDistribute) btnDistribute.disabled = true;
                setEvalButtonsEnabled(true);
                // Switch input to saved (light blue) state
                const savedInput = document.getElementById('score-' + scoreData.stationID);
                if (savedInput) {
                    savedInput.classList.remove('score-input--valid', 'score-input--invalid');
                    savedInput.classList.add('score-input--saved');
                }
                const row = document.getElementById('row-' + scoreData.stationID);
                if (row) {
                    row.classList.add('row-saved');
                    setTimeout(() => { row.classList.remove('row-saved'); }, 2000);
                }
            }
        } catch (err) {
            errorCount++;
            console.error('Exception saving score: ' + err);
        }
    }
    return { saved: savedCount, errors: errorCount };
}

// Save all scores for the selected group ("Alle Ergebnisse speichern" button).
window.saveAllScores = async function(groupID) {
    const stations = window.currentStations;
    if (!stations) return;

    let count = 0;
    for (const station of stations) {
        const input = document.getElementById('score-' + station.StationID);
        if (input && input.value.trim() !== '') count++;
    }
    if (count === 0) { alert('Keine Ergebnisse zum Speichern eingegeben.'); return; }

    const confirmed = confirm('Möchten Sie ' + count + ' Ergebnis(se) für Gruppe ' + groupID + ' speichern?');
    if (!confirmed) return;

    setStatus('Speichere alle Ergebnisse...', 'info');
    try {
        const { saved, errors } = await doSaveAll(groupID);
        if (errors > 0) {
            setStatus('⚠ ' + saved + ' gespeichert, ' + errors + ' Fehler', 'error');
            alert('Es gab Fehler beim Speichern.\nGespeichert: ' + saved + '\nFehler: ' + errors);
        } else {
            setStatus('✔ Alle ' + saved + ' Ergebnisse gespeichert', 'success');
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
        alert('Fehler: ' + err);
    }
};

// Returns true if any score input differs from its last saved (persisted) value.
function hasDirtyScores() {
    const stations = window.currentStations;
    if (!stations || !currentGroupID) return false;
    for (const station of stations) {
        const input = document.getElementById('score-' + station.StationID);
        if (!input) continue;
        const rawVal = input.value.trim();
        const inputVal = rawVal === '' ? '' : parseInt(rawVal, 10);
        const savedVal = savedScoreMap[station.StationID] !== undefined
            ? savedScoreMap[station.StationID]
            : '';
        if (String(inputVal) !== String(savedVal)) return true;
    }
    return false;
}

// Shows a modal warning when unsaved scores exist and the user tries to switch groups.
function showUnsavedWarning(groupSelector, stations) {
    // Prevent stacking duplicate modals
    if (document.querySelector('.unsaved-modal-overlay')) return;

    const overlay = document.createElement('div');
    overlay.className = 'unsaved-modal-overlay';
    overlay.innerHTML =
        '<div class="unsaved-modal">' +
            '<h3>&#9888;&#65039; Ungespeicherte Ergebnisse</h3>' +
            '<p>Für <strong>Gruppe ' + currentGroupID + '</strong> gibt es Ergebnisse, die noch nicht ' +
            'gespeichert wurden. Möchten Sie trotzdem wechseln oder zuerst alle Ergebnisse speichern?</p>' +
            '<div class="unsaved-modal-buttons">' +
                '<button class="btn-modal-discard" id="modal-discard">Ohne Speichern wechseln</button>' +
                '<button class="btn-modal-save" id="modal-save">&#128190; Alle speichern &amp; wechseln</button>' +
            '</div>' +
        '</div>';
    document.body.appendChild(overlay);

    document.getElementById('modal-discard').addEventListener('click', () => {
        overlay.remove();
        const target = pendingGroupID;
        pendingGroupID = null;
        groupSelector.value = target;
        renderStationTable(target, stations);
    });

    document.getElementById('modal-save').addEventListener('click', async () => {
        overlay.remove();
        const target = pendingGroupID;
        pendingGroupID = null;
        setStatus('Speichere alle Ergebnisse...', 'info');
        try {
            const { saved, errors } = await doSaveAll(currentGroupID);
            if (errors > 0) {
                setStatus('⚠ ' + saved + ' gespeichert, ' + errors + ' Fehler', 'error');
            } else {
                setStatus('✔ Alle ' + saved + ' Ergebnisse gespeichert', 'success');
            }
        } catch (err) {
            setStatus('FEHLER: ' + err, 'error');
        }
        groupSelector.value = target;
        renderStationTable(target, stations);
    });
}
