// participants.js — Teilnehmer hinzufügen / entfernen (admin tab)
import { setStatus, output, tabs, tabButtons, tabContents, clearAllTabs } from '../shared/dom.js';
import { escapeHtml, switchTab } from '../shared/utils.js';

let _scoresLocked = false;
let _eligibleGroups = [];

export async function handleParticipantManagement() {
    setStatus('Teilnehmerverwaltung wird geladen…', 'info');
    try {
        const [pResult, eResult] = await Promise.all([
            window.go.main.App.GetParticipantsWithGroups(),
            window.go.main.App.GetEligibleGroups(),
        ]);

        if (pResult.status === 'error') {
            setStatus('FEHLER: ' + pResult.message, 'error');
            return;
        }

        _scoresLocked = !!pResult.scoresLocked;
        _eligibleGroups = (eResult.status === 'success' && eResult.groups) ? eResult.groups : [];

        setStatus('Teilnehmerverwaltung', 'info');
        output.style.display = 'none';
        tabs.style.display = 'block';
        clearAllTabs();
        _renderTabs(pResult.rows || []);
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

function _renderTabs(participants) {
    const btnEntfernen = document.createElement('button');
    btnEntfernen.className = 'tab-button active';
    btnEntfernen.textContent = 'Entfernen (' + participants.length + ')';
    btnEntfernen.onclick = () => switchTab(0, tabButtons, tabContents);
    tabButtons.appendChild(btnEntfernen);

    const btnHinzufuegen = document.createElement('button');
    btnHinzufuegen.className = 'tab-button';
    btnHinzufuegen.textContent = 'Hinzufügen';
    btnHinzufuegen.onclick = () => switchTab(1, tabButtons, tabContents);
    tabButtons.appendChild(btnHinzufuegen);

    const content1 = document.createElement('div');
    content1.className = 'tab-content active';
    content1.innerHTML = _renderEntfernenSection(participants);
    tabContents.appendChild(content1);

    const content2 = document.createElement('div');
    content2.className = 'tab-content';
    content2.innerHTML = _renderHinzufuegenSection();
    tabContents.appendChild(content2);

    const searchInput = content1.querySelector('#_tnSearch');
    if (searchInput) {
        searchInput.addEventListener('input', () => _filterTable(searchInput.value, content1));
    }
}

function _renderEntfernenSection(participants) {
    let html = '<div style="margin-bottom:16px">';
    if (_scoresLocked) {
        html += '<div style="background:#fff3e0;border:1px solid #ff9800;border-radius:6px;padding:12px;margin-bottom:12px;color:#e65100;">'
              + '⚠️ Hinzufügen und Entfernen ist nicht mehr möglich, da bereits Wertungen eingetragen wurden.</div>';
    }
    html += '<input type="text" id="_tnSearch" placeholder="Suche nach Name oder Ortsverband…"'
          + ' style="width:100%;padding:8px 10px;border:2px solid #ddd;border-radius:6px;font-size:14px;box-sizing:border-box;margin-bottom:12px;">';
    html += '</div>';

    html += '<table class="group-table" id="_tnTable">';
    html += '<thead><tr><th>Name</th><th>Ortsverband</th><th>Alter</th><th>Geschlecht</th><th>Gruppe</th><th></th></tr></thead>';
    html += '<tbody>';

    if (!participants || participants.length === 0) {
        html += '<tr><td colspan="6" class="empty-message">Keine Teilnehmenden gefunden.</td></tr>';
    } else {
        participants.forEach(p => {
            const safeName = escapeHtml(p.Name || '');
            const safeOV   = escapeHtml(p.Ortsverband || '');
            const safeGrp  = escapeHtml(p.GroupName || (p.GroupID > 0 ? 'Gruppe ' + p.GroupID : '—'));
            const disabled = _scoresLocked
                ? ' disabled title="Es wurden bereits Wertungen eingetragen."'
                : '';
            // Use data attributes for the live filter; escape single-quotes for onclick string.
            const jsName = safeName.replace(/'/g, "\\'");
            const jsGrp  = safeGrp.replace(/'/g, "\\'");
            html += '<tr data-name="' + safeName.toLowerCase() + '" data-ov="' + safeOV.toLowerCase() + '">';
            html += '<td>' + safeName + '</td>';
            html += '<td>' + safeOV + '</td>';
            html += '<td>' + (p.Alter || '') + '</td>';
            html += '<td>' + escapeHtml(p.Geschlecht || '') + '</td>';
            html += '<td>' + safeGrp + '</td>';
            html += '<td><button class="btn-remove"' + disabled
                  + ' onclick="window._tnRemove(' + p.ID + ',\'' + jsName + '\',\'' + jsGrp + '\')">'
                  + '🗑 Entfernen</button></td>';
            html += '</tr>';
        });
    }

    html += '</tbody></table>';
    return html;
}

function _renderHinzufuegenSection() {
    if (_scoresLocked) {
        return '<div style="background:#fff3e0;border:1px solid #ff9800;border-radius:6px;padding:16px;color:#e65100;">'
             + '⚠️ Hinzufügen und Entfernen ist nicht mehr möglich, da bereits Wertungen eingetragen wurden.</div>';
    }
    if (!_eligibleGroups || _eligibleGroups.length === 0) {
        return '<div style="background:#f3f3f3;border-radius:6px;padding:16px;color:#555;">'
             + 'Keine Gruppe hat einen freien Platz (inkl. Fahrgemeinschaft).</div>';
    }

    const groupOptions = _eligibleGroups.map(g =>
        '<option value="' + g.GroupID + '">'
        + escapeHtml(g.GroupName) + ' — ' + g.CurrentCount + '/' + g.MaxSlots + ' Plätze'
        + '</option>'
    ).join('');

    const inputStyle = 'display:block;width:100%;padding:8px 10px;border:2px solid #ddd;border-radius:6px;font-size:14px;box-sizing:border-box;margin-top:4px;';

    return '<div style="max-width:480px;">'
         + '<h3 style="margin-top:0">Neue/n Teilnehmer/in hinzufügen</h3>'
         + '<div style="display:grid;gap:14px;">'
         + '<label>Name*<input type="text" id="_tnName" placeholder="Vollständiger Name" style="' + inputStyle + '"></label>'
         + '<label>Ortsverband<input type="text" id="_tnOV" placeholder="Ortsverband" style="' + inputStyle + '"></label>'
         + '<label>Alter*<input type="number" id="_tnAlter" min="6" max="99" placeholder="Alter" style="' + inputStyle + '"></label>'
         + '<label>Geschlecht*'
         +   '<select id="_tnGeschlecht" style="' + inputStyle + '">'
         +     '<option value="">— bitte wählen —</option>'
         +     '<option value="m">männlich</option>'
         +     '<option value="w">weiblich</option>'
         +     '<option value="d">divers</option>'
         +   '</select>'
         + '</label>'
         + '<label>Gruppe* '
         +   '<small style="color:#888;font-weight:normal;">(nur Gruppen mit freiem Platz inkl. Fahrgemeinschaft)</small>'
         +   '<select id="_tnGroup" style="' + inputStyle + '">' + groupOptions + '</select>'
         + '</label>'
         + '<button onclick="window._tnAdd()" style="padding:10px 20px;background:#1976d2;color:#fff;border:none;border-radius:6px;cursor:pointer;font-weight:600;font-size:14px;margin-top:4px;">✚ Hinzufügen</button>'
         + '</div>'
         + '</div>';
}

function _filterTable(query, container) {
    const q = (query || '').toLowerCase();
    const rows = container.querySelectorAll('#_tnTable tbody tr[data-name]');
    rows.forEach(row => {
        const match = (row.dataset.name || '').includes(q) || (row.dataset.ov || '').includes(q);
        row.style.display = match ? '' : 'none';
    });
}

window._tnRemove = async function(id, name, groupName) {
    if (!confirm('Teilnehmer/in "' + name + '" aus Gruppe "' + (groupName || '—') + '" wirklich entfernen?\nDiese Aktion kann nicht rückgängig gemacht werden.')) return;

    setStatus('Wird entfernt…', 'info');
    try {
        const result = await window.go.main.App.RemoveTeilnehmer(id);
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            _showResultModal('error', result.message, []);
        } else {
            setStatus('✅ ' + result.message, 'success');
            _showResultModal('success', result.message, result.pdfResults || []);
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
};

window._tnAdd = async function() {
    const name       = (document.getElementById('_tnName')?.value || '').trim();
    const ov         = (document.getElementById('_tnOV')?.value || '').trim();
    const alterStr   = document.getElementById('_tnAlter')?.value || '';
    const geschlecht = document.getElementById('_tnGeschlecht')?.value || '';
    const groupIDStr = document.getElementById('_tnGroup')?.value || '';

    if (!name)       { alert('Bitte Name eingeben.'); return; }
    if (!geschlecht) { alert('Bitte Geschlecht wählen.'); return; }
    const alter = parseInt(alterStr, 10);
    if (!alter || alter <= 0) { alert('Bitte gültiges Alter eingeben.'); return; }
    const groupID = parseInt(groupIDStr, 10);
    if (!groupID)    { alert('Bitte Gruppe wählen.'); return; }

    setStatus('Wird hinzugefügt…', 'info');
    try {
        const result = await window.go.main.App.AddTeilnehmer(name, ov, alter, geschlecht, groupID);
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
            _showResultModal('error', result.message, []);
        } else {
            setStatus('✅ ' + result.message, 'success');
            _showResultModal('success', result.message, result.pdfResults || [], true);
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
};

function _showResultModal(type, message, pdfResults, refresh) {
    const overlay = document.createElement('div');
    overlay.style.cssText = 'position:fixed;inset:0;background:rgba(0,0,0,.55);display:flex;align-items:center;justify-content:center;z-index:9999';

    const box = document.createElement('div');
    box.style.cssText = 'background:#fff;border-radius:10px;padding:28px 32px;max-width:520px;width:90%;box-shadow:0 8px 32px rgba(0,0,0,.3);font-family:Arial,sans-serif';

    const icon = type === 'success' ? '✅' : '❌';
    let pdfHtml = '';
    if (pdfResults && pdfResults.length > 0) {
        pdfHtml = '<p style="margin:14px 0 6px;font-weight:600;font-size:0.9em;">Automatisch neu erstellt:</p>'
                + '<ul style="margin:0;padding-left:20px;font-size:0.9em;">';
        pdfResults.forEach(r => {
            const ok    = r.Status === 'ok';
            const color = ok ? '#2e7d32' : '#e65100';
            pdfHtml += '<li style="color:' + color + '">'
                     + (ok ? '✔' : '⚠️') + ' ' + escapeHtml(r.Name || '')
                     + (r.Error ? ' — ' + escapeHtml(r.Error) : '')
                     + '</li>';
        });
        pdfHtml += '</ul>';
    }

    box.innerHTML = '<h2 style="margin:0 0 12px;font-size:1.1em;">' + icon + ' ' + escapeHtml(message) + '</h2>'
                  + pdfHtml
                  + '<div style="display:flex;justify-content:flex-end;margin-top:20px;">'
                  + '<button id="_tnModalClose" style="padding:9px 20px;background:#1976d2;color:#fff;border:none;border-radius:6px;cursor:pointer;font-weight:600">Schließen</button>'
                  + '</div>';

    overlay.appendChild(box);
    document.body.appendChild(overlay);

    box.querySelector('#_tnModalClose').addEventListener('click', async () => {
        document.body.removeChild(overlay);
        // Always refresh so the list and dropdown stay current.
        await handleParticipantManagement();
    });
}
