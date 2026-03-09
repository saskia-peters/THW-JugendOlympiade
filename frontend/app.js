const status = document.getElementById('status');
const output = document.getElementById('output');
const tabs = document.getElementById('tabs');
const tabButtons = document.getElementById('tabButtons');
const tabContents = document.getElementById('tabContents');
const btnShow = document.getElementById('btnShow');
const btnStations = document.getElementById('btnStations');
const btnEvaluation = document.getElementById('btnEvaluation');
const btnOrtsverband = document.getElementById('btnOrtsverband');
const btnPDF = document.getElementById('btnPDF');

function setStatus(msg, type = 'info') {
    status.textContent = msg;
    status.className = 'status ' + type;
}

function clearAllTabs() {
    // Completely clear tab containers
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    // Force a reflow to ensure cleanup
    void tabButtons.offsetHeight;
    void tabContents.offsetHeight;
}

async function openFileDialog() {
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
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = `✔ Successfully loaded ${uploadResult.count} participants and created balanced groups!\n\nNext steps:\n• Click "Show Groups" to view the groups\n• Click "Auswertung nach Gruppen" for group evaluation\n• Click "Auswertung nach Ortsverband" for location-based evaluation\n• Click "Generate PDF" to export groups to PDF`;
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        output.textContent = 'Error: ' + err;
    }
}

async function handleShowGroups() {
    setStatus('Loading groups...', 'info');
    
    try {
        const result = await window.go.main.App.ShowGroups();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Error loading groups: ' + result.message;
        } else {
            setStatus('Displaying ' + result.count + ' balanced groups', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            // Ensure complete cleanup before rendering
            clearAllTabs();
            renderGroupTabs(result.groups);
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = 'Error: ' + err;
    }
}

function renderGroupTabs(groups) {
    // Clear existing tabs - already done by clearAllTabs, but keep for safety
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!groups || groups.length === 0) {
        tabContents.innerHTML = '<div style="padding: 20px;">No groups found.</div>';
        return;
    }
    
    // Create tabs for each group
    groups.forEach((group, index) => {
        // Create tab button
        const button = document.createElement('button');
        button.className = 'tab-button' + (index === 0 ? ' active' : '');
        button.textContent = 'Group ' + group.GroupID;
        button.onclick = () => switchTab(index);
        tabButtons.appendChild(button);
        
        // Create tab content
        const content = document.createElement('div');
        content.className = 'tab-content' + (index === 0 ? ' active' : '');
        content.innerHTML = formatGroupContent(group);
        tabContents.appendChild(content);
    });
}

function switchTab(tabIndex) {
    const buttons = tabButtons.querySelectorAll('.tab-button');
    const contents = tabContents.querySelectorAll('.tab-content');
    
    buttons.forEach((btn, i) => {
        btn.classList.toggle('active', i === tabIndex);
    });
    
    contents.forEach((content, i) => {
        content.classList.toggle('active', i === tabIndex);
    });
}

