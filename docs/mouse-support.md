# Mouse Support

`fm` now features a modern, comprehensive mouse interaction system designed to make terminal file management more intuitive without sacrificing the speed of keyboard-driven workflows.

## 🖱️ General Interactions

### Scrolling
- **Standardized Scrolling:** The mouse wheel now scrolls the view (offset) independently of the selection cursor across all screens.
- **Visual Stability:** The focused/highlighted item remains fixed on its original item while you scroll, providing a familiar modern experience.
- **Boundaries:** All views have strict scroll boundaries to prevent over-scrolling into empty space.

### Selection & Navigation
- **Single-Click:** Focuses an item without performing an action.
- **Double-Click:** Opens a file in your preferred editor or enters a directory.
- **Double-Click (Empty Space):** Double-clicking the empty area below the file list triggers the **New Item** prompt (similar to VSCode).
- **Selection Markers:** Click directly on the `[ ]` or `[x]` markers to toggle an item's selection state.
- **Shift + Click:** 
  - Click a single item to toggle selection.
  - Click another item to select a range from the current cursor position.
- **Drag-to-Select:** Click and drag on an empty area to select multiple items. Dragging back toward the starting point will dynamically unselect items.

### Drag & Drop (Move)
- **Drag-to-Move:** You can move files or folders by clicking and dragging them onto a target directory.
- **Batch Move:** If you have multiple items selected, dragging any one of them will move the entire selection into the target directory.

---

## 🧭 Header Interactions

### Breadcrumb Navigation
- **Direct Jump:** Every part of the path breadcrumb in the header is clickable. Click any parent folder name to jump directly to that directory.
- **Root Access:** Click the root indicator (`/`, `C:`, or `user@host`) to jump to the filesystem root.

### Tab Management
- **Switching Tabs:** Click on any tab indicator (e.g., `[1]`, `[2]`) to switch to that tab instantly.

---

## 🔎 Fuzzy Search Interactions

- **Result Focus:** Click any search match to focus it.
- **Double-Click to Open:** Double-clicking a match opens the file directly at the specific line number.
- **Expand/Collapse:** Click the `▼` or `▶` arrows next to filenames to expand or collapse the match results for that file.
- **Full Scrolling:** Use the mouse wheel to navigate through long lists of search results.

---

## ⚙️ Settings & Management Views

### Settings Menu
- **Navigation:** Click any setting to focus it or use the mouse wheel to scroll.
- **Toggling:** Double-click a setting to toggle its value (for ON/OFF options) or cycle through available choices (for Themes and Editors).

### Logs & Clipboard
- **Interaction:** Full mouse wheel scrolling and click-to-focus support for both the Operation Logs and Clipboard Manager views.
- **Back Button:** Click anywhere in the footer area of these views to exit (equivalent to pressing `Esc`).

---

## ⌨️ Text Input Interactions

- **Cursor Positioning:** Click anywhere within a text input field (Rename, Search, Go to, etc.) to move the cursor to that specific character.
- **Prompt Focus:** Click on the input prompt (e.g., "Rename:") to jump the cursor to the beginning of the text.
- **Interactive Hints:** Click on the tab hints on the right side of the footer (e.g., `[Tab] Folder`) to toggle between different input modes.

---

## 🛠️ Footer & Prompts

- **Action Shortcuts:** The action hints in the footer (e.g., `[c] Copy`, `[x] Cut`) are clickable.
- **Sort Mode:** Click the Sort Mode indicator on the right side of the footer to cycle through available sorting methods.
- **Confirmation Prompts:** 
  - All "Yes/No" prompts are clickable.
  - Conflict resolution prompts (Overwrite, Skip, Rename) are fully interactive via the mouse.

---

## 🔧 Configuration

Mouse support is enabled by default. You can toggle it in the Settings menu or by modifying your `config.json`:

```json
{
  "enable_mouse": true
}
```

> **Note:** If you find that mouse support is lost after returning from a terminal editor (like `vim`), `fm` automatically attempts to re-enable tracking on focus.
