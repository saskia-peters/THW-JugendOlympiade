// DOM element references
export const status = document.getElementById('status');
export const output = document.getElementById('output');
export const tabs = document.getElementById('tabs');
export const tabButtons = document.getElementById('tabButtons');
export const tabContents = document.getElementById('tabContents');
export const btnShow = document.getElementById('btnShow');
export const btnDistribute = document.getElementById('btnDistribute');
export const btnStations = document.getElementById('btnStations');
export const btnEvaluation = document.getElementById('btnEvaluation');
export const btnOrtsverband = document.getElementById('btnOrtsverband');
export const btnPDF = document.getElementById('btnPDF');
export const btnCertificates = document.getElementById('btnCertificates');
export const btnOVCertificates = document.getElementById('btnOVCertificates');

// Status message handler
export function setStatus(msg, type = 'info') {
    status.textContent = msg;
    status.className = 'status ' + type;
}

// Clear all tab content
export function clearAllTabs() {
    // Completely clear tab containers
    tabButtons.innerHTML = '';
    tabContents.innerHTML = '';
    // Force a reflow to ensure cleanup
    void tabButtons.offsetHeight;
    void tabContents.offsetHeight;
}