function formatGroupContent(group) {
    let html = '<h2 style="margin-bottom: 15px; color: #333;">Group ' + group.GroupID + '</h2>';
    
    // Participants table
    html += '<table class="group-table">';
    html += '<thead><tr>';
    html += '<th>Name</th>';
    html += '<th>Ortsverband</th>';
    html += '<th>Alter</th>';
    html += '<th>Geschlecht</th>';
    html += '</tr></thead><tbody>';
    
    if (group.Teilnehmers && group.Teilnehmers.length > 0) {
        group.Teilnehmers.forEach(t => {
            html += '<tr>';
            html += '<td>' + escapeHtml(t.Name) + '</td>';
            html += '<td>' + escapeHtml(t.Ortsverband) + '</td>';
            html += '<td>' + t.Alter + '</td>';
            html += '<td>' + escapeHtml(t.Geschlecht) + '</td>';
            html += '</tr>';
        });
    } else {
        html += '<tr><td colspan="4">No participants</td></tr>';
    }
    
    html += '</tbody></table>';
    
    // Statistics panel
    html += '<div class="stats-panel">';
    html += '<h3>📊 Group Statistics</h3>';
    html += '<div class="stats-grid">';
    
    // Total participants
    html += '<div class="stat-item">';
    html += '<strong>Total Participants</strong>';
    html += '<span>' + (group.Teilnehmers ? group.Teilnehmers.length : 0) + '</span>';
    html += '</div>';
    
    // Average age
    if (group.Teilnehmers && group.Teilnehmers.length > 0) {
        const avgAge = (group.AlterSum / group.Teilnehmers.length).toFixed(1);
        html += '<div class="stat-item">';
        html += '<strong>Average Age</strong>';
        html += '<span>' + avgAge + ' years</span>';
        html += '</div>';
    }
    
    // Ortsverband distribution
    if (group.Ortsverbands && Object.keys(group.Ortsverbands).length > 0) {
        html += '<div class="stat-item">';
        html += '<strong>Ortsverbände</strong>';
        for (const [ort, count] of Object.entries(group.Ortsverbands)) {
            html += '<div>' + escapeHtml(ort) + ': ' + count + '</div>';
        }
        html += '</div>';
    }
    
    // Gender distribution
    if (group.Geschlechts && Object.keys(group.Geschlechts).length > 0) {
        html += '<div class="stat-item">';
        html += '<strong>Geschlecht</strong>';
        for (const [geschlecht, count] of Object.entries(group.Geschlechts)) {
            html += '<div>' + escapeHtml(geschlecht) + ': ' + count + '</div>';
        }
        html += '</div>';
    }
    
    html += '</div></div>';
    
    return html;
}

function checkForExistingScore() {
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

async function handleGlobalAssignScore() {
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
                showStationDetails(stationIndex >= 0 ? stationIndex : currentStationIndex);
            }
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Error saving score: ' + err);
    }
}

async function handleAssignScore(stationID) {
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
                showStationDetails(currentStationIndex);
            }
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Error saving score: ' + err);
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

async function handleEvaluation() {
    setStatus('Loading group evaluation...', 'info');
    
    try {
        const result = await window.go.main.App.GetGroupEvaluations();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Error loading evaluations: ' + result.message;
        } else {
            setStatus('Displaying group rankings', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            clearAllTabs();
            renderGroupEvaluations(result.evaluations);
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = 'Error: ' + err;
    }
}

function renderGroupEvaluations(evaluations) {
    // Clear existing tabs
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!evaluations || evaluations.length === 0) {
        tabContents.innerHTML = '<div style="padding: 20px;">No evaluations found.</div>';
        return;
    }
    
    // Create a single content area with the rankings table
    const content = document.createElement('div');
    content.style.cssText = 'padding: 20px;';
    
    let html = '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">';
    html += '<h2 style="margin: 0; color: #333;">🏆 Group Rankings - Total Scores</h2>';
    html += '<button onclick="handleGenerateGroupEvaluationPDF()" style="padding: 10px 20px; background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white; border: none; border-radius: 6px; font-weight: 600; cursor: pointer; font-size: 14px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">📄 Generate PDF</button>';
    html += '</div>';
    
    // Rankings table
    html += '<table class="group-table" style="max-width: 800px; margin: 0 auto;">';
    html += '<thead><tr>';
    html += '<th style="width: 100px; text-align: center;">Rank</th>';
    html += '<th>Group</th>';
    html += '<th style="width: 150px; text-align: center;">Stations Visited</th>';
    html += '<th style="width: 200px; text-align: center;">Total Score</th>';
    html += '</tr></thead><tbody>';
    
    evaluations.forEach((evaluation, index) => {
        const rankEmoji = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : (index + 1) + '.';
        const rowStyle = index < 3 ? 'background: #fff3cd;' : '';
        
        html += '<tr style="' + rowStyle + '">';
        html += '<td style="text-align: center; font-size: 24px;">' + rankEmoji + '</td>';
        html += '<td style="font-weight: bold; font-size: 18px;">Group ' + evaluation.GroupID + '</td>';
        html += '<td style="text-align: center; font-size: 16px;">' + evaluation.StationCount + '</td>';
        html += '<td style="text-align: center; font-weight: bold; font-size: 22px; color: #667eea;">' + evaluation.TotalScore + '</td>';
        html += '</tr>';
    });
    
    html += '</tbody></table>';
    
    // Statistics summary
    html += '<div class="stats-panel" style="max-width: 800px; margin: 20px auto 0;">';
    html += '<h3>📊 Summary Statistics</h3>';
    html += '<div class="stats-grid">';
    
    // Total groups
    html += '<div class="stat-item">';
    html += '<strong>Total Groups</strong>';
    html += '<span>' + evaluations.length + '</span>';
    html += '</div>';
    
    // Highest score
    html += '<div class="stat-item">';
    html += '<strong>Highest Score</strong>';
    html += '<span>' + evaluations[0].TotalScore + ' (Group ' + evaluations[0].GroupID + ')</span>';
    html += '</div>';
    
    // Lowest score
    const lastEval = evaluations[evaluations.length - 1];
    html += '<div class="stat-item">';
    html += '<strong>Lowest Score</strong>';
    html += '<span>' + lastEval.TotalScore + ' (Group ' + lastEval.GroupID + ')</span>';
    html += '</div>';
    
    // Average score
    const totalScore = evaluations.reduce((sum, e) => sum + e.TotalScore, 0);
    const avgScore = (totalScore / evaluations.length).toFixed(1);
    html += '<div class="stat-item">';
    html += '<strong>Average Score</strong>';
    html += '<span>' + avgScore + '</span>';
    html += '</div>';
    
    html += '</div></div>';
    
    content.innerHTML = html;
    tabContents.appendChild(content);
}

