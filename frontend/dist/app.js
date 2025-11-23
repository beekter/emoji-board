// Main application logic
let emojis = [];
let selectedIndex = -1;
const COLUMNS = 5;

const searchInput = document.getElementById('search');
const emojiGrid = document.getElementById('emoji-grid');

// Initialize app
async function init() {
    // Load initial emojis
    await updateEmojis('');
    
    // Focus search input
    searchInput.focus();
    
    // Set up event listeners
    searchInput.addEventListener('input', handleSearch);
    searchInput.addEventListener('keydown', handleSearchKeydown);
    emojiGrid.addEventListener('keydown', handleGridKeydown);
    
    // Global keyboard handler to always capture typing into search
    document.addEventListener('keydown', handleGlobalKeydown);
    
    // Prevent context menu
    document.addEventListener('contextmenu', (e) => e.preventDefault());
}

// Global keyboard handler to ensure search input always receives text input
function handleGlobalKeydown(e) {
    // If the search input already has focus, let it handle all keys normally (including Backspace/Delete)
    if (document.activeElement === searchInput) {
        return;
    }
    
    // List of non-printable keys to ignore when not in search
    const specialKeys = [
        'Escape', 'Enter', 'Tab', 'Backspace', 'Delete',
        'ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight',
        'Home', 'End', 'PageUp', 'PageDown',
        'F1', 'F2', 'F3', 'F4', 'F5', 'F6', 'F7', 'F8', 'F9', 'F10', 'F11', 'F12',
        'Shift', 'Control', 'Alt', 'Meta', 'CapsLock', 'NumLock', 'ScrollLock'
    ];
    
    // Check if this is a printable character
    const isPrintable = e.key.length === 1 && 
                       !e.ctrlKey && 
                       !e.altKey && 
                       !e.metaKey &&
                       !specialKeys.includes(e.key);
    
    // If it's a printable character, focus search and let it handle the input
    if (isPrintable) {
        searchInput.focus();
        // The keydown will be processed by the search input automatically
    }
}

// Update emojis based on search
async function updateEmojis(query) {
    try {
        const results = await window.go.main.App.SearchEmojis(query, 100);
        emojis = results || [];
        renderEmojis();
    } catch (err) {
        console.error('Failed to search emojis:', err);
    }
}

// Render emoji grid
function renderEmojis() {
    emojiGrid.innerHTML = '';
    
    emojis.forEach((emoji, index) => {
        const item = document.createElement('div');
        item.className = 'emoji-item';
        item.textContent = emoji.emoji;
        item.dataset.index = index;
        
        // Click handler
        item.addEventListener('click', () => selectEmoji(index));
        
        emojiGrid.appendChild(item);
    });
    
    // Clear selection when emojis change
    selectedIndex = -1;
    updateSelection();
}

// Handle search input
async function handleSearch(e) {
    await updateEmojis(e.target.value);
}

// Handle search input keydown
function handleSearchKeydown(e) {
    if (e.key === 'Escape') {
        window.runtime.Quit();
    } else if (e.key === 'ArrowDown' || e.key === 'Enter') {
        e.preventDefault();
        if (emojis.length > 0) {
            selectedIndex = 0;
            emojiGrid.focus();
            updateSelection();
        }
    }
}

// Handle grid keydown
function handleGridKeydown(e) {
    if (e.key === 'Escape') {
        // Return to search
        scrollToTop();
        searchInput.focus();
        selectedIndex = -1;
        updateSelection();
        return;
    }
    
    if (emojis.length === 0) return;
    
    // Initialize selection if needed
    if (selectedIndex === -1) {
        selectedIndex = 0;
    }
    
    const oldIndex = selectedIndex;
    
    switch (e.key) {
        case 'ArrowDown':
            e.preventDefault();
            if (selectedIndex + COLUMNS < emojis.length) {
                selectedIndex += COLUMNS;
            }
            break;
        case 'ArrowUp':
            e.preventDefault();
            if (selectedIndex >= COLUMNS) {
                selectedIndex -= COLUMNS;
            } else {
                // Return to search
                scrollToTop();
                searchInput.focus();
                selectedIndex = -1;
                updateSelection();
                return;
            }
            break;
        case 'ArrowLeft':
            e.preventDefault();
            if (selectedIndex > 0) {
                selectedIndex--;
            }
            break;
        case 'ArrowRight':
            e.preventDefault();
            if (selectedIndex < emojis.length - 1) {
                selectedIndex++;
            }
            break;
        case 'Enter':
        case ' ':
            e.preventDefault();
            if (selectedIndex >= 0 && selectedIndex < emojis.length) {
                selectEmoji(selectedIndex);
            }
            return;
    }
    
    if (oldIndex !== selectedIndex) {
        updateSelection();
        scrollToSelected();
    }
}

// Update visual selection
function updateSelection() {
    const items = emojiGrid.querySelectorAll('.emoji-item');
    items.forEach((item, index) => {
        if (index === selectedIndex) {
            item.classList.add('selected');
        } else {
            item.classList.remove('selected');
        }
    });
}

// Scroll to selected emoji
function scrollToSelected() {
    if (selectedIndex < 0 || selectedIndex >= emojis.length) return;
    
    const items = emojiGrid.querySelectorAll('.emoji-item');
    if (items.length === 0) return;
    
    const row = Math.floor(selectedIndex / COLUMNS);
    
    // Get actual cell height from the first item
    const cellHeight = items[0].offsetHeight;
    const targetY = row * cellHeight;
    
    const gridRect = emojiGrid.getBoundingClientRect();
    const scrollTop = emojiGrid.scrollTop;
    const visibleTop = scrollTop;
    const visibleBottom = scrollTop + gridRect.height;
    
    const emojiTop = targetY;
    const emojiBottom = targetY + cellHeight;
    
    if (emojiTop < visibleTop) {
        emojiGrid.scrollTop = emojiTop;
    } else if (emojiBottom > visibleBottom) {
        emojiGrid.scrollTop = emojiBottom - gridRect.height;
    }
}

// Scroll to top
function scrollToTop() {
    emojiGrid.scrollTop = 0;
}

// Select and type emoji
async function selectEmoji(index) {
    if (index < 0 || index >= emojis.length) return;
    
    selectedIndex = index;
    updateSelection();
    
    const emoji = emojis[index];
    try {
        await window.go.main.App.TypeEmoji(emoji.emoji);
        // Close the app after typing
        window.runtime.Quit();
    } catch (err) {
        console.error('Failed to type emoji:', err);
        // Still quit on error
        window.runtime.Quit();
    }
}

// Start the app when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}
