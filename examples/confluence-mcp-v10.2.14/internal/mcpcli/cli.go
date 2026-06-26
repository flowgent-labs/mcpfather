package mcpcli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	mcputils "confluence-mcp-v10.2.14/internal/helpers"
	mcptools "confluence-mcp-v10.2.14/internal/mcptools"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListTools prints all available tools as subcommands with one-line descriptions.
// Respects the config.yaml tools.enabled filter if present.
func ListTools() {
	cfg, _ := mcputils.LoadConfig("confluence-mcp-v10.2.14")
	enabled := make(map[string]bool)
	if cfg != nil && len(cfg.Tools.Include) > 0 {
		for _, name := range cfg.Tools.Include {
			enabled[name] = true
		}
	}

	names := make([]string, 0, len(mcptools.Registry))
	for name := range mcptools.Registry {
		if len(enabled) == 0 || enabled[name] {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	fmt.Printf("Available subcommands (%d):\n", len(names))
	for _, name := range names {
		entry := mcptools.Registry[name]
		fmt.Printf("  %-35s %s\n", name, shortDesc(entry.Tool.Description))
	}
}

// Call invokes a tool by name with GNU-style --flag arguments, forwarding as an HTTP request upstream.
func Call(binName, toolName string, args []string) error {
	entry, ok := mcptools.Registry[toolName]
	if !ok {
		return fmt.Errorf("unknown tool %q — use -t cli list to see available tools", toolName)
	}

	cfg, _ := mcputils.LoadConfig("confluence-mcp-v10.2.14")
	if cfg != nil && len(cfg.Tools.Include) > 0 {
		found := false
		for _, name := range cfg.Tools.Include {
			if name == toolName {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("tool %q is not enabled in config.yaml", toolName)
		}
	}

	// Check for --help / -h
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			printToolHelp(binName, toolName, entry)
			return nil
		}
	}

	mcpArgs, err := parseGNUArgs(args)
	if err != nil {
		return err
	}

	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: mcpArgs,
		},
	}

	result, err := entry.Handler(context.Background(), req)
	if err != nil {
		return fmt.Errorf("tool %s failed: %w", toolName, err)
	}

	if result == nil {
		fmt.Println("(no result)")
		return nil
	}

	if result.IsError {
		for _, content := range result.Content {
			if tc, ok := content.(mcp.TextContent); ok {
				fmt.Fprintln(os.Stderr, tc.Text)
			}
		}
		return nil
	}

	for _, content := range result.Content {
		switch c := content.(type) {
		case mcp.TextContent:
			fmt.Print(c.Text)
		default:
			data, _ := json.MarshalIndent(content, "", "  ")
			fmt.Println(string(data))
		}
	}

	return nil
}

// parseGNUArgs parses GNU-style --flag value and --flag=value arguments.
func parseGNUArgs(args []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			return nil, fmt.Errorf("unexpected argument %q — arguments must be in --name=value or --name value format", arg)
		}
		arg = arg[2:] // strip --
		if arg == "" {
			return nil, fmt.Errorf("empty flag name after '--'")
		}
		var key, rawVal string
		if idx := strings.IndexByte(arg, '='); idx >= 0 {
			key = arg[:idx]
			rawVal = arg[idx+1:]
		} else {
			key = arg
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
				return nil, fmt.Errorf("--%s requires a value", key)
			}
			i++
			rawVal = args[i]
		}
		if key == "" {
			return nil, fmt.Errorf("empty flag name in %q", "--"+arg)
		}
		result[key] = parseValue(rawVal)
	}
	return result, nil
}

func parseValue(raw string) interface{} {
	if raw == "true" {
		return true
	}
	if raw == "false" {
		return false
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return float64(n)
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return f
	}
	return raw
}

// printToolHelp prints GNU-style usage for a single tool.
func printToolHelp(binName, toolName string, entry mcptools.ToolEntry) {
	t := entry.Tool

	fmt.Printf("Usage: %s -t cli %s [OPTIONS]\n\n", binName, toolName)
	if t.Description != "" {
		fmt.Println(plainText(t.Description))
		fmt.Println()
	}

	var schema struct {
		Properties map[string]interface{} `json:"properties"`
		Required   []string               `json:"required"`
	}
	if err := json.Unmarshal(t.RawInputSchema, &schema); err != nil || len(schema.Properties) == 0 {
		fmt.Println("No options.")
		return
	}

	required := make(map[string]bool)
	for _, r := range schema.Required {
		required[r] = true
	}

	props := schema.Properties

	fmt.Println("Options:")
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names)

	colWidth := 0
	for _, name := range names {
		w := len(name) + 3 // "--" + name
		if !isBoolProp(props[name]) {
			w += 4 // " <type>"
		}
		if required[name] {
			w += 11 // " (required)"
		}
		if w > colWidth {
			colWidth = w
		}
	}
	if colWidth > 40 {
		colWidth = 40
	}

	for _, name := range names {
		prop, _ := props[name].(map[string]interface{})
		origDesc, _ := prop["description"].(string)
		desc := firstLine(plainText(origDesc))
		propType, _ := prop["type"].(string)

		s := fmt.Sprintf("  --%s", name)
		if !isBoolProp(props[name]) {
			if propType != "" {
				s += fmt.Sprintf(" <%s>", propType)
			} else {
				s += " <value>"
			}
		}
		if required[name] {
			s += " (required)"
		}
		fmt.Print(s)
		if desc != "" {
			if len(s) < colWidth {
				fmt.Print(strings.Repeat(" ", colWidth-len(s)))
			}
			fmt.Print("  ", desc)
		}
		fmt.Println()
	}
}

func isBoolProp(prop interface{}) bool {
	p, ok := prop.(map[string]interface{})
	if !ok {
		return false
	}
	t, _ := p["type"].(string)
	return t == "boolean"
}

// plainText strips common Markdown formatting for GNU-style help output.
func plainText(s string) string {
	var out strings.Builder
	inCodeBlock := false
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Code block fences
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			out.WriteByte('\n')
			continue
		}
		if inCodeBlock {
			out.WriteString("    ")
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}

		// Table separator rows
		if strings.Count(trimmed, "|") >= 3 && strings.Contains(trimmed, "---") {
			continue
		}
		// Table data rows
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") && strings.Count(trimmed, "|") >= 2 {
			line = strings.TrimPrefix(trimmed, "| ")
			line = strings.TrimSuffix(line, " |")
			line = strings.ReplaceAll(line, " | ", "  ")
			line = "  " + line
		}

		// Headers (use processed line so far, but for structural transforms use trimmed)
		if strings.HasPrefix(trimmed, "### ") {
			line = "  " + strings.TrimPrefix(trimmed, "### ")
		} else if strings.HasPrefix(trimmed, "## ") {
			line = strings.TrimPrefix(trimmed, "## ")
		} else if strings.HasPrefix(trimmed, "# ") {
			line = strings.TrimPrefix(trimmed, "# ")
		}

		// Blockquotes
		if strings.HasPrefix(trimmed, "> ") {
			line = "  " + strings.TrimPrefix(trimmed, "> ")
		} else if strings.HasPrefix(trimmed, ">") {
			line = "  " + strings.TrimPrefix(trimmed, ">")
		}

		// Strip inline formatting (after structural transforms)
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "`", "")
		line = strings.ReplaceAll(line, "__", "")

		out.WriteString(line)
		out.WriteByte('\n')
	}
	return strings.TrimSpace(out.String())
}

func shortDesc(desc string) string {
	s := desc
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	s = strings.TrimRight(s, ". ")
	return s
}

func firstLine(s string) string {
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}