async function handleOrtsverbandEvaluation() {
    setStatus('Loading Ortsverband evaluation...', 'info');
    
    try {
        const result = await window.go.main.App.GetOrtsverbandEvaluations();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Error loading ortsverband evaluations: ' + result.message;
        } else {
            setStatus('Displaying ortsverband rankings', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            clearAllTabs();
            renderOrtsverbandEvaluations(result.evaluations);
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = 'Error: ' + err;
    }
}

function renderOrtsverbandEvaluations(evaluations) {
    // Clear existing tabs
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!evaluations || evaluations.length === 0) {
        tabContents.innerHTML = '<div style="padding: 20px;">No ortsverband evaluations found.</div>';
        return;
    }
    
    // Create a single content area with the rankings table
    const content = document.createElement('div');
    content.style.cssText = 'padding: 20px;';
    
    let html = '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">';
    html += '<h2 style="margin: 0; color: #333;">🏆 Ortsverband Rankings - Average Scores</h2>';
    html += '<button onclick="handleGenerateOrtsverbandEvaluationPDF()" style="padding: 10px 20px; background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white; border: none; border-radius: 6px; font-weight: 600; cursor: pointer; font-size: 14px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">📄 Generate PDF</button>';
    html += '</div>';
    
    // Rankings table
    html += '<table class="group-table" style="max-width: 1000px; margin: 0 auto;">';
    html += '<thead><tr>';
    html += '<th style="width: 100px; text-align: center;">Rank</th>';
    html += '<th>Ortsverband</th>';
    html += '<th style="width: 150px; text-align: center;">Participants</th>';
    html += '<th style="width: 180px; text-align: center;">Total Score</th>';
    html += '<th style="width: 180px; text-align: center;">Average Score</th>';
    html += '</tr></thead><tbody>';
    
    evaluations.forEach((evaluation, index) => {
        const rankEmoji = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : (index + 1) + '.';
        const rowStyle = index < 3 ? 'background: #fff3cd;' : '';
        
        html += '<tr style="' + rowStyle + '">';
        html += '<td style="text-align: center; font-size: 24px;">' + rankEmoji + '</td>';
        html += '<td style="font-weight: bold; font-size: 18px;">' + escapeHtml(evaluation.Ortsverband) + '</td>';
        html += '<td style="text-align: center; font-size: 16px;">' + evaluation.ParticipantCount + '</td>';
        html += '<td style="text-align: center; font-size: 18px;">' + evaluation.TotalScore + '</td>';
        html += '<td style="text-align: center; font-weight: bold; font-size: 22px; color: #667eea;">' + evaluation.AverageScore.toFixed(1) + '</td>';
        html += '</tr>';
    });
    
    html += '</tbody></table>';
    
    // Statistics summary
    html += '<div class="stats-panel" style="max-width: 900px; margin: 20px auto 0;">';
    html += '<h3>📊 Summary Statistics</h3>';
    html += '<div class="stats-grid">';
    
    // Total ortsverbands
    html += '<div class="stat-item">';
    html += '<strong>Total Ortsverbands</strong>';
    html += '<span>' + evaluations.length + '</span>';
    html += '</div>';
    
    // Highest score
    html += '<div class="stat-item">';
    html += '<strong>Highest Average Score</strong>';
    html += '<span>' + evaluations[0].AverageScore.toFixed(1) + ' (' + escapeHtml(evaluations[0].Ortsverband) + ')</span>';
    html += '</div>';
    
    // Lowest score
    const lastEval = evaluations[evaluations.length - 1];
    html += '<div class="stat-item">';
    html += '<strong>Lowest Average Score</strong>';
    html += '<span>' + lastEval.AverageScore.toFixed(1) + ' (' + escapeHtml(lastEval.Ortsverband) + ')</span>';
    html += '</div>';
    
    // Average score across all ortsverbands
    const totalScore = evaluations.reduce((sum, e) => sum + e.TotalScore, 0);
    const totalParticipants = evaluations.reduce((sum, e) => sum + e.ParticipantCount, 0);
    const overallAvg = totalParticipants > 0 ? (totalScore / totalParticipants).toFixed(1) : '0.0';
    html += '<div class="stat-item">';
    html += '<strong>Overall Average Score</strong>';
    html += '<span>' + overallAvg + '</span>';
    html += '</div>';
    
    html += '</div></div>';
    
    content.innerHTML = html;
    tabContents.appendChild(content);
}

