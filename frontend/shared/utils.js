// Utility functions

// Escape HTML to prevent XSS
export function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Switch between tabs
export function switchTab(tabIndex, tabButtons, tabContents) {
    const buttons = tabButtons.querySelectorAll('.tab-button');
    const contents = tabContents.querySelectorAll('.tab-content');
    
    buttons.forEach((btn, i) => {
        btn.classList.toggle('active', i === tabIndex);
    });
    
    contents.forEach((content, i) => {
        content.classList.toggle('active', i === tabIndex);
    });
}
