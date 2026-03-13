// PDF generation handlers
import { setStatus } from '../shared/dom.js';

export async function handleGeneratePDF() {
    setStatus('Generating groups report PDF...', 'info');
    
    try {
        const result = await window.go.main.App.GeneratePDF();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
        } else {
            setStatus('✅ Groups report PDF generated successfully!', 'success');
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
    }
}

export async function handleGenerateGroupEvaluationPDF() {
    setStatus('Generating group evaluation PDF...', 'info');
    
    try {
        const result = await window.go.main.App.GenerateGroupEvaluationPDF();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
        } else {
            setStatus('✅ Group evaluation PDF generated successfully!', 'success');
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
    }
}

export async function handleGenerateOrtsverbandEvaluationPDF() {
    setStatus('Generating Ortsverband evaluation PDF...', 'info');
    
    try {
        const result = await window.go.main.App.GenerateOrtsverbandEvaluationPDF();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
        } else {
            setStatus('✅ Ortsverband evaluation PDF generated successfully!', 'success');
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
    }
}

export async function handleGenerateCertificates() {
    setStatus('Generating certificates...', 'info');
    
    try {
        const result = await window.go.main.App.GenerateParticipantCertificates();
        
        if (result.status === 'error') {
            setStatus('ERROR: ' + result.message, 'error');
        } else {
            setStatus('✅ Certificates generated successfully!', 'success');
        }
    } catch (err) {
        setStatus('ERROR: ' + err, 'error');
    }
}