async function handleShowStations() {
    setStatus('Loading stations...', 'info');
    
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
            setStatus('Displaying ' + stationsResult.count + ' stations', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            // Ensure complete cleanup before rendering
            clearAllTabs();
            renderStationTabs(stationsResult.stations, groupsResult.groups);
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = 'Error: ' + err;
    }
}

function renderStationTabs(stations, groups) {
    // Clear existing tabs - already done by clearAllTabs, but keep for safety
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!stations || stations.length === 0) {
        tabContents.innerHTML = '<div style="padding: 20px;">No stations found.</div>';
        return;
    }
    
    // Store stations and groups for later access
    window.currentStations = stations;
    window.currentGroups = groups;
    window.currentStationIndex = 0;
    
    // Create score entry form ABOVE station buttons
    const scoreForm = document.createElement('div');
    scoreForm.id = 'score-entry-form';
    scoreForm.style.cssText = 'background: #f0f8ff; padding: 20px; border-radius: 8px; margin: 20px; border: 2px solid #4facfe;';
    
    let formHtml = '<h3 style="margin: 0 0 15px 0; color: #333; font-size: 18px;">📝 Assign Score to Group</h3>';
    formHtml += '<div style="display: flex; gap: 15px; align-items: center; flex-wrap: wrap;">';
    
    // Group selector
    formHtml += '<label style="font-weight: 600; color: #333;">Group:</label>';
    formHtml += '<select id="global-group-select" style="padding: 10px; border-radius: 4px; border: 1px solid #ddd; font-size: 14px; min-width: 150px;">';
    formHtml += '<option value="">Select Group...</option>';
    if (groups && groups.length > 0) {
        groups.forEach(groupID => {
            formHtml += '<option value="' + groupID + '">Group ' + groupID + '</option>';
        });
    }
    formHtml += '</select>';
    
    // Score slider (1200 to 100 in steps of 50)
    formHtml += '<label style="font-weight: 600; color: #333;">Score:</label>';
    formHtml += '<div style="display: flex; align-items: center; gap: 10px;">';
    formHtml += '<input type="range" id="global-score-select" min="100" max="1200" step="50" value="650" style="width: 200px; height: 30px; cursor: pointer;" oninput="document.getElementById(\'score-display\').textContent = this.value">';
    formHtml += '<span id="score-display" style="font-weight: bold; color: #667eea; font-size: 18px; min-width: 60px; text-align: center;">650</span>';
    formHtml += '</div>';
    
    // Station display (read-only, shows which station is selected via buttons)
    formHtml += '<label style="font-weight: 600; color: #333;">Station:</label>';
    formHtml += '<input type="text" id="global-station-select" readonly data-station-id="' + stations[0].StationID + '" value="' + stations[0].StationName + '" style="padding: 10px; border-radius: 4px; border: 1px solid #ddd; font-size: 14px; min-width: 200px; background: #f5f5f5; color: #333; font-weight: 600;">';
    
    // Submit button
    formHtml += '<button onclick="handleGlobalAssignScore()" style="padding: 10px 20px; background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); color: white; border: none; border-radius: 4px; font-weight: 600; cursor: pointer; font-size: 14px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">Save Score</button>';
    
    formHtml += '</div>';
    
    // Warning indicator area
    formHtml += '<div id="score-warning" style="margin-top: 10px; padding: 10px; border-radius: 4px; display: none; background: #fff3cd; border: 2px solid #ffc107; color: #856404;">';
    formHtml += '<strong>⚠️ Warning:</strong> <span id="score-warning-text"></span>';
    formHtml += '</div>';
    
    scoreForm.innerHTML = formHtml;
    tabButtons.appendChild(scoreForm);
    
    // Add event listener to group selector to check for existing scores
    setTimeout(() => {
        const groupSelect = document.getElementById('global-group-select');
        if (groupSelect) {
            groupSelect.addEventListener('change', checkForExistingScore);
        }
    }, 0);
    
    // Create a 4-column grid of station buttons
    const buttonGrid = document.createElement('div');
    buttonGrid.className = 'station-button-grid';
    buttonGrid.style.cssText = 'display: grid; grid-template-columns: repeat(4, 1fr); gap: 15px; padding: 20px;';
    
    stations.forEach((station, index) => {
        const button = document.createElement('button');
        button.className = 'station-grid-button';
        button.style.cssText = 'padding: 20px; font-size: 18px; font-weight: 600; min-width: 0; background: linear-gradient(135deg, #fbc2eb 0%, #a6c1ee 100%); color: white; border: none; border-radius: 8px; cursor: pointer; transition: all 0.3s; box-shadow: 0 4px 6px rgba(0,0,0,0.1);';
        button.textContent = station.StationName;
        button.onclick = () => showStationDetails(index);
        button.onmouseover = () => { button.style.transform = 'translateY(-2px)'; button.style.boxShadow = '0 6px 12px rgba(0,0,0,0.15)'; };
        button.onmouseout = () => { button.style.transform = 'translateY(0)'; button.style.boxShadow = '0 4px 6px rgba(0,0,0,0.1)'; };
        buttonGrid.appendChild(button);
    });
    
    tabButtons.appendChild(buttonGrid);
    
    // Create content area for station details
    const contentArea = document.createElement('div');
    contentArea.id = 'station-content-area';
    contentArea.style.cssText = 'padding: 20px; border-top: 2px solid #ddd; display: none;';
    tabContents.appendChild(contentArea);
    
    // Show first station by default
    showStationDetails(0);
}

