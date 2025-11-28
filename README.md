# SSH Buddy

A beautiful TUI app to manage SSH connections with live status indicators, multiple themes, and an intuitive interface.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap javedh-dev/tap
brew install sshbuddy
```

### From Source

```bash
git clone https://github.com/javedh-dev/sshbuddy.git
cd sshbuddy
go build .
./sshbuddy
```

### Download Binary

Download the latest release from the [releases page](https://github.com/javedh-dev/sshbuddy/releases).

## Quick Start

1. **Run the app:**
   ```bash
   sshbuddy
   ```

3. **Add a host:**
   - Press `n`
   - Fill in the details
   - Press Enter to save

4. **Connect:**
   - Select a host with arrow keys
   - Press Enter to connect

## Features

- 🟢 Live ping status indicators
- 🎨 Multiple color themes (6 themes available)
- ✨ Beautiful, modern UI
- ⚡ Fast and responsive
- 🔍 Built-in search (press `/`)
- 💾 Automatic config persistence

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `n` | Add new host |
| `e` | Edit selected host |
| `d` | Delete selected host |
| `p` | Ping all hosts |
| `t` | Switch theme |
| `Enter` | Connect to selected host |
| `/` | Search/filter |
| `↑`/`↓` | Navigate |
| `q` | Quit |
| `Esc` | Cancel (in forms) |

## Themes

Press `t` to cycle through available themes:
- **Purple Dream** - Soft purple (default)
- **Ocean Blue** - Cool blue tones
- **Matrix Green** - Classic terminal green
- **Bubblegum Pink** - Vibrant pink
- **Sunset Amber** - Warm amber/orange
- **Cyber Cyan** - Electric cyan

Your theme preference is automatically saved.

---

Enjoy managing your SSH connections! 🚀
