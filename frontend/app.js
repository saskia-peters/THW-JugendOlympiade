// Main application orchestrator - imports and wires up all modules
import { openFileDialog, handleBackupDatabase, handleRestoreDatabase, handleDistributeGroups } from './admin/file-handler.js';
import { handleEditConfig } from './admin/config-editor.js';
import { handleShowGroups } from './groups/groups.js';
import { handleShowStations, handleShowStationsForGroup } from './stations/stations.js';
import { handleGroupEvaluation, handleOrtsverbandEvaluation } from './evaluations/evaluations.js';
import { 
    handleGeneratePDF, 
    handleGenerateGroupEvaluationPDF, 
    handleGenerateOrtsverbandEvaluationPDF, 
    handleGenerateCertificates,
    handleGenerateOVCertificates
} from './reports/pdf-handlers.js';

// Load configuration from backend and store globally for use by all modules
(async () => {
    try {
        window.appConfig = await window.go.main.App.GetConfig();
    } catch (e) {
        window.appConfig = { scoreMin: 100, scoreMax: 1200, maxGroupSize: 8 };
    }
})();

// Expose functions to window object for onclick handlers
window.openFileDialog = openFileDialog;
window.handleBackupDatabase = handleBackupDatabase;
window.handleRestoreDatabase = handleRestoreDatabase;
window.handleDistributeGroups = handleDistributeGroups;
window.handleEditConfig = handleEditConfig;
window.handleShowGroups = handleShowGroups;
window.handleShowStations = handleShowStations;
window.handleShowStationsForGroup = handleShowStationsForGroup;
window.handleEvaluation = handleGroupEvaluation;
window.handleOrtsverbandEvaluation = handleOrtsverbandEvaluation;
window.handleGeneratePDF = handleGeneratePDF;
window.handleGenerateGroupEvaluationPDF = handleGenerateGroupEvaluationPDF;
window.handleGenerateOrtsverbandEvaluationPDF = handleGenerateOrtsverbandEvaluationPDF;
window.handleGenerateCertificates = handleGenerateCertificates;
window.handleGenerateOVCertificates = handleGenerateOVCertificates;
