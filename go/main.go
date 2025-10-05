package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Detectar sistema operativo
var osType = runtime.GOOS

// isWSL detecta si estamos en WSL
func isWSL() bool {
	if osType != "linux" {
		return false
	}
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

// setBrightness ajusta el brillo de la pantalla (0-100)
func setBrightness(level int) string {
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}

	var cmd *exec.Cmd

	if osType == "windows" || isWSL() {
		// Windows o WSL - usando PowerShell
		psCommand := "powershell"
		if isWSL() {
			psCommand = "powershell.exe"
		}
		script := fmt.Sprintf("(Get-WmiObject -Namespace root/WMI -Class WmiMonitorBrightnessMethods).WmiSetBrightness(1,%d)", level)
		cmd = exec.Command(psCommand, "-Command", script)
	} else if osType == "darwin" {
		// macOS - usando brightness CLI tool
		brightness := float64(level) / 100.0
		cmd = exec.Command("brightness", fmt.Sprintf("%.2f", brightness))
		if err := cmd.Run(); err != nil {
			// Fallback a AppleScript
			script := fmt.Sprintf("tell application \"System Events\" to set brightness of item 1 of (get displays) to %.2f", brightness)
			cmd = exec.Command("osascript", "-e", script)
		}
	} else {
		// Linux - usando xrandr
		output, err := exec.Command("sh", "-c", "xrandr | grep ' connected' | cut -d' ' -f1").Output()
		if err != nil {
			return fmt.Sprintf("❌ Error al obtener displays: %v", err)
		}
		displays := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(displays) > 0 && displays[0] != "" {
			brightness := float64(level) / 100.0
			cmd = exec.Command("xrandr", "--output", displays[0], "--brightness", fmt.Sprintf("%.2f", brightness))
		} else {
			return "❌ No se encontraron displays conectados"
		}
	}

	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("❌ Error al ajustar brillo: %v", err)
	}

	return fmt.Sprintf("✅ Brillo ajustado a %d%%", level)
}

// getBrightness obtiene el brillo actual de la pantalla
func getBrightness() string {
	var cmd *exec.Cmd
	current := 50 // Valor por defecto

	switch osType {
	case "windows":
		// Windows - usando PowerShell
		script := "(Get-WmiObject -Namespace root/WMI -Class WmiMonitorBrightness).CurrentBrightness"
		cmd = exec.Command("powershell", "-Command", script)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Sprintf("❌ Error al obtener brillo: %v", err)
		}
		val, _ := strconv.Atoi(strings.TrimSpace(string(output)))
		current = val
	case "darwin":
		// macOS - usando AppleScript
		script := "tell application \"System Events\" to get brightness of item 1 of (get displays)"
		cmd = exec.Command("osascript", "-e", script)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Sprintf("❌ Error al obtener brillo: %v", err)
		}
		val, _ := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
		current = int(val * 100)
	default:
		// Linux - xrandr no devuelve brillo fácilmente
		return "⚠️ Obtener brillo no implementado en Linux (usa xrandr manualmente)"
	}

	return fmt.Sprintf("💡 Brillo actual: %d%%", current)
}

// playSystemSound reproduce un sonido del sistema
func playSystemSound(soundType string) string {
	if soundType == "" {
		soundType = "default"
	}

	var cmd *exec.Cmd

	switch osType {
	case "windows":
		// Windows - usando PowerShell Beep
		frequencies := map[string][2]int{
			"beep":    {1000, 500},
			"alert":   {800, 300},
			"success": {1200, 200},
			"error":   {400, 500},
			"default": {1000, 500},
		}

		freq, ok := frequencies[soundType]
		if !ok {
			freq = frequencies["default"]
		}

		script := fmt.Sprintf("[console]::beep(%d,%d)", freq[0], freq[1])
		cmd = exec.Command("powershell", "-Command", script)
	case "darwin":
		// macOS - usando afplay
		sounds := map[string]string{
			"beep":    "/System/Library/Sounds/Ping.aiff",
			"alert":   "/System/Library/Sounds/Sosumi.aiff",
			"success": "/System/Library/Sounds/Glass.aiff",
			"error":   "/System/Library/Sounds/Basso.aiff",
			"default": "/System/Library/Sounds/Glass.aiff",
		}

		soundPath, ok := sounds[soundType]
		if !ok {
			soundPath = sounds["default"]
		}

		cmd = exec.Command("afplay", soundPath)
	default:
		// Linux - usando paplay
		cmd = exec.Command("paplay", "/usr/share/sounds/freedesktop/stereo/complete.oga")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("❌ Error al reproducir sonido: %v", err)
	}

	return fmt.Sprintf("🔔 Sonido '%s' reproducido", soundType)
}

