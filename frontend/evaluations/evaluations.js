// Evaluation management and display
import { setStatus, output, tabs, tabButtons, tabContents, clearAllTabs } from '../shared/dom.js';
import { escapeHtml } from '../shared/utils.js';

export async function handleGroupEvaluation() {
    setStatus('Gruppenauswertungen werden geladen...', 'info');
    
    try {
        const result = await window.go.main.App.GetGroupEvaluations();
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Fehler beim Laden der Gruppenauswertungen: ' + result.message;
        } else {
            setStatus('Auswertung für ' + result.evaluations.length + ' Gruppen wird angezeigt', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            // Ensure complete cleanup before rendering
            clearAllTabs();
            renderGroupEvaluations(result.evaluations);
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = 'Fehler: ' + err;
    }
}

export async function handleOrtsverbandEvaluation() {
    setStatus('Ortsverband-Auswertungen werden geladen...', 'info');
    
    try {
        const result = await window.go.main.App.GetOrtsverbandEvaluations();
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Fehler beim Laden der Ortsverband-Auswertungen: ' + result.message;
        } else {
            setStatus('Auswertung für ' + result.evaluations.length + ' Ortsverbände wird angezeigt', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            // Ensure complete cleanup before rendering
            clearAllTabs();
            renderOrtsverbandEvaluations(result.evaluations);
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = 'Fehler: ' + err;
    }
}

function renderGroupEvaluations(evaluations) {
    // Clear existing tabs - already done by clearAllTabs, but keep for safety
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!evaluations || evaluations.length === 0) {
        tabContents.innerHTML = '<div class="empty-message">Keine Auswertungen gefunden.</div>';
        return;
    }
    
    // Create single content area for all evaluations
    const contentArea = document.createElement('div');
    contentArea.className = 'evaluation-content';
    
    let html = '<div class="evaluation-header">';
    html += '<h2 class="evaluation-title">🏆 Gruppenrangliste - Gesamtergebnis</h2>';
    html += '<button onclick="handleGenerateGroupEvaluationPDF()" class="btn-pdf-generate">📄 PDF erstellen</button>';
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
    html += '<h3>📊 Gesamtstatistik</h3>';
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
        tabContents.innerHTML = '<div class="empty-message">Keine Auswertungen gefunden.</div>';
        return;
    }
    
    // Create single content area for all evaluations
    const contentArea = document.createElement('div');
    contentArea.className = 'evaluation-content';
    
    let html = '<div class="evaluation-header">';
    html += '<h2 class="evaluation-title">🏆 Ortsverband Rangliste</h2>';
    html += '<div class="evaluation-header-actions">';
    html += '<button id="btnToggleScores" onclick="toggleOVScores()" class="btn-toggle-scores">🔒 Ergebnisse einblenden</button>';
    html += '<button onclick="handleGenerateOrtsverbandEvaluationPDF()" class="btn-pdf-generate">📄 PDF erstellen</button>';
    html += '</div>';
    html += '</div>';
    
    // Evaluation table
    html += '<table class="group-table evaluation-table" id="ovEvalTable">';
    html += '<thead><tr>';
    html += '<th>Ortsverband</th>';
    html += '<th>Teilnehmende</th>';
    html += '<th class="ov-score-col" style="display:none">Gesamtergebnis</th>';
    html += '<th class="ov-score-col" style="display:none">Durchschnitt</th>';
    html += '</tr></thead><tbody>';
    
    evaluations.forEach((evalItem, index) => {
        const isWinner = index === 0;
        const rowClass = isWinner ? 'podium-row' : '';
        const nameCell = isWinner
            ? '🏆 ' + escapeHtml(evalItem.Ortsverband) + ' <span class="winner-badge">Bester Ortsverband</span>'
            : escapeHtml(evalItem.Ortsverband);
        html += '<tr class="' + rowClass + '">';
        html += '<td class="ortsverband-cell">' + nameCell + '</td>';
        html += '<td class="text-center">' + evalItem.ParticipantCount + '</td>';
        html += '<td class="text-center ov-score-col" style="display:none">' + evalItem.TotalScore + '</td>';
        html += '<td class="average-score-cell ov-score-col" style="display:none">' + evalItem.AverageScore.toFixed(1) + '</td>';
        html += '</tr>';
    });
    
    html += '</tbody></table>';
    
    // Statistics panel — no scores shown here
    html += '<div class="stats-panel">';
    html += '<h3>📊 Gesamtstatistik</h3>';
    html += '<div class="stats-grid">';
    
    html += '<div class="stat-item">';
    html += '<strong>Ortsverbände gesamt</strong>';
    html += '<span>' + evaluations.length + '</span>';
    html += '</div>';

    const totalParticipants = evaluations.reduce((sum, e) => sum + e.ParticipantCount, 0);
    html += '<div class="stat-item">';
    html += '<strong>Teilnehmende gesamt</strong>';
    html += '<span>' + totalParticipants + '</span>';
    html += '</div>';
    
    html += '</div></div>';
    
    contentArea.innerHTML = html;
    tabContents.appendChild(contentArea);
}

export function toggleOVScores() {
    const cols = document.querySelectorAll('.ov-score-col');
    const btn = document.getElementById('btnToggleScores');
    const hidden = cols.length > 0 && cols[0].style.display === 'none';
    cols.forEach(el => { el.style.display = hidden ? '' : 'none'; });
    btn.textContent = hidden ? '🔓 Ergebnisse ausblenden' : '🔒 Ergebnisse einblenden';
}
