# Promptly 🎨

A fast, minimalist Zsh prompt with an interactive installer. Choose your style, install instantly.

## Features

✨ **Interactive installer** - Preview themes before installing  
⚡ **High performance** - Optimized Git status parsing  
🎨 **Multiple themes** - Text-based and icon variants  
🔧 **Zero configuration** - Auto-installs and configures  
🌍 **Cross-platform** - Works on Linux, macOS, and Windows  

## Themes

### Default - Clean Text
![Default prompt](default.png)
Clean, readable prompt with Git status indicators using text symbols.

### Icons - Nerd Font Enhanced  
![Icons prompt](icons.png)
Beautiful icons with enhanced Git visualization (requires Nerd Font).

## Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/owlfacegames/promptly/master/install.sh | bash
```

Then run:
```bash
promptly
```

Use arrow keys to preview themes, press Enter to install. Restart your terminal to see changes.

## What it does

1. 🎯 Shows interactive theme selector with live previews
2. 📁 Installs chosen theme to `~/.promptly.zsh`  
3. ⚙️ Adds `source ~/.promptly.zsh` to your `.zshrc`
4. ✅ Ready to use immediately

## Manual Installation

Download the binary for your platform from [releases](https://github.com/owlfacegames/promptly/releases):

```bash
# Make executable and move to PATH
chmod +x promptly-*
sudo mv promptly-* /usr/local/bin/promptly
```

## Requirements

- **Zsh shell**
- **curl** (for installer)
- **Nerd Font** (for icons theme) - [Install here](https://www.nerdfonts.com/)

## Git Status Indicators

| Symbol | Meaning |
|--------|---------|
| `+2` | 2 staged files |
| `!1` | 1 modified file |
| `?3` | 3 untracked files |
| `$1` | 1 stashed change |
| `↑2` | 2 commits ahead |
| `↓1` | 1 commit behind |

## Uninstall

Remove from `.zshrc`:
```bash
# Delete this line from ~/.zshrc
source ~/.promptly.zsh
```

## License

MIT License - see [LICENSE](LICENSE) file.