function showStationDetails(stationIndex) {
    const stations = window.currentStations;
    const groups = window.currentGroups;
    const contentArea = document.getElementById('station-content-area');
    
    if (!stations || !contentArea) return;
    
    window.currentStationIndex = stationIndex;
    const station = stations[stationIndex];
    
    // Update station display in the form
    const stationSelect = document.getElementById('global-station-select');
    if (stationSelect) {
        stationSelect.value = station.StationName;
        stationSelect.setAttribute('data-station-id', station.StationID);
    }
    
    // Check for existing score when station changes
    checkForExistingScore();
    
    contentArea.innerHTML = formatStationContent(station, groups);
    contentArea.style.display = 'block';
    
    // Update button states
    const buttons = document.querySelectorAll('.station-grid-button');
    buttons.forEach((btn, index) => {
        if (index === stationIndex) {
            btn.style.background = 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)';
            btn.style.boxShadow = '0 6px 12px rgba(0,0,0,0.2)';
        } else {
            btn.style.background = 'linear-gradient(135deg, #fbc2eb 0%, #a6c1ee 100%)';
            btn.style.boxShadow = '0 4px 6px rgba(0,0,0,0.1)';
        }
    });
}

function formatStationContent(station, groups) {
    let html = '<h2 style="margin-bottom: 15px; color: #333;">🏆 ' + escapeHtml(station.StationName) + '</h2>';
    
    // Group scores table (score entry form is now at the top)
    html += '<table class="group-table">';
    html += '<thead><tr>';
    html += '<th style="width: 80px;">Rank</th>';
    html += '<th>Group</th>';
    html += '<th style="width: 150px;">Score</th>';
    html += '</tr></thead><tbody>';
    
    if (station.GroupScores && station.GroupScores.length > 0) {
        station.GroupScores.forEach((groupScore, index) => {
            const rankEmoji = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : (index + 1) + '.';
            html += '<tr>';
            html += '<td style="text-align: center; font-size: 18px;">' + rankEmoji + '</td>';
            html += '<td>Group ' + groupScore.GroupID + '</td>';
            html += '<td style="font-weight: bold; font-size: 18px;">' + groupScore.Score + '</td>';
            html += '</tr>';
        });
    } else {
        html += '<tr><td colspan="3" style="text-align: center; padding: 20px; color: #999;">No scores recorded yet</td></tr>';
    }
    
    html += '</tbody></table>';
    
    // Statistics panel
    if (station.GroupScores && station.GroupScores.length > 0) {
        html += '<div class="stats-panel">';
        html += '<h3>📊 Station Statistics</h3>';
        html += '<div class="stats-grid">';
        
        html += '<div class="stat-item">';
        html += '<strong>Total Groups</strong>';
        html += '<span>' + station.GroupScores.length + '</span>';
        html += '</div>';
        
        // Calculate average score
        const totalScore = station.GroupScores.reduce((sum, gs) => sum + gs.Score, 0);
        const avgScore = (totalScore / station.GroupScores.length).toFixed(1);
        html += '<div class="stat-item">';
        html += '<strong>Average Score</strong>';
        html += '<span>' + avgScore + '</span>';
        html += '</div>';
        
        // Highest score
        html += '<div class="stat-item">';
        html += '<strong>Highest Score</strong>';
        html += '<span>' + station.GroupScores[0].Score + ' (Group ' + station.GroupScores[0].GroupID + ')</span>';
        html += '</div>';
        
        // Lowest score
        const lastScore = station.GroupScores[station.GroupScores.length - 1];
        html += '<div class="stat-item">';
        html += '<strong>Lowest Score</strong>';
        html += '<span>' + lastScore.Score + ' (Group ' + lastScore.GroupID + ')</span>';
        html += '</div>';
        
        html += '</div></div>';
    }
    
    return html;
}

