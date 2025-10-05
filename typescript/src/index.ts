#!/usr/bin/env node

/**
 * Servidor MCP para controlar hardware local (brillo, volumen, sonidos)
 * Instalación: npm install @modelcontextprotocol/sdk zod
 * Dependencias adicionales:
 * - Windows: npm install node-screen-brightness loudness
 * - macOS/Linux: Solo requiere comandos del sistema
 */

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { z } from "zod";
import { spawn, execSync } from "child_process";
import { platform } from "os";

// Detectar sistema operativo
const OS = platform();
const isWSL = OS === "linux" && existsSync("/proc/version") && 
              readFileSync("/proc/version", "utf8").toLowerCase().includes("microsoft");

import { existsSync, readFileSync } from "fs";

// === FUNCIONES DE CONTROL DE HARDWARE ===

/**
 * Ajusta el brillo de la pantalla (0-100)
 */
async function setBrightness(level: number): Promise<string> {
  try {
    const clampedLevel = Math.max(0, Math.min(100, level));

    if (OS === "win32" || isWSL) {
      // Windows o WSL - usando PowerShell
      const psCommand = isWSL ? "powershell.exe" : "powershell";
      const script = `(Get-WmiObject -Namespace root/WMI -Class WmiMonitorBrightnessMethods).WmiSetBrightness(1,${clampedLevel})`;
      execSync(`${psCommand} -Command "${script}"`, { stdio: "pipe" });
    } else if (OS === "darwin") {
      // macOS - usando brightness CLI tool
      try {
        execSync(`brightness ${clampedLevel / 100}`, { stdio: "pipe" });
      } catch {
        // Fallback a AppleScript si brightness no está instalado
        const script = `tell application "System Events" to set brightness of item 1 of (get displays) to ${clampedLevel / 100}`;
        execSync(`osascript -e '${script}'`, { stdio: "pipe" });
      }
    } else {
      // Linux - usando xrandr
      const displays = execSync("xrandr | grep ' connected' | cut -d' ' -f1")
        .toString()
        .trim()
        .split("\n");
      if (displays.length > 0) {
        execSync(
          `xrandr --output ${displays[0]} --brightness ${clampedLevel / 100}`,
          { stdio: "pipe" }
        );
      }
    }

    return `✅ Brillo ajustado a ${clampedLevel}%`;
  } catch (error) {
    return `❌ Error al ajustar brillo: ${error instanceof Error ? error.message : String(error)}`;
  }
}

/**
 * Obtiene el brillo actual de la pantalla
 */
async function getBrightness(): Promise<string> {
  try {
    let current = 50; // Valor por defecto

    if (OS === "win32") {
      // Windows - usando PowerShell
      const script = `(Get-WmiObject -Namespace root/WMI -Class WmiMonitorBrightness).CurrentBrightness`;
      const output = execSync(`powershell -Command "${script}"`, {
        encoding: "utf-8",
      });
      current = parseInt(output.trim());
    } else if (OS === "darwin") {
      // macOS - usando AppleScript
      const script = `tell application "System Events" to get brightness of item 1 of (get displays)`;
      const output = execSync(`osascript -e '${script}'`, {
        encoding: "utf-8",
      });
      current = Math.round(parseFloat(output.trim()) * 100);
    } else {
      // Linux - xrandr no devuelve brillo fácilmente
      return "⚠️ Obtener brillo no implementado en Linux (usa xrandr manualmente)";
    }

    return `💡 Brillo actual: ${current}%`;
  } catch (error) {
    return `❌ Error al obtener brillo: ${error instanceof Error ? error.message : String(error)}`;
  }
}

/**
 * Reproduce un sonido del sistema
 */
