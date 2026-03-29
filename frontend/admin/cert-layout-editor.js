// cert-layout-editor.js — Graphical certificate layout editor panel.
// Imported by config-editor.js; all internal symbols are prefixed _gfx.

// ---- module state -------------------------------------------------------
let _gfxData         = null;   // CertLayoutFile as JS object
let _gfxVariant      = 'participant';
let _gfxImageList    = [];     // filenames from ListBackgroundImages (root dir)
let _gfxGroupPicList = [];     // filenames from ListGroupPictures (picture dir)
let _gfxGroupPicDir  = '';     // picture directory name (for display)
const _gfxExpanded = new Set();
const _gfxImgCache    = new Map(); // filename → base64 data URL  (root dir)
const _gfxGroupCache  = new Map(); // filename → base64 data URL  (picture dir)

const _gfxTemplatePrefix = 'templates/';

function _gfxBgBaseName(path) {
    if (!path) return '';
    return String(path).split(/[/\\]/).pop() || '';
}

function _gfxBgStorePath(fileName) {
    if (!fileName) return '';
    return _gfxTemplatePrefix + _gfxBgBaseName(fileName);
}

const _GFX_VARIANTS = {
    participant:         'Teilnehmende',
    participant_picture: 'Teiln. (Foto)',
    ov_winner:           'OV Sieger',
    ov_participant:      'OV Teilnahme',
};

