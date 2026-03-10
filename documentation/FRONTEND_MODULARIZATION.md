# Frontend Modularization

The frontend has been restructured from a monolithic 1000+ line `app.js` into logical ES6 modules for improved maintainability and testability.

## Module Structure

### Core Modules

1. **dom.js** (32 lines)
   - **Purpose**: DOM element references and basic UI utilities
   - **Exports**: 
     - DOM elements: `status`, `output`, `tabs`, `tabButtons`, `tabContents`, all 5 buttons
     - Functions: `setStatus()`, `clearAllTabs()`
   - **Dependencies**: None

2. **utils.js** (21 lines)
   - **Purpose**: Shared utility functions
   - **Exports**: 
     - `escapeHtml()` - XSS prevention for user-generated content
     - `switchTab()` - Tab switching logic
   - **Dependencies**: None

### Feature Modules

3. **file-handler.js** (42 lines)
   - **Purpose**: File upload and database loading
   - **Exports**: `openFileDialog()`
   - **Functionality**: 
     - CheckDB confirmation dialog
     - LoadFile call to backend
     - Button enablement on success
   - **Dependencies**: dom.js

4. **groups.js** (138 lines)
   - **Purpose**: Group display with tabs and statistics
   - **Exports**: `handleShowGroups()`
   - **Internal**: `renderGroupTabs()`, `formatGroupContent()`
   - **UI Elements**: 
     - Participant table (Name, Ortsverband, Alter, Geschlecht)
     - Statistics panel (total, avg age, ortsverband distribution, gender distribution)
   - **Dependencies**: dom.js, utils.js

5. **scores.js** (204 lines)
   - **Purpose**: Score assignment with validation and warnings
   - **Exports**: 
     - `checkForExistingScore()` - Duplicate score detection
     - `handleGlobalAssignScore()` - Global score form handler
     - `handleAssignScore()` - Per-station score handler
   - **Features**: 
     - Existing score detection
     - Overwrite confirmation dialogs
     - Form clearing and reset
     - Optimistic UI updates with refresh
   - **Dependencies**: dom.js, dynamic import of stations.js (to avoid circular dependency)

6. **stations.js** (243 lines)
   - **Purpose**: Station display and management
   - **Exports**: 
     - `handleShowStations()` - Load and display all stations
     - `showStationDetails()` - Show specific station details
   - **Internal**: `renderStationTabs()`, `formatStationContent()`
   - **UI Elements**: 
     - 4-column grid of station buttons
     - Global score entry form (above buttons)
     - Station detail view with scores table
     - Statistics panel (total groups, avg score, highest/lowest)
   - **Dependencies**: dom.js, utils.js, scores.js

7. **evaluations.js** (186 lines)
   - **Purpose**: Evaluation rendering for groups and ortsverbände
   - **Exports**: 
     - `handleGroupEvaluation()` - Display group rankings
     - `handleOrtsverbandEvaluation()` - Display ortsverband rankings
   - **Internal**: `renderGroupEvaluations()`, `renderOrtsverbandEvaluations()`
   - **UI Elements**: 
     - Rankings table with 🥇🥈🥉 medals
     - Statistics panels (total, averages, highest/lowest)
   - **Dependencies**: dom.js, utils.js

8. **pdf-handlers.js** (67 lines)
   - **Purpose**: PDF generation wrappers
   - **Exports**: 
     - `handleGeneratePDF()` - Groups report PDF
     - `handleGenerateGroupEvaluationPDF()` - Group evaluation PDF
     - `handleGenerateOrtsverbandEvaluationPDF()` - Ortsverband evaluation PDF
     - `handleGenerateCertificates()` - Participant certificates
   - **Dependencies**: dom.js

9. **app.js** (27 lines)
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
├── file-handler.js
│   └── dom.js
├── groups.js
│   ├── dom.js
│   └── utils.js
├── scores.js
│   ├── dom.js
│   └── stations.js (dynamic import)
├── stations.js
│   ├── dom.js
│   ├── utils.js
│   └── scores.js
├── evaluations.js
│   ├── dom.js
│   └── utils.js
└── pdf-handlers.js
    └── dom.js
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
<button onclick="openFileDialog()">Load Excel File</button>
<button onclick="handleShowGroups()">Gruppen</button>
```

## File Sizes

| Module | Lines | Purpose |
|--------|-------|---------|
| dom.js | 32 | DOM references & UI utilities |
| utils.js | 21 | Helper functions |
| file-handler.js | 42 | File loading |
| groups.js | 138 | Group display |
| scores.js | 204 | Score management |
| stations.js | 243 | Station display |
| evaluations.js | 186 | Rankings & evaluations |
| pdf-handlers.js | 67 | PDF generation |
| app.js | 27 | Main orchestrator |
| **Total** | **960** | **9 modules** |

Original monolithic `app.js`: ~1000 lines → Now split into 9 focused modules

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
