# Nerd Font Icons

`fm` supports modern file and folder icons using Nerd Fonts. Because icons are not part of standard unicode, they require a specific font to be installed on your system.

## 1. Prerequisites

To see icons correctly, you must have a **Nerd Font** installed and active in your terminal. We recommend:
- JetBrainsMono Nerd Font
- Hack Nerd Font
- FiraCode Nerd Font

Download them from the [Official Nerd Fonts Website](https://www.nerdfonts.com/font-downloads).

## 2. Platform Setup

### Windows
1. Download and extract your chosen Nerd Font.
2. Select all `.ttf` files, right-click, and choose **Install**.
3. In your terminal (Windows Terminal, PowerShell, or Cmd), go to settings and set the font to the installed Nerd Font (e.g., `Hack Nerd Font Mono`).

### macOS
1. Install via Homebrew:
   ```bash
   brew tap homebrew/cask-fonts
   brew install --cask font-hack-nerd-font
   ```
2. In iTerm2 or Terminal.app, go to **Preferences > Profiles > Text** and change the font to the Nerd Font.

### Linux
1. Download and unzip the font to `~/.local/share/fonts`.
2. Update font cache:
   ```bash
   fc-cache -fv
   ```
3. Set your terminal emulator's font to the Nerd Font in its respective settings.

## 3. Special Case: VS Code (Integrated Terminal)

If you are using the terminal inside VS Code, you must explicitly tell VS Code to use the Nerd Font.

1. Open **Settings** (`Ctrl+,`).
2. Search for `Terminal Integrated Font Family`.
3. Set it to `'Hack Nerd Font Mono', monospace` (ensure the Nerd Font name matches exactly what is installed on your OS).

### VS Code in the Browser (Chrome/Edge)
If you are accessing VS Code via a browser (e.g., via SSH or Codespaces), you must also configure your browser:
1. Install the Nerd Font on your **Local Machine** (the one running the browser).
2. Open **Browser Settings > Appearance > Customize Fonts**.
3. Ensure the font is available or set as a fallback.
4. **Restart your browser** completely for the new OS fonts to be detected.

## 4. Enabling Icons in `fm`

Once your font is configured:
1. Open `fm`.
2. Press `.` to open **Settings**.
3. Navigate to **Enable Nerd Font Icons** and toggle it to **[ON]**.
4. If this is your first time, `fm` will download the icon mapping and perform a short visual test to ensure your font is working correctly.

## 5. Troubleshooting

If you see boxes with question marks (`?`) or hex codes:
- **Font Name**: Double check the font name string in your terminal settings. It often needs to end in `Mono`.
- **Restart**: Terminals and Browsers often require a full restart to see newly installed system fonts.
- **Support Test**: Run the built-in icon test in `fm`'s settings to verify rendering.
