package ui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"

	"sekai-overlay-go/internal/config"
)

type Console struct {
	statusColor  *color.Color
	errorColor   *color.Color
	successColor *color.Color
	infoColor    *color.Color
}

func NewConsole() *Console {
	return &Console{
		statusColor:  color.New(color.FgCyan),
		errorColor:   color.New(color.FgRed, color.Bold),
		successColor: color.New(color.FgGreen, color.Bold),
		infoColor:    color.New(color.FgYellow),
	}
}

func (c *Console) PrintStatus(message string) {
	c.statusColor.Printf("🔄 %s\n", message)
}

func (c *Console) PrintError(message string) {
	c.errorColor.Printf("❌ %s\n", message)
}

func (c *Console) PrintSuccess(message string) {
	c.successColor.Printf("✅ %s\n", message)
}

func (c *Console) PrintInfo(message string) {
	c.infoColor.Printf("ℹ️  %s\n", message)
}

func (c *Console) PrintHeader(title string) {
	headerColor := color.New(color.FgMagenta, color.Bold)
	padded := ") " + title + "  "
	underline := strings.Repeat("=", len(padded))
	headerColor.Printf("\n%s\n", padded)
	fmt.Println(underline)
}

// PrintBanner はバナーを表示する
func (c *Console) PrintBanner() {
	banner := fmt.Sprintf(`
═════════════════════════════════════════════

- Sekai Overlay Go                
- Developed: はるくん(@Hallkun19)
- Repository: 
- Version: %s

═════════════════════════════════════════════
`, config.AppVersion)
	color.New(color.FgBlue, color.Bold).Print(banner)
}

func (c *Console) PrintPrompt(prompt string) {
	promptColor := color.New(color.FgCyan, color.Bold)
	promptColor.Printf("➜ %s", prompt)
}

func (c *Console) PrintKVTable(kv map[string]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	keyColor := color.New(color.FgHiCyan)
	for k, v := range kv {
		fmt.Fprintf(w, "%s:\t%s\n", keyColor.Sprintf(k), v)
	}
	w.Flush()
}

func (c *Console) OpenFolder(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // linux
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}

func (c *Console) ClearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
