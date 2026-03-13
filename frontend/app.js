// Main application orchestrator - imports and wires up all modules
import { openFileDialog, handleBackupDatabase, handleRestoreDatabase } from './admin/file-handler.js';
import { handleShowGroups } from './groups/groups.js';
import { handleShowStations, handleShowStationsForGroup } from './stations/stations.js';
import { handleGroupEvaluation, handleOrtsverbandEvaluation } from './evaluations/evaluations.js';
import { 
    handleGeneratePDF, 
    handleGenerateGroupEvaluationPDF, 
    handleGenerateOrtsverbandEvaluationPDF, 
    handleGenerateCertificates 
} from './reports/pdf-handlers.js';

// Expose functions to window object for onclick handlers
window.openFileDialog = openFileDialog;
window.handleBackupDatabase = handleBackupDatabase;
window.handleRestoreDatabase = handleRestoreDatabase;
window.handleShowGroups = handleShowGroups;
window.handleShowStations = handleShowStations;
window.handleShowStationsForGroup = handleShowStationsForGroup;
window.handleEvaluation = handleGroupEvaluation;
window.handleOrtsverbandEvaluation = handleOrtsverbandEvaluation;
window.handleGeneratePDF = handleGeneratePDF;
window.handleGenerateGroupEvaluationPDF = handleGenerateGroupEvaluationPDF;
window.handleGenerateOrtsverbandEvaluationPDF = handleGenerateOrtsverbandEvaluationPDF;
window.handleGenerateCertificates = handleGenerateCertificates;
