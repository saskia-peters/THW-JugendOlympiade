// PDF generation handlers
import { setStatus } from '../shared/dom.js';

export async function handleGeneratePDF() {
    setStatus('Gruppen-PDF wird erstellt...', 'info');
    
    try {
        const result = await window.go.main.App.GeneratePDF();
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
        } else {
            setStatus('✅ Gruppen-PDF erfolgreich erstellt!', 'success');
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

export async function handleGenerateGroupEvaluationPDF() {
    setStatus('Auswertungs-PDF wird erstellt...', 'info');
    
    try {
        const result = await window.go.main.App.GenerateGroupEvaluationPDF();
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
        } else {
            setStatus('✅ Auswertungs-PDF erfolgreich erstellt!', 'success');
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

export async function handleGenerateOrtsverbandEvaluationPDF() {
    setStatus('Ortsverband-Auswertungs-PDF wird erstellt...', 'info');
    
    try {
        const result = await window.go.main.App.GenerateOrtsverbandEvaluationPDF();
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
        } else {
            setStatus('✅ Ortsverband-Auswertungs-PDF erfolgreich erstellt!', 'success');
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

export async function handleGenerateCertificates() {
    setStatus('Urkunden Teilnehmende werden erstellt...', 'info');
    
    try {
        const result = await window.go.main.App.GenerateParticipantCertificates();
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
        } else {
            setStatus('✅ ' + result.message, 'success');
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}

export async function handleGenerateOVCertificates() {
    setStatus('Urkunden Ortsverbände werden erstellt...', 'info');
    
    try {
        const result = await window.go.main.App.GenerateOrtsverbandCertificates();
        
        if (result.status === 'error') {
            setStatus('FEHLER: ' + result.message, 'error');
        } else {
            setStatus('✅ ' + result.message, 'success');
        }
    } catch (err) {
        setStatus('FEHLER: ' + err, 'error');
    }
}