async function playSystemSound(soundType: string = "default"): Promise<string> {
  try {
    if (OS === "win32") {
      // Windows - usando PowerShell Beep
      const frequencies: Record<string, [number, number]> = {
        beep: [1000, 500],
        alert: [800, 300],
        success: [1200, 200],
        error: [400, 500],
        default: [1000, 500],
      };

      const [freq, duration] = frequencies[soundType] || frequencies.default;
      const script = `[console]::beep(${freq},${duration})`;
      execSync(`powershell -Command "${script}"`, { stdio: "pipe" });
    } else if (OS === "darwin") {
      // macOS - usando afplay
      const sounds: Record<string, string> = {
        beep: "/System/Library/Sounds/Ping.aiff",
        alert: "/System/Library/Sounds/Sosumi.aiff",
        success: "/System/Library/Sounds/Glass.aiff",
        error: "/System/Library/Sounds/Basso.aiff",
        default: "/System/Library/Sounds/Glass.aiff",
      };

      const soundPath = sounds[soundType] || sounds.default;
      execSync(`afplay "${soundPath}"`, { stdio: "pipe" });
    } else {
      // Linux - usando paplay
      execSync(`paplay /usr/share/sounds/freedesktop/stereo/complete.oga`, {
        stdio: "pipe",
      });
    }

    return `🔔 Sonido '${soundType}' reproducido`;
  } catch (error) {
    return `❌ Error al reproducir sonido: ${error instanceof Error ? error.message : String(error)}`;
  }
}

/**
 * Abre una aplicación específica
 */
async function openApplication(appName: string): Promise<string> {
  try {
    if (OS === "win32") {
      // Windows - usando start
      execSync(`start ${appName}`, { stdio: "pipe" });
    } else if (OS === "darwin") {
      // macOS - usando open
      execSync(`open -a "${appName}"`, { stdio: "pipe" });
    } else {
      // Linux - usando xdg-open o comando directo
      execSync(`${appName} &`, { stdio: "pipe" });
    }

    return `🚀 Aplicación '${appName}' abierta`;
  } catch (error) {
    return `❌ Error al abrir aplicación: ${error instanceof Error ? error.message : String(error)}`;
  }
}

// === CONFIGURACIÓN DEL SERVIDOR MCP ===

const server = new McpServer({
  name: "hardware-control",
  version: "1.0.0",
});

// Registrar herramienta: Ajustar brillo
server.registerTool(
  "set_brightness",
  {
    title: "Ajustar Brillo",
    description:
      "Ajusta el brillo de la pantalla. Útil para presentaciones o trabajo nocturno.",
    inputSchema: {
      level: z
        .number()
        .min(0)
        .max(100)
        .describe("Nivel de brillo (0-100). 0=mínimo, 100=máximo"),
    },
  },
  async ({ level }) => {
    const result = await setBrightness(level);
    return {
      content: [{ type: "text", text: result }],
    };
  }
);

// Registrar herramienta: Obtener brillo
server.registerTool(
  "get_brightness",
  {
    title: "Obtener Brillo",
    description: "Obtiene el nivel de brillo actual de la pantalla",
    inputSchema: {},
  },
  async () => {
    const result = await getBrightness();
    return {
      content: [{ type: "text", text: result }],
    };
  }
);

// Registrar herramienta: Reproducir sonido
server.registerTool(
  "play_sound",
  {
    title: "Reproducir Sonido",
    description: "Reproduce un sonido del sistema para notificar al usuario",
    inputSchema: {
      sound_type: z
        .enum(["beep", "alert", "success", "error", "default"])
        .optional()
        .describe("Tipo de sonido a reproducir"),
    },
  },
  async ({ sound_type }) => {
    const result = await playSystemSound(sound_type || "default");
    return {
      content: [{ type: "text", text: result }],
    };
  }
);

// Registrar herramienta: Abrir aplicación
server.registerTool(
  "open_app",
  {
    title: "Abrir Aplicación",
    description:
      "Abre una aplicación específica en el sistema. En Windows usa el nombre del ejecutable, en macOS el nombre de la app.",
    inputSchema: {
      app_name: z
        .string()
        .describe(
          "Nombre de la aplicación (ej: 'Calculator', 'Safari', 'chrome')"
        ),
    },
  },
  async ({ app_name }) => {
    const result = await openApplication(app_name);
    return {
      content: [{ type: "text", text: result }],
    };
  }
);

// === INICIAR SERVIDOR ===

async function main() {
  console.error("🚀 Iniciando servidor MCP de Control de Hardware...");
  console.error(`📱 Sistema detectado: ${OS}`);
  console.error(`💡 Herramientas disponibles: 
  - set_brightness: Ajustar brillo (0-100)
  - get_brightness: Obtener brillo actual
  - play_sound: Reproducir sonido del sistema
  - open_app: Abrir aplicación
  `);

  const transport = new StdioServerTransport();
  await server.connect(transport);

  console.error("✅ Servidor MCP iniciado y listo para recibir conexiones");
}

main().catch((error) => {
  console.error("❌ Error fatal:", error);
  process.exit(1);
});