// ---- helpers ------------------------------------------------------------
function _escHtml(s) {
    return String(s ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
function _rgbToHex(rgb) {
    if (!rgb || rgb.length < 3) return '#000000';
    return '#' + rgb.map(v => Math.max(0, Math.min(255, v)).toString(16).padStart(2, '0')).join('');
}
function _hexToRgb(hex) {
    return [parseInt(hex.slice(1,3),16), parseInt(hex.slice(3,5),16), parseInt(hex.slice(5,7),16)];
}

// ---- public API ---------------------------------------------------------
export function getCertLayoutData()     { return _gfxData; }
export function resetCertLayoutEditor() {
    _gfxData         = null;
    _gfxVariant      = 'participant';
    _gfxImageList    = [];
    _gfxGroupPicList = [];
    _gfxGroupPicDir  = '';
    _gfxExpanded.clear();
    // Image caches intentionally kept — data doesn't change within a session
}

/** Returns the static inner HTML planted inside the graphical editor panel. */
export function gfxPanelHTML() {
    return `
        <div id="gfx-loading" style="padding:24px;color:#888;font-size:13px;">Lade Layoutdaten \u2026</div>
        <div id="gfx-main" style="display:none;flex-direction:column;flex:1;overflow:hidden;min-height:0;">
            <div id="gfx-variant-bar" style="display:flex;gap:6px;margin-bottom:12px;flex-shrink:0;"></div>
            <div style="display:flex;gap:16px;flex:1;overflow:hidden;min-height:0;">
                <!-- Left: form controls (scrollable) -->
                <div style="flex:1;overflow-y:auto;padding-right:6px;min-width:0;display:flex;flex-direction:column;">
                    <div style="background:#f0f4ff;border-radius:8px;padding:12px;margin-bottom:12px;flex-shrink:0;">
                        <div style="font-size:11px;font-weight:700;color:#5c6bc0;text-transform:uppercase;letter-spacing:0.5px;margin-bottom:8px;">Inhaltsbereich (mm)</div>
                        <div id="gfx-area-form" style="display:grid;grid-template-columns:1fr 1fr;gap:6px 14px;"></div>
                        <div id="gfx-bg-form" style="margin-top:10px;"></div>
                    </div>
                    <div style="font-size:11px;font-weight:700;color:#5c6bc0;text-transform:uppercase;letter-spacing:0.5px;margin-bottom:8px;flex-shrink:0;">Elemente</div>
                    <div id="gfx-elements-container" style="display:flex;flex-direction:column;gap:5px;"></div>
                    <button onclick="window._gfxAddElement()"
                        style="margin-top:10px;padding:9px;border:2px dashed #b0bec5;border-radius:6px;background:none;cursor:pointer;color:#78909c;font-size:12px;transition:all 0.2s;flex-shrink:0;"
                        onmouseover="this.style.borderColor='#667eea';this.style.color='#667eea';"
                        onmouseout="this.style.borderColor='#b0bec5';this.style.color='#78909c';">
                        + Element hinzuf\u00fcgen
                    </button>
                </div>
                <!-- Right: A4 preview -->
                <div style="flex-shrink:0;display:flex;flex-direction:column;align-items:center;width:256px;">
                    <div style="font-size:11px;color:#aaa;margin-bottom:6px;text-align:center;">Vorschau (vereinfacht)</div>
                    <div id="gfx-preview" style="position:relative;width:252px;height:357px;background:#fff;border:1px solid #ccc;overflow:hidden;flex-shrink:0;box-shadow:2px 3px 10px rgba(0,0,0,0.12);"></div>
                </div>
            </div>
        </div>`;
}

/** Loads layout data and image list from the backend, then renders the editor. */
export async function loadCertLayoutEditor() {
    try {
        const [layoutResult, imgResult, grpResult] = await Promise.all([
            window.go.main.App.GetCertLayoutJSON(),
            window.go.main.App.ListBackgroundImages(),
            window.go.main.App.ListGroupPictures(),
        ]);

        if (layoutResult.status === 'error') {
            const el = document.getElementById('gfx-loading');
            if (el) el.textContent = 'Fehler: ' + layoutResult.message;
            return;
        }

        _gfxData         = layoutResult.data;
        _gfxVariant      = 'participant';
        _gfxImageList    = (imgResult.status === 'ok' && imgResult.files) ? imgResult.files : [];
        _gfxGroupPicList = (grpResult.status === 'ok' && grpResult.files) ? grpResult.files : [];
        _gfxGroupPicDir  = (grpResult.status === 'ok' && grpResult.dir)   ? grpResult.dir  : 'pictures';
        _gfxExpanded.clear();

        // Pre-fetch background images for all variants
        const bgNames = [...new Set(
            Object.keys(_GFX_VARIANTS)
                .map(v => _gfxData[v]?.background_image)
                .filter(f => f && !_gfxImgCache.has(f))
        )];

        // Pre-fetch ov_image content files for all variants' elements
        const ovNames = [...new Set(
            Object.keys(_GFX_VARIANTS).flatMap(v =>
                (_gfxData[v]?.elements || [])
                    .filter(el => el.type === 'ov_image')
                    .map(el => el.content || 'templates/ov_winner_image.png')
            ).filter(f => f && !_gfxImgCache.has(f))
        )];

        await Promise.all([
            ...bgNames.map(async f => {
                try {
                    const r = await window.go.main.App.GetImageAsBase64(f);
                    if (r.status === 'ok') _gfxImgCache.set(f, r.dataURL);
                } catch (_) { /* non-fatal */ }
            }),
            ...ovNames.map(async f => {
                try {
                    const r = await window.go.main.App.GetImageAsBase64(f);
                    if (r.status === 'ok') _gfxImgCache.set(f, r.dataURL);
                } catch (_) { /* non-fatal */ }
            }),
            // Pre-fetch first group picture as a sample for preview
            (async () => {
                if (_gfxGroupPicList.length > 0) {
                    const sample = _gfxGroupPicList[0];
                    if (!_gfxGroupCache.has(sample)) {
                        try {
                            const r = await window.go.main.App.GetGroupPictureAsBase64(sample);
                            if (r.status === 'ok') _gfxGroupCache.set(sample, r.dataURL);
                        } catch (_) { /* non-fatal */ }
                    }
                }
            })(),
        ]);

        const loading = document.getElementById('gfx-loading');
        const main    = document.getElementById('gfx-main');
        if (loading) loading.style.display = 'none';
        if (main)    main.style.display    = 'flex';
        _gfxRender();
    } catch (err) {
        const el = document.getElementById('gfx-loading');
        if (el) el.textContent = 'Fehler beim Laden: ' + err;
    }
}

// ---- rendering ----------------------------------------------------------

function _gfxCurrentVariantData() {
    return _gfxData ? _gfxData[_gfxVariant] : null;
}

function _gfxRender() {
    if (!_gfxData) return;
    _gfxRenderVariantBar();
    _gfxRenderAreaForm();
    _gfxRenderBgForm();
    _gfxRenderElements();
    _gfxRenderPreview();
}

function _gfxRenderVariantBar() {
    const bar = document.getElementById('gfx-variant-bar');
    if (!bar) return;
    bar.innerHTML = Object.entries(_GFX_VARIANTS).map(([key, label]) => {
        const active = key === _gfxVariant;
        return `<button onclick="window._gfxSelectVariant('${key}')"
            style="padding:6px 14px;border:1px solid ${active ? '#667eea' : '#ddd'};border-radius:20px;
                   background:${active ? '#667eea' : 'white'};color:${active ? 'white' : '#555'};
                   cursor:pointer;font-size:12px;font-weight:${active ? '600' : '400'};">${label}</button>`;
    }).join('');
}

function _gfxRenderAreaForm() {
    const div = document.getElementById('gfx-area-form');
    if (!div) return;
    const vd   = _gfxCurrentVariantData();
    const area = (vd && vd.content_area) || {};
    const iS   = 'padding:4px 6px;border:1px solid #ccc;border-radius:4px;font-size:12px;width:100%;box-sizing:border-box;';
    const lS   = 'display:flex;flex-direction:column;gap:2px;font-size:12px;color:#555;';
    div.innerHTML = [
        { key: 'left',   label: 'Links'  },
        { key: 'right',  label: 'Rechts' },
        { key: 'top',    label: 'Oben'   },
        { key: 'bottom', label: 'Unten'  },
    ].map(f => `<label style="${lS}">${f.label}
        <input type="number" step="0.5" style="${iS}" value="${area[f.key] ?? 0}"
            oninput="window._gfxUpdateArea('${f.key}',+this.value)">
    </label>`).join('');
}

function _gfxRenderBgForm() {
    const div = document.getElementById('gfx-bg-form');
    if (!div) return;
    const vd          = _gfxCurrentVariantData();
    const currentPath = (vd && vd.background_image) || '';
    const currentFile = _gfxBgBaseName(currentPath);
    const lS      = 'display:flex;flex-direction:column;gap:2px;font-size:12px;color:#555;';
    // Always include the currently configured file as an option even if the file
    // doesn't exist on disk yet (so the dropdown shows what is actually saved).
    const fileMissing = currentPath !== '' && !_gfxImageList.includes(currentFile);
    const allFiles = fileMissing
        ? [currentFile, ..._gfxImageList]
        : _gfxImageList;
    const borderColor = fileMissing ? '#e53935' : '#ccc';
    const sS = `padding:4px 6px;border:1px solid ${borderColor};border-radius:4px;font-size:12px;width:100%;background:white;box-sizing:border-box;`;
    const options = ['', ...allFiles].map(f => {
        const sel   = f === currentFile ? 'selected' : '';
        const label = f === '' ? '(kein Hintergrundbild)' : f;
        return `<option value="${_escHtml(f)}" ${sel}>${_escHtml(label)}</option>`;
    }).join('');
    const warning = fileMissing
        ? `<span style="color:#e53935;font-size:11px;">&#9888; Datei nicht gefunden: ${_escHtml(currentFile || currentPath)}</span>`
        : '';
    div.innerHTML = `<label style="${lS}">Hintergrundbild
        <select style="${sS}" onchange="window._gfxUpdateBgImage(this.value)">${options}</select>
        ${warning}
    </label>`;
}

function _gfxRenderElements() {
    const container = document.getElementById('gfx-elements-container');
    if (!container) return;
    const vd = _gfxCurrentVariantData();
    container.innerHTML = (vd && vd.elements)
        ? vd.elements.map((el, i) => _gfxElementCard(el, i)).join('')
        : '';
}

function _gfxElementCard(el, idx) {
    const BADGE = { text:'#4caf50', dynamic:'#667eea', members_table:'#ff9800', group_picture:'#e91e63', ov_image:'#9c27b0' };
    const badge = BADGE[el.type] || '#999';
    const label = el.type === 'text'    ? (_escHtml(el.content) || '(leer)')
                : el.type === 'dynamic' ? _escHtml(el.field) || '(kein Feld)'
                : el.type;
    const swatch = (el.color && el.color.length >= 3)
        ? `<div style="width:14px;height:14px;border-radius:3px;border:1px solid #ccc;background:rgb(${el.color[0]},${el.color[1]},${el.color[2]});flex-shrink:0;"></div>`
        : '';
    const bS  = 'padding:2px 6px;border:1px solid #ddd;border-radius:3px;background:#f5f5f5;cursor:pointer;font-size:11px;';
    const exp = _gfxExpanded.has(idx);
    return `<div style="border:1px solid #e0e0e0;border-radius:6px;overflow:hidden;background:#fff;">
        <div style="display:flex;align-items:center;padding:7px 10px;cursor:pointer;gap:7px;background:#fafafa;user-select:none;"
             onclick="window._gfxToggleExpand(${idx})">
            <span style="font-size:10px;padding:2px 5px;border-radius:3px;background:${badge};color:#fff;flex-shrink:0;font-weight:600;">${el.type}</span>
            <span style="flex:1;font-size:12px;color:#333;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${label}</span>
            <span style="font-size:11px;color:#aaa;flex-shrink:0;">y:${el.y}</span>
            ${swatch}
            <button onclick="event.stopPropagation();window._gfxMoveElement(${idx},-1)" style="${bS}" title="Nach oben">\u2191</button>
            <button onclick="event.stopPropagation();window._gfxMoveElement(${idx}, 1)" style="${bS}" title="Nach unten">\u2193</button>
            <button onclick="event.stopPropagation();window._gfxDeleteElement(${idx})" style="${bS}color:#c00;" title="L\u00f6schen">\u2715</button>
            <span style="font-size:10px;color:#bbb;">${exp ? '\u25b2' : '\u25bc'}</span>
        </div>
        ${exp ? _gfxElementForm(el, idx) : ''}
    </div>`;
}

function _gfxElementForm(el, idx) {
    const iS  = 'padding:4px 6px;border:1px solid #ccc;border-radius:4px;font-size:12px;width:100%;box-sizing:border-box;';
    const sS  = 'padding:4px 6px;border:1px solid #ccc;border-radius:4px;font-size:12px;width:100%;background:white;';
    const lS  = 'display:flex;flex-direction:column;gap:2px;font-size:11px;color:#666;';
    const s2  = 'grid-column:span 2;';

    const typeOpts  = ['dynamic','text','members_table','group_picture','ov_image']
        .map(t => `<option value="${t}" ${el.type===t?'selected':''}>${t}</option>`).join('');
    const fieldLabels = {
        event_date:     'Datum',                // aktuelles Datum
        group:          'Gruppe',
        year:           'Jahr',                 // veranstaltung.jahr
        name:           'Name',
        location_date:  'Ort & Datum',          // veranstaltung.ort + Datum
        ortsverband:    'Ortsverband',
        rank:           'Platzierung',
        winner_label:   'Siegerbezeichnung',
        event_name:     'Veranstaltungsname',   // veranstaltung.name
        event_location: 'Veranstaltungsort',    // veranstaltung.ort
    };
    const fieldOpts = Object.entries(fieldLabels)
        .map(([f, label]) => `<option value="${f}" ${el.field===f?'selected':''}>${label}</option>`).join('');

    const isText  = el.type === 'text' || el.type === 'dynamic';
    const isImage = el.type === 'group_picture' || el.type === 'ov_image';

    return `<div style="padding:12px;border-top:1px solid #eee;">
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px 12px;">
            <label style="${lS};${s2}">Typ
                <select style="${sS}" onchange="window._gfxUpdateElement(${idx},'type',this.value)">${typeOpts}</select>
            </label>
            ${el.type==='text' ? `<label style="${lS};${s2}">Inhalt
                <input type="text" style="${iS}" value="${_escHtml(el.content||'')}"
                    oninput="window._gfxUpdateElement(${idx},'content',this.value)">
            </label>` : ''}
            ${el.type==='dynamic' ? `<label style="${lS};${s2}">Feld
                <select style="${sS}" onchange="window._gfxUpdateElement(${idx},'field',this.value)">${fieldOpts}</select>
            </label>` : ''}
            <label style="${lS}">X (mm, \u22121=auto)
                <input type="number" step="0.5" style="${iS}" value="${el.x??-1}"
                    oninput="window._gfxUpdateElement(${idx},'x',+this.value)">
            </label>
            <label style="${lS}">Y (mm)
                <input type="number" step="0.5" style="${iS}" value="${el.y??0}"
                    oninput="window._gfxUpdateElement(${idx},'y',+this.value)">
            </label>
            <label style="${lS}">Breite (0=auto)
                <input type="number" step="0.5" style="${iS}" value="${el.width??0}"
                    oninput="window._gfxUpdateElement(${idx},'width',+this.value)">
            </label>
            <label style="${lS}">H\u00f6he (0=auto)
                <input type="number" step="0.5" style="${iS}" value="${el.height??0}"
                    oninput="window._gfxUpdateElement(${idx},'height',+this.value)">
            </label>
            ${isText ? `<label style="${lS};${s2}">Schriftfamilie
                <select style="${sS}" onchange="window._gfxUpdateElement(${idx},'font_family',this.value)">
                    <option value=""          ${!el.font_family?'selected':''}>Standard (Theme)</option>
                    <option value="Arial"      ${el.font_family==='Arial'?'selected':''}>Arial</option>
                    <option value="Helvetica"  ${el.font_family==='Helvetica'?'selected':''}>Helvetica</option>
                    <option value="Times"      ${el.font_family==='Times'?'selected':''}>Times</option>
                    <option value="Courier"    ${el.font_family==='Courier'?'selected':''}>Courier</option>
                </select>
            </label>
            <label style="${lS}">Schriftgrad (pt)
                <input type="number" step="0.5" min="6" max="72" style="${iS}" value="${el.font_size??12}"
                    oninput="window._gfxUpdateElement(${idx},'font_size',+this.value)">
            </label>
            <label style="${lS}">Stil
                <select style="${sS}" onchange="window._gfxUpdateElement(${idx},'font_style',this.value)">
                    <option value=""   ${!el.font_style?'selected':''}>Normal</option>
                    <option value="B"  ${el.font_style==='B'?'selected':''}>Fett</option>
                    <option value="I"  ${el.font_style==='I'?'selected':''}>Kursiv</option>
                    <option value="BI" ${el.font_style==='BI'?'selected':''}>Fett+Kursiv</option>
                </select>
            </label>
            <label style="${lS}">Ausrichtung
                <select style="${sS}" onchange="window._gfxUpdateElement(${idx},'align',this.value)">
                    <option value="C" ${el.align==='C'||!el.align?'selected':''}>Zentriert</option>
                    <option value="L" ${el.align==='L'?'selected':''}>Links</option>
                    <option value="R" ${el.align==='R'?'selected':''}>Rechts</option>
                </select>
            </label>
            <label style="${lS}">Farbe
                <div style="display:flex;align-items:center;gap:6px;">
                    <input type="color" value="${_rgbToHex(el.color)}"
                        style="width:40px;height:28px;border:1px solid #ccc;border-radius:4px;cursor:pointer;padding:1px;"
                        oninput="window._gfxUpdateElementColor(${idx},this.value)">
                    <span id="gfx-ct-${idx}" style="font-size:11px;color:#888;">${_rgbToHex(el.color)}</span>
                </div>
            </label>` : ''}
            ${isImage ? `<label style="${lS}">Bildbreite (mm)
                <input type="number" step="1" min="10" max="200" style="${iS}" value="${el.img_width??120}"
                    oninput="window._gfxUpdateElement(${idx},'img_width',+this.value)">
            </label>` : ''}            ${el.type==='ov_image' ? _gfxOvImagePicker(el, idx) : ''}
            ${el.type==='group_picture' ? _gfxGroupPicInfo(el) : ''}            ${el.type==='members_table' ? `<p style="${s2}margin:4px 0 0;font-size:11px;color:#888;font-style:italic;">Nur X, Y und Breite sind relevant. Inhalt wird automatisch bef\u00fcllt.</p>` : ''}
        </div>
    </div>`;
}

function _gfxOvImagePicker(el, idx) {
    const sS  = 'padding:4px 6px;border:1px solid #ccc;border-radius:4px;font-size:12px;width:100%;background:white;box-sizing:border-box;';
    const lS  = 'display:flex;flex-direction:column;gap:2px;font-size:11px;color:#666;';
    const s2  = 'grid-column:span 2;';
    const current = el.content || 'templates/ov_winner_image.png';
    const fileMissing = !_gfxImgCache.has(current);
    const borderColor = fileMissing ? '#e53935' : '#ccc';
    const sSc = sS.replace('border:1px solid #ccc', `border:1px solid ${borderColor}`);
    // Include current file even if it doesn't appear in imageList
    const allFiles = _gfxImageList.includes(current) ? _gfxImageList : [current, ..._gfxImageList];
    const options = allFiles.map(f => {
        const sel = f === current ? 'selected' : '';
        return `<option value="${_escHtml(f)}" ${sel}>${_escHtml(f)}</option>`;
    }).join('');
    const warning = fileMissing
        ? `<span style="color:#e53935;font-size:11px;">&#9888; Datei nicht gefunden: ${_escHtml(current)}</span>`
        : '';
    return `<label style="${lS};${s2}">Bilddatei
        <select style="${sSc}" onchange="window._gfxUpdateElementImage(${idx},this.value)">${options}</select>
        ${warning}
    </label>`;
}

function _gfxGroupPicInfo(el) {
    const lS = 'display:flex;flex-direction:column;gap:2px;font-size:11px;color:#666;';
    const s2 = 'grid-column:span 2;';
    const sample = _gfxGroupPicList.length > 0
        ? `Beispiel: <code>${_escHtml(_gfxGroupPicDir)}/${_escHtml(_gfxGroupPicList[0])}</code>`
        : `(noch keine Fotos in <code>${_escHtml(_gfxGroupPicDir)}/</code>)`;
    return `<div style="${lS};${s2};background:#fff8e1;border-radius:4px;padding:6px 8px;border-left:3px solid #ffb300;">
        <span style="font-weight:600;color:#e65100;">Dynamisches Foto</span>
        <span>Muster: <code>${_escHtml(_gfxGroupPicDir)}/group_picture_NNN.jpg</code></span>
        <span>${sample}</span>
    </div>`;
}

function _gfxRenderPreview() {
    const preview = document.getElementById('gfx-preview');
    if (!preview) return;
    const vd = _gfxCurrentVariantData();
    if (!vd) { preview.innerHTML = ''; return; }

    const S    = 252 / 210;   // px per mm  (A4 = 210 mm wide, preview = 252 px wide)
    const area = vd.content_area || { left: 15, top: 20, right: 195, bottom: 277 };

    let html = '';

    // Background image
    const bgDataURL = vd.background_image ? _gfxImgCache.get(vd.background_image) : null;
    if (bgDataURL) {
        html += `<img src="${bgDataURL}" style="position:absolute;left:0;top:0;width:100%;height:100%;object-fit:fill;pointer-events:none;">`;
    }

    // Content-area highlight
    const aL = area.left  * S, aT = area.top * S;
    const aW = (area.right - area.left) * S, aH = (area.bottom - area.top) * S;
    html += `<div style="position:absolute;left:${aL}px;top:${aT}px;width:${aW}px;height:${aH}px;
                background:rgba(102,126,234,0.06);border:1px dashed rgba(102,126,234,0.35);
                box-sizing:border-box;pointer-events:none;"></div>`;

    for (const el of (vd.elements || [])) {
        const xEff = (el.x === undefined || el.x < 0) ? area.left           : el.x;
        const wEff = (!el.width || el.width <= 0)     ? (area.right - area.left) : el.width;
        const px = xEff * S, py = el.y * S, pw = wEff * S;

        if (el.type === 'text' || el.type === 'dynamic') {
            const text  = el.type === 'text' ? _escHtml(el.content || '') : `{${_escHtml(el.field || '?')}}`;
            const fw    = (el.font_style === 'B' || el.font_style === 'BI') ? 'bold'   : 'normal';
            const fi    = (el.font_style === 'I' || el.font_style === 'BI') ? 'italic' : 'normal';
            const fsPx  = Math.max(7, (el.font_size || 12) * S * 0.3528);
            const color = (el.color && el.color.length >= 3) ? `rgb(${el.color[0]},${el.color[1]},${el.color[2]})` : '#000';
            const ta    = el.align === 'L' ? 'left' : el.align === 'R' ? 'right' : 'center';
            const ff    = el.font_family ? `'${el.font_family}', sans-serif` : 'Arial, sans-serif';
            html += `<div style="position:absolute;left:${px}px;top:${py}px;width:${pw}px;
                        font-size:${fsPx}px;font-weight:${fw};font-style:${fi};font-family:${ff};color:${color};
                        text-align:${ta};overflow:hidden;white-space:nowrap;line-height:1.2;pointer-events:none;">${text}</div>`;

        } else if (el.type === 'members_table') {
            html += `<div style="position:absolute;left:${px}px;top:${py}px;width:${pw}px;height:${22*S}px;
                        background:rgba(0,0,0,0.04);border:1px dashed #ccc;font-size:8px;color:#999;
                        display:flex;align-items:center;justify-content:center;pointer-events:none;">Mitgliederliste</div>`;

        } else if (el.type === 'group_picture' || el.type === 'ov_image') {
            const imgW  = (el.img_width || 120) * S;
            const imgX  = px + (pw - imgW) / 2;
            const imgH  = 22 * S;   // estimated placeholder height

            if (el.type === 'ov_image') {
                // Static image — show the actual file if available
                const ovFile = el.content || 'templates/ov_winner_image.png';
                const dataURL = _gfxImgCache.get(ovFile);
                if (dataURL) {
                    html += `<img src="${dataURL}" style="position:absolute;left:${imgX}px;top:${py}px;width:${imgW}px;object-fit:contain;pointer-events:none;">`;
                } else {
                    html += `<div style="position:absolute;left:${imgX}px;top:${py}px;width:${imgW}px;height:${imgH}px;
                        background:#e8eaf6;border:1px solid #9fa8da;font-size:8px;color:#5c6bc0;
                        display:flex;align-items:center;justify-content:center;pointer-events:none;">\ud83c\udfc6 ${_escHtml(ovFile)}</div>`;
                }
            } else {
                // group_picture — dynamic per group, show a sample if available
                const sampleName = _gfxGroupPicList[0];
                const sampleURL  = sampleName ? _gfxGroupCache.get(sampleName) : null;
                if (sampleURL) {
                    const badge = `<div style="position:absolute;left:${imgX}px;top:${py}px;padding:2px 4px;font-size:7px;background:rgba(0,0,0,0.45);color:#fff;border-radius:2px;pointer-events:none;">Beispielfoto</div>`;
                    html += `<img src="${sampleURL}" style="position:absolute;left:${imgX}px;top:${py}px;width:${imgW}px;object-fit:contain;pointer-events:none;opacity:0.85;">` + badge;
                } else {
                    html += `<div style="position:absolute;left:${imgX}px;top:${py}px;width:${imgW}px;height:${imgH}px;
                        background:#fce4ec;border:1px dashed #e91e63;font-size:8px;color:#e91e63;
                        display:flex;align-items:center;justify-content:center;pointer-events:none;">\ud83d\udcf7 Gruppenfoto</div>`;
                }
            }
        }
    }
    preview.innerHTML = html;
}

// ---- window-scoped event handlers (used by inline onclick in generated HTML) ----

window._gfxSelectVariant = function (variant) {
    _gfxVariant = variant;
    _gfxExpanded.clear();
    _gfxRender();
};

window._gfxUpdateArea = function (field, value) {
    const vd = _gfxCurrentVariantData();
    if (!vd) return;
    if (!vd.content_area) vd.content_area = {};
    vd.content_area[field] = value;
    _gfxRenderPreview();
};

window._gfxUpdateBgImage = async function (filename) {
    const vd = _gfxCurrentVariantData();
    if (!vd) return;
    const storedPath = _gfxBgStorePath(filename);
    vd.background_image = storedPath;
    if (storedPath && !_gfxImgCache.has(storedPath)) {
        try {
            const r = await window.go.main.App.GetImageAsBase64(storedPath);
            if (r.status === 'ok') _gfxImgCache.set(storedPath, r.dataURL);
        } catch (_) { /* non-fatal */ }
    }
    _gfxRenderPreview();
};

window._gfxUpdateElementImage = async function (idx, filename) {
    const vd = _gfxCurrentVariantData();
    if (!vd || !vd.elements || !vd.elements[idx]) return;
    vd.elements[idx].content = filename;
    if (filename && !_gfxImgCache.has(filename)) {
        try {
            const r = await window.go.main.App.GetImageAsBase64(filename);
            if (r.status === 'ok') _gfxImgCache.set(filename, r.dataURL);
        } catch (_) { /* non-fatal */ }
    }
    _gfxRenderElements();
    _gfxRenderPreview();
};

window._gfxUpdateElement = function (idx, field, value) {
    const vd = _gfxCurrentVariantData();
    if (!vd || !vd.elements || !vd.elements[idx]) return;
    vd.elements[idx][field] = value;
    _gfxRenderPreview();
};

window._gfxUpdateElementColor = function (idx, hex) {
    const vd = _gfxCurrentVariantData();
    if (!vd || !vd.elements || !vd.elements[idx]) return;
    vd.elements[idx].color = _hexToRgb(hex);
    const span = document.getElementById('gfx-ct-' + idx);
    if (span) span.textContent = hex;
    _gfxRenderPreview();
};

window._gfxToggleExpand = function (idx) {
    _gfxExpanded.has(idx) ? _gfxExpanded.delete(idx) : _gfxExpanded.add(idx);
    _gfxRenderElements();
    _gfxRenderPreview();
};

window._gfxMoveElement = function (idx, dir) {
    const vd = _gfxCurrentVariantData();
    if (!vd || !vd.elements) return;
    const newIdx = idx + dir;
    if (newIdx < 0 || newIdx >= vd.elements.length) return;
    [vd.elements[idx], vd.elements[newIdx]] = [vd.elements[newIdx], vd.elements[idx]];
    const wasA = _gfxExpanded.has(idx), wasB = _gfxExpanded.has(newIdx);
    _gfxExpanded.delete(idx); _gfxExpanded.delete(newIdx);
    if (wasA) _gfxExpanded.add(newIdx);
    if (wasB) _gfxExpanded.add(idx);
    _gfxRenderElements();
    _gfxRenderPreview();
};

window._gfxDeleteElement = function (idx) {
    const vd = _gfxCurrentVariantData();
    if (!vd || !vd.elements) return;
    vd.elements.splice(idx, 1);
    _gfxExpanded.clear();
    _gfxRenderElements();
    _gfxRenderPreview();
};

window._gfxAddElement = function () {
    const vd = _gfxCurrentVariantData();
    if (!vd) return;
    if (!vd.elements) vd.elements = [];
    vd.elements.push({
        type: 'text', content: 'Neues Element', field: 'event_name',
        x: -1, y: 100, width: 0, height: 10,
        font_family: '', font_style: '', font_size: 14, align: 'C', color: [0, 0, 0],
        img_width: 0, space_before: 0,
    });
    const newIdx = vd.elements.length - 1;
    _gfxExpanded.add(newIdx);
    _gfxRenderElements();
    _gfxRenderPreview();
    setTimeout(() => {
        const cards = document.querySelectorAll('#gfx-elements-container > div');
        cards[newIdx]?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }, 50);
};