// openApplication abre una aplicación específica
func openApplication(appName string) string {
	var cmd *exec.Cmd

	switch osType {
	case "windows":
		// Windows - usando start
		cmd = exec.Command("cmd", "/c", "start", appName)
	case "darwin":
		// macOS - usando open
		cmd = exec.Command("open", "-a", appName)
	default:
		// Linux - usando comando directo
		cmd = exec.Command("sh", "-c", appName+" &")
	}

	if err := cmd.Start(); err != nil {
		return fmt.Sprintf("❌ Error al abrir aplicación: %v", err)
	}

	return fmt.Sprintf("🚀 Aplicación '%s' abierta", appName)
}

// Estructuras para los inputs de las herramientas

type SetBrightnessInput struct {
	Level int `json:"level" jsonschema:"Nivel de brillo (0-100). 0=mínimo, 100=máximo"`
}

type PlaySoundInput struct {
	SoundType string `json:"sound_type,omitempty" jsonschema:"Tipo de sonido a reproducir: beep, alert, success, error, default"`
}

type OpenAppInput struct {
	AppName string `json:"app_name" jsonschema:"Nombre de la aplicación (ej: 'Calculator', 'Safari', 'chrome')"`
}

// Handlers de las herramientas

func HandleSetBrightness(ctx context.Context, req *mcp.CallToolRequest, input SetBrightnessInput) (*mcp.CallToolResult, any, error) {
	result := setBrightness(input.Level)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func HandleGetBrightness(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
	result := getBrightness()
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func HandlePlaySound(ctx context.Context, req *mcp.CallToolRequest, input PlaySoundInput) (*mcp.CallToolResult, any, error) {
	result := playSystemSound(input.SoundType)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func HandleOpenApp(ctx context.Context, req *mcp.CallToolRequest, input OpenAppInput) (*mcp.CallToolResult, any, error) {
	result := openApplication(input.AppName)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func main() {
	// Crear servidor MCP
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "hardware-control",
			Version: "1.0.0",
		},
		nil,
	)

	// Registrar herramienta: Ajustar brillo
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "set_brightness",
			Description: "Ajusta el brillo de la pantalla. Útil para presentaciones o trabajo nocturno.",
		},
		HandleSetBrightness,
	)

	// Registrar herramienta: Obtener brillo
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_brightness",
			Description: "Obtiene el nivel de brillo actual de la pantalla",
		},
		HandleGetBrightness,
	)

	// Registrar herramienta: Reproducir sonido
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "play_sound",
			Description: "Reproduce un sonido del sistema para notificar al usuario",
		},
		HandlePlaySound,
	)

	// Registrar herramienta: Abrir aplicación
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "open_app",
			Description: "Abre una aplicación específica en el sistema. En Windows usa el nombre del ejecutable, en macOS el nombre de la app.",
		},
		HandleOpenApp,
	)

	// Iniciar servidor
	log.Println("🚀 Iniciando servidor MCP de Control de Hardware...")
	log.Printf("📱 Sistema detectado: %s\n", osType)
	log.Println("💡 Herramientas disponibles:")
	log.Println("  - set_brightness: Ajustar brillo (0-100)")
	log.Println("  - get_brightness: Obtener brillo actual")
	log.Println("  - play_sound: Reproducir sonido del sistema")
	log.Println("  - open_app: Abrir aplicación")

	// Ejecutar servidor sobre stdin/stdout
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("❌ Error fatal: %v", err)
	}
}