async function handleGeneratePDF() {
    setStatus('Generating PDF report...', 'info');
    
    try {
        const result = await window.go.main.App.GeneratePDF();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
        } else {
            setStatus('✔ ' + result.message + ': ' + result.file, 'success');
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
    }
}

async function handleGenerateGroupEvaluationPDF() {
    setStatus('Generating group evaluation PDF...', 'info');
    
    try {
        const result = await window.go.main.App.GenerateGroupEvaluationPDF();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            alert('Failed to generate PDF: ' + result.message);
        } else {
            setStatus('✔ ' + result.message + ': ' + result.file, 'success');
            alert('PDF generated successfully!\n\nFile: ' + result.file + '\nLocation: ' + result.path);
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Error generating PDF: ' + err);
    }
}

async function handleGenerateOrtsverbandEvaluationPDF() {
    setStatus('Generating ortsverband evaluation PDF...', 'info');
    
    try {
        const result = await window.go.main.App.GenerateOrtsverbandEvaluationPDF();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            alert('Failed to generate PDF: ' + result.message);
        } else {
            setStatus('✔ ' + result.message + ': ' + result.file, 'success');
            alert('PDF generated successfully!\n\nFile: ' + result.file + '\nLocation: ' + result.path);
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
        alert('Error generating PDF: ' + err);
    }
}

// Initial instructions
output.textContent = 'Willkommen bei der Jugendolympiade!\n\nINSTRUCTIONS:\n• Click "Load Excel File" to select and upload your .xlsx file\n  - File will be imported and groups will be automatically created\n• Click "Show Groups" to view the balanced groups\n• Click "Auswertung nach Gruppen" for group evaluation\n• Click "Auswertung nach Ortsverband" for location-based evaluation\n• Click "Generate PDF" to export the groups to a PDF file\n\nReady to begin!';
