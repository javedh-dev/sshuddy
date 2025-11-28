<pre>
╔═╗┌─┐┬ ┬  ╔╗ ┬ ┬┌┬┐┌┬┐┬ ┬
╚═╗└─┐├─┤  ╠╩╗│ │ ││ ││└┬┘
╚═╝└─┘┴ ┴  ╚═╝└─┘─┴┘─┴┘ ┴
</pre>

---

A beautiful TUI app to manage SSH connections with live status indicators, multiple themes, and an intuitive 2-column interface.

![SSH Buddy Screenshot](screenshots/purple.png)

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

- 🟢 **Live ping status indicators** - See which hosts are online at a glance
- 🎨 **Multiple color themes** - 6 beautiful themes to choose from
- 📋 **2-column layout** - View more hosts at once with row-wise display
- 📝 **Duplicate hosts** - Quickly copy existing configurations
- 🏷️ **Host tagging** - Organize hosts with custom tags
- ✨ **Beautiful, modern UI** - Clean and intuitive interface
- ⚡ **Fast and responsive** - Instant feedback and smooth navigation
- 🔍 **Built-in search** - Filter hosts on the fly
- 💾 **Automatic config persistence** - Saved in `~/.config/sshbuddy/config.json`
- ⌨️ **Keyboard-driven** - Navigate without touching the mouse

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Enter` | Connect to selected host |
| `n` | Add new host |
| `e` | Edit selected host |
| `c` | Duplicate/copy selected host |
| `d` | Delete selected host |
| `p` | Ping all hosts |
| `t` | Switch theme |
| `/` | Search/filter hosts |
| `↑`/`↓` | Navigate up/down |
| `←`/`→` | Navigate left/right (between columns) |
| `q` | Quit |
| `Esc` | Cancel (in forms/dialogs) |
| `Ctrl+C` | Force quit |

## Themes

Press `t` to cycle through available themes. Your theme preference is automatically saved.

### Purple Dream (Default)
Soft purple tones for a calm, modern look.

![Purple Theme](screenshots/purple.png)

### Ocean Blue
Cool blue tones inspired by the ocean.

![Blue Theme](screenshots/blue.png)

### Matrix Green
Classic terminal green for that retro hacker vibe.

![Green Theme](screenshots/green.png)

### Bubblegum Pink
Vibrant pink for a fun, energetic interface.

![Pink Theme](screenshots/pink.png)

### Sunset Amber
Warm amber/orange tones like a beautiful sunset.

![Amber Theme](screenshots/amber.png)

### Cyber Cyan
Electric cyan for a futuristic cyberpunk aesthetic.

![Cyan Theme](screenshots/cyan.png)

## Configuration

SSH Buddy stores its configuration in:
- **Linux/Unix**: `~/.config/sshbuddy/config.json`
- **Respects XDG_CONFIG_HOME**: If set, uses `$XDG_CONFIG_HOME/sshbuddy/config.json`

The config file is automatically created on first run and includes:
- Host configurations (alias, hostname, user, port, tags)
- Theme preference
- All settings are validated on load

### Example Config

```json
{
  "hosts": [
    {
      "alias": "Production Server",
      "hostname": "prod.example.com",
      "user": "admin",
      "port": "22",
      "tags": ["production", "web"]
    },
    {
      "alias": "Dev Server",
      "hostname": "192.168.1.100",
      "user": "developer",
      "port": "2222",
      "tags": ["development"]
    }
  ],
  "theme": "purple"
}
```

## Tips

- **Quick duplicate**: Press `c` on any host to create a copy with " (copy)" appended to the alias
- **Arrow navigation**: Use arrow keys to navigate the 2-column grid naturally (left/right for columns, up/down for rows)
- **Search**: Press `/` and start typing to filter hosts by alias or hostname
- **Ping status**: Green dot = online, Red dot = offline, Gray dot = unknown, Yellow dot = checking
- **Tags**: Add comma-separated tags when creating/editing hosts for better organization

---

Enjoy managing your SSH connections! 🚀
