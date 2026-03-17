# Frontend Modularization

The frontend has been restructured from a monolithic 1000+ line `app.js` into logical ES6 modules for improved maintainability and testability.

## Module Structure

### Shared Modules (`shared/`)

1. **shared/dom.js**
   - **Purpose**: DOM element references and basic UI utilities
   - **Exports**: 
     - DOM elements: `status`, `output`, `tabs`, `tabButtons`, `tabContents`, all 5 buttons
     - Functions: `setStatus()`, `clearAllTabs()`
   - **Dependencies**: None

2. **shared/utils.js**
   - **Purpose**: Shared utility functions
   - **Exports**: 
     - `escapeHtml()` - XSS prevention for user-generated content
     - `switchTab()` - Tab switching logic
   - **Dependencies**: None

### Feature Modules

3. **admin/file-handler.js**
   - **Purpose**: File loading, database backup and restore
   - **Exports**: `openFileDialog()`, `handleBackupDatabase()`, `handleRestoreDatabase()`
   - **Functionality**: 
     - CheckDB confirmation dialog before overwriting
     - LoadFile call to backend with button enablement
     - Backup creation with status feedback
     - Restore with backup-selection dialog
   - **Dependencies**: shared/dom.js

4. **groups/groups.js**
   - **Purpose**: Group display with tabs and statistics
   - **Exports**: `handleShowGroups()`
   - **Internal**: `renderGroupTabs()`, `formatGroupContent()`
   - **UI Elements**: 
     - Participant table (Name, Ortsverband, Alter, Geschlecht)
     - Statistics panel (total, avg age, ortsverband distribution, gender distribution)
     - Per-group link to Ergebniseingabe
   - **Dependencies**: shared/dom.js, shared/utils.js

5. **stations/scores.js**
   - **Purpose**: Legacy per-station score assignment helpers
   - **Exports**: 
     - `checkForExistingScore()` - Duplicate score detection
     - `handleGlobalAssignScore()` - Global score form handler
     - `handleAssignScore()` - Per-station score handler
   - **Dependencies**: shared/dom.js, dynamic import of stations/stations.js

6. **stations/stations.js**
   - **Purpose**: Group-based results entry with dirty-tracking
   - **Exports**: 
     - `handleShowStations()` - Entry point, loads stations + groups
     - `handleShowStationsForGroup(groupID)` - Jump to specific group
   - **Internal**: `renderGroupBasedEntry()`, `renderStationTable()`, `doSaveAll()`, `hasDirtyScores()`, `showUnsavedWarning()`
   - **UI Elements**: 
     - Group selector dropdown
     - Station scores table (one row per station, input + save button)
     - "Alle Ergebnisse speichern" button
     - Unsaved-changes modal when switching groups
   - **State**: Module-level `savedScoreMap`, `currentGroupID`, `pendingGroupID`
   - **Dependencies**: shared/dom.js, shared/utils.js

7. **evaluations/evaluations.js**
   - **Purpose**: Evaluation rendering for groups and ortsverbände
   - **Exports**: 
     - `handleGroupEvaluation()` - Display group rankings
     - `handleOrtsverbandEvaluation()` - Display ortsverband rankings
   - **Internal**: `renderGroupEvaluations()`, `renderOrtsverbandEvaluations()`
   - **UI Elements**: 
     - Rankings table with 🥇🥈🥉 medals
     - Statistics panels (total, averages, highest/lowest)
     - "PDF erstellen" button
   - **Dependencies**: shared/dom.js, shared/utils.js

8. **reports/pdf-handlers.js**
   - **Purpose**: PDF generation wrappers
   - **Exports**: 
     - `handleGeneratePDF()` - Groups report PDF
     - `handleGenerateGroupEvaluationPDF()` - Group evaluation PDF
     - `handleGenerateOrtsverbandEvaluationPDF()` - Ortsverband evaluation PDF
     - `handleGenerateCertificates()` - Participant certificates
   - **Dependencies**: shared/dom.js

9. **app.js**
   - **Purpose**: Main orchestrator - imports all modules and wires up onclick handlers
   - **Functionality**: 
     - Imports all feature modules
     - Exposes functions to `window` object for HTML onclick attributes
   - **Dependencies**: All feature modules

## Architecture Benefits

### Before (Monolithic app.js)
- ❌ 1000+ lines in single file
- ❌ Difficult to navigate and maintain
- ❌ No clear separation of concerns
- ❌ Hard to test individual features
- ❌ Risk of naming conflicts
- ❌ Difficult to reuse code

### After (Modular ES6 Structure)
- ✅ 9 focused modules averaging ~110 lines each
- ✅ Clear separation of concerns (UI, data, logic)
- ✅ Easy to navigate: dom → utils → features → orchestrator
- ✅ Testable: Each module can be imported and tested independently
- ✅ Clear dependencies via import statements
- ✅ Reusable: Modules can be imported by other modules
- ✅ Type-safe: Can add JSDoc types to exports
- ✅ Better code organization: Related code grouped logically

## Dependency Graph

```
app.js (orchestrator)
├── admin/file-handler.js
│   └── shared/dom.js
├── groups/groups.js
│   ├── shared/dom.js
│   └── shared/utils.js
├── stations/scores.js
│   ├── shared/dom.js
│   └── stations/stations.js (dynamic import)
├── stations/stations.js
│   ├── shared/dom.js
│   └── shared/utils.js
├── evaluations/evaluations.js
│   ├── shared/dom.js
│   └── shared/utils.js
└── reports/pdf-handlers.js
    └── shared/dom.js
```

## Integration with HTML

The `index.html` file loads the app as an ES6 module:

```html
<script type="module" src="app.js"></script>
```

Functions are exposed to the global `window` object to support onclick handlers:

```javascript
window.openFileDialog = openFileDialog;
window.handleShowGroups = handleShowGroups;
// ... etc
```

HTML onclick attributes work as before:

```html
<button onclick="openFileDialog()">Lade Excel Datei</button>
<button onclick="handleShowGroups()">Gruppen</button>
```

## File Sizes

| Module | Purpose |
|--------|----------|
| shared/dom.js | DOM references & UI utilities |
| shared/utils.js | Helper functions |
| admin/file-handler.js | File loading, backup, restore |
| groups/groups.js | Group display |
| stations/scores.js | Legacy per-station score helpers |
| stations/stations.js | Group-based results entry |
| evaluations/evaluations.js | Rankings & evaluations |
| reports/pdf-handlers.js | PDF generation |
| app.js | Main orchestrator |

## Testing Strategy

Each module can now be tested independently:

```javascript
// Example: Testing groups.js
import { handleShowGroups } from './groups.js';

// Mock dependencies
jest.mock('./dom.js');

// Test group rendering logic
test('handleShowGroups displays groups correctly', async () => {
    // ...
});
```

## Future Improvements

1. **TypeScript**: Add `.d.ts` type definition files for type safety
2. **Unit Tests**: Create Jest/Vitest tests for each module
3. **Code Splitting**: Lazy load evaluation/PDF modules on demand
4. **State Management**: Add centralized state management if complexity grows
5. **JSDoc**: Add comprehensive JSDoc comments for better IDE support

## Migration Notes

- Original `app.js` backed up as `app.old.js`
- No changes to backend Go code
- No changes to HTML structure (except module script tag)
- All functionality preserved - only internal organization changed
- Build process unchanged (Wails build works identically)

## Verification

✅ Build successful: `wails build` completed without errors
✅ Module imports: All ES6 imports resolve correctly
✅ Function exposure: Window object has all required functions
✅ Dependency graph: No circular dependencies (except dynamic import in scores.js)
