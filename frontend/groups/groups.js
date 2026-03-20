// Group display and rendering
import { setStatus, output, tabs, tabButtons, tabContents, clearAllTabs } from '../shared/dom.js';
import { escapeHtml, switchTab } from '../shared/utils.js';

export async function handleShowGroups() {
    setStatus('Gruppen werden geladen...', 'info');
    
    try {
        const result = await window.go.main.App.ShowGroups();
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            output.style.display = 'block';
            tabs.style.display = 'none';
            output.textContent = 'Fehler beim Laden der Gruppen: ' + result.message;
        } else {
            setStatus(result.count + ' ausgewogene Gruppen werden angezeigt', 'success');
            output.style.display = 'none';
            tabs.style.display = 'block';
            // Ensure complete cleanup before rendering
            clearAllTabs();
            renderGroupTabs(result.groups);
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
        output.style.display = 'block';
        tabs.style.display = 'none';
        output.textContent = 'Fehler: ' + err;
    }
}

function renderGroupTabs(groups) {
    // Clear existing tabs - already done by clearAllTabs, but keep for safety
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    
    if (!groups || groups.length === 0) {
        tabContents.innerHTML = '<div class="empty-message">Keine Gruppen gefunden.</div>';
        return;
    }
    
    // Create tabs for each group
    groups.forEach((group, index) => {
        // Create tab button
        const button = document.createElement('button');
        button.className = 'tab-button' + (index === 0 ? ' active' : '');
        button.textContent = 'Gruppe ' + group.GroupID;
        button.onclick = () => switchTab(index, tabButtons, tabContents);
        tabButtons.appendChild(button);
        
        // Create tab content
        const content = document.createElement('div');
        content.className = 'tab-content' + (index === 0 ? ' active' : '');
        content.innerHTML = formatGroupContent(group);
        tabContents.appendChild(content);
    });
}

function formatGroupContent(group) {
    let html = '<div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 15px;">';
    html += '<h2 class="group-title" style="margin: 0;">Gruppe ' + group.GroupID + '</h2>';
    html += '<button onclick="window.handleShowStationsForGroup(' + group.GroupID + ')" class="btn-stations">📝 Ergebniseingabe</button>';
    html += '</div>';
    
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
        html += '<tr><td colspan="4">Keine Teilnehmer</td></tr>';
    }
    
    html += '</tbody></table>';

    // Betreuende section
    if (group.Betreuende && group.Betreuende.length > 0) {
        html += '<h3 style="margin: 20px 0 10px 0; color: #555;">👥 Betreuende</h3>';
        html += '<table class="group-table betreuende-table">';
        html += '<thead><tr><th>Name</th><th>Ortsverband</th></tr></thead><tbody>';
        group.Betreuende.forEach(b => {
            html += '<tr class="betreuende-row">';
            html += '<td>' + escapeHtml(b.Name) + '</td>';
            html += '<td>' + escapeHtml(b.Ortsverband) + '</td>';
            html += '</tr>';
        });
        html += '</tbody></table>';
    }

    // Statistics panel
    html += '<div class="stats-panel">';
    html += '<h3>📊 Gruppenstatistik</h3>';
    html += '<div class="stats-grid">';
    
    // Total participants
    html += '<div class="stat-item">';
    html += '<strong>Teilnehmer gesamt</strong>';
    html += '<span>' + (group.Teilnehmers ? group.Teilnehmers.length : 0) + '</span>';
    html += '</div>';
    
    // Average age
    if (group.Teilnehmers && group.Teilnehmers.length > 0) {
        const avgAge = (group.AlterSum / group.Teilnehmers.length).toFixed(1);
        html += '<div class="stat-item">';
        html += '<strong>Durchschnittsalter</strong>';
        html += '<span>' + avgAge + ' Jahre</span>';
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
