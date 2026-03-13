// Evaluation management and display
import { setStatus, output, tabs, tabButtons, tabContents, clearAllTabs } from '../shared/dom.js';
import { escapeHtml } from '../shared/utils.js';

export async function handleGroupEvaluation() {
    setStatus('Loading group evaluations...', 'info');
    
    try {
        const result = await window.go.main.App.GetGroupEvaluations();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Error loading group evaluations: ' + result.message;
        } else {
            setStatus('Displaying evaluations for ' + result.count + ' groups', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            // Ensure complete cleanup before rendering
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

export async function handleOrtsverbandEvaluation() {
    setStatus('Loading Ortsverband evaluations...', 'info');
    
    try {
        const result = await window.go.main.App.GetOrtsverbandEvaluations();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Error loading Ortsverband evaluations: ' + result.message;
        } else {
            setStatus('Displaying evaluations for ' + result.count + ' Ortsverbände', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            // Ensure complete cleanup before rendering
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

function renderGroupEvaluations(evaluations) {
    // Clear existing tabs - already done by clearAllTabs, but keep for safety
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!evaluations || evaluations.length === 0) {
        tabContents.innerHTML = '<div class="empty-message">No evaluations found.</div>';
        return;
    }
    
    // Create single content area for all evaluations
    const contentArea = document.createElement('div');
    contentArea.className = 'evaluation-content';
    
    let html = '<div class="evaluation-header">';
    html += '<h2 class="evaluation-title">🏆 Gruppenrangliste - Gesamtergebnis</h2>';
    html += '<button onclick="handleGenerateGroupEvaluationPDF()" class="btn-pdf-generate">📄 Generate PDF</button>';
    html += '</div>';
    
    // Evaluation table
    html += '<table class="group-table evaluation-table">';
    html += '<thead><tr>';
    html += '<th class="col-rank">Rang</th>';
    html += '<th class="col-group">Gruppe</th>';
    html += '<th>Stationen</th>';
    html += '<th>Gesamtergebnis</th>';
    html += '</tr></thead><tbody>';
    
    evaluations.forEach((evalItem, index) => {
        const rankEmoji = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : (index + 1) + '.';
        const rowClass = index < 3 ? 'podium-row' : '';
        html += '<tr class="' + rowClass + '">';
        html += '<td class="rank-cell">' + rankEmoji + '</td>';
        html += '<td class="group-cell">Gruppe ' + evalItem.GroupID + '</td>';
        html += '<td class="station-count-cell">' + evalItem.StationCount + '</td>';
        html += '<td class="total-score-cell">' + evalItem.TotalScore + '</td>';
        html += '</tr>';
    });
    
    html += '</tbody></table>';
    
    // Statistics panel
    html += '<div class="stats-panel">';
    html += '<h3>📊 Overall Statistics</h3>';
    html += '<div class="stats-grid">';
    
    html += '<div class="stat-item">';
    html += '<strong>Gruppen gesamt</strong>';
    html += '<span>' + evaluations.length + '</span>';
    html += '</div>';
    
    // Calculate overall average
    const totalScore = evaluations.reduce((sum, e) => sum + e.TotalScore, 0);
    const overallAvg = (totalScore / evaluations.length).toFixed(1);
    html += '<div class="stat-item">';
    html += '<strong>Durchschnittsergebnis</strong>';
    html += '<span>' + overallAvg + '</span>';
    html += '</div>';
    
    // Highest score
    html += '<div class="stat-item">';
    html += '<strong>Höchstes Ergebnis</strong>';
    html += '<span>' + evaluations[0].TotalScore + ' (Gruppe ' + evaluations[0].GroupID + ')</span>';
    html += '</div>';
    
    // Lowest score
    const lastEval = evaluations[evaluations.length - 1];
    html += '<div class="stat-item">';
    html += '<strong>Niedrigstes Ergebnis</strong>';
    html += '<span>' + lastEval.TotalScore + ' (Gruppe ' + lastEval.GroupID + ')</span>';
    html += '</div>';
    
    html += '</div></div>';
    
    contentArea.innerHTML = html;
    tabContents.appendChild(contentArea);
}

function renderOrtsverbandEvaluations(evaluations) {
    // Clear existing tabs - already done by clearAllTabs, but keep for safety
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!evaluations || evaluations.length === 0) {
        tabContents.innerHTML = '<div class="empty-message">No evaluations found.</div>';
        return;
    }
    
    // Create single content area for all evaluations
    const contentArea = document.createElement('div');
    contentArea.className = 'evaluation-content';
    
    let html = '<div class="evaluation-header">';
    html += '<h2 class="evaluation-title">🏆 Ortsverband Rangliste - Durchschnittsergebnisse</h2>';
    html += '<button onclick="handleGenerateOrtsverbandEvaluationPDF()" class="btn-pdf-generate">📄 Generate PDF</button>';
    html += '</div>';
    
    // Evaluation table
    html += '<table class="group-table evaluation-table">';
    html += '<thead><tr>';
    html += '<th class="col-rank">Rang</th>';
    html += '<th>Ortsverband</th>';
    html += '<th>Teilnehmer</th>';
    html += '<th>Gesamtergebnis</th>';
    html += '<th>Durchschnitt</th>';
    html += '</tr></thead><tbody>';
    
    evaluations.forEach((evalItem, index) => {
        const rankEmoji = index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : (index + 1) + '.';
        const rowClass = index < 3 ? 'podium-row' : '';
        html += '<tr class="' + rowClass + '">';
        html += '<td class="rank-cell">' + rankEmoji + '</td>';
        html += '<td class="ortsverband-cell">' + escapeHtml(evalItem.Ortsverband) + '</td>';
        html += '<td class="text-center">' + evalItem.ParticipantCount + '</td>';
        html += '<td class="text-center">' + evalItem.TotalScore + '</td>';
        html += '<td class="average-score-cell">' + evalItem.AverageScore.toFixed(1) + '</td>';
        html += '</tr>';
    });
    
    html += '</tbody></table>';
    
    // Statistics panel
    html += '<div class="stats-panel">';
    html += '<h3>📊 Overall Statistics</h3>';
    html += '<div class="stats-grid">';
    
    html += '<div class="stat-item">';
    html += '<strong>Total Ortsverbände</strong>';
    html += '<span>' + evaluations.length + '</span>';
    html += '</div>';
    
    // Highest average score
    html += '<div class="stat-item">';
    html += '<strong>Höchstes Durchschnittsergebnis</strong>';
    html += '<span>' + evaluations[0].AverageScore.toFixed(1) + ' (' + escapeHtml(evaluations[0].Ortsverband) + ')</span>';
    html += '</div>';
    
    // Lowest average score
    const lastEval = evaluations[evaluations.length - 1];
    html += '<div class="stat-item">';
    html += '<strong>Niedrigstes Durchschnittsergebnis</strong>';
    html += '<span>' + lastEval.AverageScore.toFixed(1) + ' (' + escapeHtml(lastEval.Ortsverband) + ')</span>';
    html += '</div>';
    
    // Overall average score across all ortsverbände
    const totalScore = evaluations.reduce((sum, e) => sum + e.TotalScore, 0);
    const totalParticipants = evaluations.reduce((sum, e) => sum + e.ParticipantCount, 0);
    const overallAvg = totalParticipants > 0 ? (totalScore / totalParticipants).toFixed(1) : '0.0';
    html += '<div class="stat-item">';
    html += '<strong>Overall Average Score</strong>';
    html += '<span>' + overallAvg + '</span>';
    html += '</div>';
    
    html += '</div></div>';
    
    contentArea.innerHTML = html;
    tabContents.appendChild(contentArea);
}
