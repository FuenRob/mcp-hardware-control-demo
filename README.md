# MCP Hardware Control Demo

A demonstration project showcasing Model Context Protocol (MCP) servers for local hardware control. This project provides cross-platform implementations in both Go and TypeScript that allow AI assistants to interact with system hardware through standardized MCP tools.

## Features

This MCP server exposes four hardware control tools:

- **set_brightness**: Adjust screen brightness (0-100%)
- **get_brightness**: Get current screen brightness level
- **play_sound**: Play system notification sounds (beep, alert, success, error, default)
- **open_app**: Launch applications by name

## Supported Platforms

- **Windows** (including WSL)
- **macOS**
- **Linux** (with X11)

## Project Structure

```
mcp-hardware-control-demo-main/
├── go/                    # Go implementation
│   ├── main.go           # Main server code
│   ├── go.mod            # Go module dependencies
│   └── go.sum            # Go module checksums
├── typescript/           # TypeScript implementation
│   ├── src/index.ts      # Main server code
│   ├── package.json      # Node.js dependencies
│   ├── tsconfig.json     # TypeScript configuration
│   └── build/            # Compiled JavaScript output
└── README.md             # This file
```

## Installation

### Go Version

1. Ensure you have Go 1.25+ installed
2. Navigate to the go directory:
   ```bash
   cd go
   ```
3. Install dependencies:
   ```bash
   go mod tidy
   ```
4. Build the server:
   ```bash
   go build -o mcp-hardware-control main.go
   ```

### TypeScript Version

1. Ensure you have Node.js 18+ installed
2. Navigate to the typescript directory:
   ```bash
   cd typescript
   ```
3. Install dependencies:
   ```bash
   npm install
   ```
4. Build the project:
   ```bash
   npm run build
   ```

## Usage

### Running the Server

#### Go Version
```bash
cd go
./mcp-hardware-control
```

#### TypeScript Version
```bash
cd typescript
npm start
```

The server communicates via stdin/stdout using the MCP protocol and should be integrated with MCP-compatible clients.

### Tool Descriptions

#### set_brightness
Adjusts the screen brightness level.

**Parameters:**
- `level` (integer, 0-100): Brightness level (0 = minimum, 100 = maximum)

**Example:**
```json
{
  "level": 75
}
```

#### get_brightness
Retrieves the current screen brightness level.

**Parameters:** None

**Returns:** Current brightness percentage

#### play_sound
Plays a system notification sound.

**Parameters:**
- `sound_type` (string, optional): Type of sound to play
  - `"beep"`: Standard beep
  - `"alert"`: Alert sound
  - `"success"`: Success notification
  - `"error"`: Error notification
  - `"default"`: Default system sound

#### open_app
Opens a specified application.

**Parameters:**
- `app_name` (string): Name of the application to open
  - Windows: Executable name (e.g., "notepad", "calc")
  - macOS: Application name (e.g., "Calculator", "Safari")
  - Linux: Command name

## Platform-Specific Notes

### Windows
- Uses PowerShell for brightness control via WMI
- Uses `[console]::beep()` for sound playback
- Uses `start` command for opening applications

### macOS
- Uses `brightness` CLI tool (with AppleScript fallback) for brightness
- Uses `afplay` for system sounds
- Uses `open -a` for applications

### Linux
- Uses `xrandr` for brightness control
- Uses `paplay` for sound playback
- Uses direct command execution for applications

## Dependencies

### Go Version
- `github.com/modelcontextprotocol/go-sdk v1.0.0`

### TypeScript Version
- `@modelcontextprotocol/sdk ^1.0.4`
- `zod ^3.23.8`

## Development

### Go Development
```bash
cd go
go run main.go
```

### TypeScript Development
```bash
cd typescript
npm run dev  # Watch mode compilation
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test on multiple platforms
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Author

Roberto Morais

## Related Projects

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP TypeScript SDK](https://github.com/modelcontextprotocol/typescript-sdk)
