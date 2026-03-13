// Score assignment functionality
import { setStatus } from '../shared/dom.js';

export function checkForExistingScore() {
    const groupSelect = document.getElementById('global-group-select');
    const stationInput = document.getElementById('global-station-select');
    const warningDiv = document.getElementById('score-warning');
    const warningText = document.getElementById('score-warning-text');
    
    if (!groupSelect || !stationInput || !warningDiv || !warningText) return;
    
    const groupID = parseInt(groupSelect.value);
    const stationID = parseInt(stationInput.getAttribute('data-station-id'));
    
    // Hide warning if no group selected
    if (!groupID || isNaN(groupID)) {
        warningDiv.style.display = 'none';
        return;
    }
    
    // Check if the group already has a score for this station
    if (window.currentStations) {
        const currentStation = window.currentStations.find(s => s.StationID === stationID);
        if (currentStation && currentStation.GroupScores) {
            const existingScore = currentStation.GroupScores.find(gs => gs.GroupID === groupID);
            if (existingScore) {
                warningText.textContent = `Group ${groupID} already has a score of ${existingScore.Score} for ${currentStation.StationName}. Saving will overwrite this score.`;
                warningDiv.style.display = 'block';
                return;
            }
        }
    }
    
    // No existing score found
    warningDiv.style.display = 'none';
}

export async function handleGlobalAssignScore() {
    const groupSelect = document.getElementById('global-group-select');
    const scoreSelect = document.getElementById('global-score-select');
    const stationInput = document.getElementById('global-station-select');
    
    const groupID = parseInt(groupSelect.value);
    const score = parseInt(scoreSelect.value);
    const stationID = parseInt(stationInput.getAttribute('data-station-id'));
    
    if (!groupID || isNaN(groupID)) {
        alert('Please select a group');
        return;
    }
    
    if (!score || isNaN(score)) {
        alert('Please select a score');
        return;
    }
    
    if (!stationID || isNaN(stationID)) {
        alert('Please select a station');
        return;
    }
    
    // Check if the group already has a score for this station
    if (window.currentStations) {
        const currentStation = window.currentStations.find(s => s.StationID === stationID);
        if (currentStation && currentStation.GroupScores) {
            const existingScore = currentStation.GroupScores.find(gs => gs.GroupID === groupID);
            if (existingScore) {
                const confirmed = confirm(
                    `⚠️ WARNING: Group ${groupID} already has a score for ${currentStation.StationName}!\n\n` +
                    `Current Score: ${existingScore.Score}\n` +
                    `New Score: ${score}\n\n` +
                    `Do you want to OVERWRITE the existing score?`
                );
                if (!confirmed) {
                    setStatus('Score assignment cancelled', 'info');
                    return;
                }
            }
        }
    }
    
    try {
        setStatus('Saving score...', 'info');
        const result = await window.go.main.App.AssignScore(groupID, stationID, score);
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            alert('Failed to save score: ' + result.message);
        } else {
            setStatus('✔ ' + result.message, 'success');
            // Clear the form
            groupSelect.value = '';
            scoreSelect.value = '650';
            document.getElementById('score-display').textContent = '650';
            
            // Clear the warning
            const warningDiv = document.getElementById('score-warning');
            if (warningDiv) {
                warningDiv.style.display = 'none';
            }
            
            // Find current station index to restore selection after refresh
            const currentStationID = stationID;
            let currentStationIndex = window.currentStationIndex || 0;
            
            // Refresh the stations view and restore selection
            const [stationsResult, groupsResult] = await Promise.all([
                window.go.main.App.ShowStations(),
                window.go.main.App.GetAllGroups()
            ]);
            
            if (stationsResult.status === 'success' && groupsResult.status === 'success') {
                window.currentStations = stationsResult.stations;
                window.currentGroups = groupsResult.groups;
                // Find the station index by ID
                const stationIndex = stationsResult.stations.findIndex(s => s.StationID === currentStationID);
                // Import showStationDetails from stations.js
                const { showStationDetails } = await import('./stations.js');
                showStationDetails(stationIndex >= 0 ? stationIndex : currentStationIndex);
            }
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Error saving score: ' + err);
    }
}

export async function handleAssignScore(stationID) {
    const groupSelect = document.getElementById('group-select-' + stationID);
    const scoreInput = document.getElementById('score-input-' + stationID);
    
    const groupID = parseInt(groupSelect.value);
    const score = parseInt(scoreInput.value);
    
    if (!groupID || isNaN(groupID)) {
        alert('Please select a group');
        return;
    }
    
    if (isNaN(score) || score < 0) {
        alert('Please enter a valid score (0 or higher)');
        return;
    }
    
    try {
        setStatus('Saving score...', 'info');
        const result = await window.go.main.App.AssignScore(groupID, stationID, score);
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            alert('Failed to save score: ' + result.message);
        } else {
            setStatus('✔ ' + result.message, 'success');
            // Clear the form
            groupSelect.value = '';
            scoreInput.value = '';
            
            // Find current station index to restore selection after refresh
            const stations = window.currentStations;
            let currentStationIndex = 0;
            if (stations) {
                currentStationIndex = stations.findIndex(s => s.StationID === stationID);
                if (currentStationIndex === -1) currentStationIndex = 0;
            }
            
            // Refresh the stations view and restore selection
            const [stationsResult, groupsResult] = await Promise.all([
                window.go.main.App.ShowStations(),
                window.go.main.App.GetAllGroups()
            ]);
            
            if (stationsResult.status === 'success' && groupsResult.status === 'success') {
                window.currentStations = stationsResult.stations;
                window.currentGroups = groupsResult.groups;
                // Import showStationDetails from stations.js
                const { showStationDetails } = await import('./stations.js');
                showStationDetails(currentStationIndex);
            }
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Error saving score: ' + err);
    }
}
