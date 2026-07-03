package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/server"
	mcputils "sonarqube-mcp/pkg/helpers"
	mcpcli "sonarqube-mcp/pkg/mcpcli"
	mcpserver "sonarqube-mcp/pkg/mcpserver"
)

// logHTTP wraps any handler and logs every incoming HTTP request + response
func logHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		v := mcputils.GetVerbosity()

		// Extract session ID for logging and pass it downstream via context
		sessionID := r.Header.Get("Mcp-Session-Id")
		ctx := mcputils.WithSessionID(r.Context(), sessionID)
		r = r.WithContext(ctx)

		if v >= 2 {
			var bodyLog string
			if r.Body != nil {
				body, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(strings.NewReader(string(body)))
				bodyLog = string(body)
			}
			sid := sessionID
			if sid == "" {
				sid = "-"
			}
			fmt.Fprintf(os.Stderr, "%s [http] sid=%s %s %s body=%s\n", start.Format(time.RFC3339), sid, r.Method, r.URL.RequestURI(), truncate(bodyLog, 300))
		}

		lw := &loggingWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(lw, r)

		if v >= 1 {
			sid := sessionID
			if sid == "" {
				sid = "-"
			}
			fmt.Fprintf(os.Stderr, "%s [http] sid=%s %d %s %s (%s)\n", time.Now().Format(time.RFC3339), sid, lw.status, r.Method, r.URL.RequestURI(), time.Since(start).Round(time.Millisecond))
		}
	})
}

type loggingWriter struct {
	http.ResponseWriter
	status int
}

func (l *loggingWriter) WriteHeader(code int) {
	l.status = code
	l.ResponseWriter.WriteHeader(code)
}

// Flush implements http.Flusher by delegating to the underlying ResponseWriter.
func (l *loggingWriter) Flush() {
	if f, ok := l.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	if s == "" {
		return "(empty)"
	}
	return s
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage: sonarqube-mcp [OPTIONS] [TOOL_NAME] [JSON_ARGS]

Options:
  -t, --transport <stdio|http|cli>  Transport mode (default "stdio")
                                     cli mode: invoke a tool directly from the command line
  -p, --port <number>                HTTP server port (default 8080)
  -v, --verbose <0-10>              Request logging verbosity level
                                     0=silent, 1=access log, 2+=method+URL,
                                     3+=query, 5+=headers, 7+=body, 9+=pretty JSON
  --print-default-config             Print default config.yaml to stdout and exit
  -h, --help                        Show this help message

CLI Mode:
  sonarqube-mcp -t cli list                   List all available tools
  sonarqube-mcp -t cli <tool-name> [OPTIONS]   Invoke a tool with GNU-style options
                                                   Use --help for tool-specific help
`)
}

func printDefaultConfigYAML() {
	fmt.Println("# sonarqube-mcp MCP server configuration")
	fmt.Println("# Place this file at: $HOME/." + "sonarqube-mcp" + "/config.yaml")
	fmt.Println()
	fmt.Println("# ---- Native MCP Tools ----")
	fmt.Println("tools:")
	fmt.Println("  expose:")
	fmt.Println("    # When true, all native tools are registered by default")
	fmt.Println("    # (individual tools can be hidden via excludes).")
	fmt.Println("    # When false (the default), only tools listed in includes")
	fmt.Println("    # are exposed.")
	fmt.Println("    all-native-tools-by-default: false")
	fmt.Println()
	fmt.Println("    # Explicitly expose these tools (operationId values).")
	fmt.Println("    # Takes precedence: tools listed here are always exposed,")
	fmt.Println("    # even if all-native-tools-by-default is false.")
	fmt.Println("    includes: []")
	fmt.Println()
	fmt.Println("    # Explicitly hide these tools from external agents (operationId values).")
	fmt.Println("    # Takes precedence over all-native-tools-by-default.")
	fmt.Println("    # A tool listed in BOTH includes and excludes will cause")
	fmt.Println("    # the server to fail at startup.")
	fmt.Println("    excludes: []")
	fmt.Println()
	fmt.Println("# ---- Upstream Request Forwarding ----")
	fmt.Println("upstream:")
	fmt.Println("  # When true, the MCP session ID from client requests is forwarded")
	fmt.Println("  # to upstream APIs as the X-MCP-Session-ID HTTP header.")
	fmt.Println("  # (DEPRECATED: session ID is always forwarded as X-MCP-Session-ID)")
	fmt.Println("  enable_mcp_session_in_forwarding: false")
	fmt.Println()
	fmt.Println("# ---- Virtual Tools ----")
	fmt.Println("# Compose multiple native tools into a single virtual tool via a")
	fmt.Println("# declarative e.g: 5-step pipeline (call → jq → foreach → emit → return).")
	fmt.Println("# Schema: https://github.com/wl4g-ai/mcpgen-go/blob/main/.agents/skills/virtual-tool-creator/resources/dsl-schema.json")
	fmt.Println("virtualTools: []")
	fmt.Println("# Example:")
	fmt.Println("#   - name: MyVirtualTool")
	fmt.Println("#     description: Retrieve application details with remediation suggestions")
	fmt.Println("#     inputSchema:")
	fmt.Println("#       type: object")
	fmt.Println("#       properties:")
	fmt.Println("#         appId:")
	fmt.Println("#           type: string")
	fmt.Println("#       required: [appId]")
	fmt.Println("#     pipeline:")
	fmt.Println("#       - id: getApp")
	fmt.Println("#         kind: call")
	fmt.Println("#         spec:")
	fmt.Println("#           tool: GetApplication")
	fmt.Println("#           parse: json")
	fmt.Println("#           args:")
	fmt.Println("#             applicationId: $input.appId")
	fmt.Println("#       - id: appName")
	fmt.Println("#         kind: jq")
	fmt.Println("#         spec:")
	fmt.Println("#           from: $getApp")
	fmt.Println("#           expr: .name")
	fmt.Println("#       - id: done")
	fmt.Println("#         kind: return")
	fmt.Println("#         spec:")
	fmt.Println("#           from: $appName")
}

func main() {
	flag.CommandLine.SetOutput(os.Stderr)
	flag.Usage = usage

	transport := flag.String("t", "stdio", "Transport mode: stdio, http, or cli")
	flag.StringVar(transport, "transport", "stdio", "Transport mode: stdio, http, or cli")
	port := flag.Int("p", 8080, "HTTP server port (only used when transport=http)")
	flag.IntVar(port, "port", 8080, "HTTP server port (only used when transport=http)")
	verbose := flag.Int("v", 0, "Request logging verbosity level (0-10)")
	flag.IntVar(verbose, "verbose", 0, "Request logging verbosity level (0-10)")
	printDefaultConfig := flag.Bool("print-default-config", false, "Print default config.yaml to stdout and exit")
	flag.Parse()

	if *printDefaultConfig {
		printDefaultConfigYAML()
		return
	}

	upstreamEndpoint := mcputils.GetUpstreamEndpoint()
	isDefault := mcputils.IsDefaultUpstreamEndpoint()

	fmt.Fprintf(os.Stderr, "MCP_UPSTREAM_ENDPOINT=%s\n", upstreamEndpoint)
	if isDefault {
		fmt.Fprintln(os.Stderr, "Warning: MCP_UPSTREAM_ENDPOINT is not set. Using default (httpbin.org/anything) which echoes requests. Set it to your actual upstream service endpoint.")
	}

	mcputils.SetVerbosity(*verbose)

	switch *transport {
	case "cli":
		if flag.NArg() == 0 {
			fmt.Fprintf(os.Stderr, "Usage: %s -t cli <tool-name> [OPTIONS]\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "       %s -t cli list\n", os.Args[0])
			fmt.Fprintf(os.Stderr, "\nUse -t cli <tool-name> --help for tool-specific help.\n")
			os.Exit(1)
		}
		subcmd := flag.Arg(0)
		switch subcmd {
		case "list":
			mcpcli.ListTools()
		default:
			if subcmd == "--help" || subcmd == "-h" {
				fmt.Fprintf(os.Stderr, "Usage: %s -t cli <tool-name> [OPTIONS]\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "       %s -t cli list\n", os.Args[0])
				fmt.Fprintf(os.Stderr, "\nUse -t cli <tool-name> --help for tool-specific help.\n")
				os.Exit(0)
			}
			toolName := subcmd
			args := flag.Args()[1:]
			if err := mcpcli.Call(os.Args[0], toolName, args); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}

	case "http":
		s, err := mcpserver.NewMCPServer()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		mcpServer := server.NewStreamableHTTPServer(s,
			server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
				return mcputils.WithHTTPHeaders(ctx, r.Header)
			}),
		)

		mux := http.NewServeMux()
		mux.Handle("/mcp", mcpServer)

		addr := fmt.Sprintf(":%d", *port)
		httpServer := &http.Server{Addr: addr, Handler: logHTTP(mux)}

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		go func() {
			fmt.Fprintf(os.Stderr, "MCP server listening on %s/mcp\n", addr)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
				os.Exit(1)
			}
		}()

		<-ctx.Done()
		fmt.Fprintln(os.Stderr, "Shutting down...")
		if err := httpServer.Shutdown(context.WithoutCancel(ctx)); err != nil {
			fmt.Fprintf(os.Stderr, "HTTP server shutdown error: %v\n", err)
		}

	case "stdio":
		s, err := mcpserver.NewMCPServer()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := server.ServeStdio(s); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown transport: %s (use 'stdio', 'http', or 'cli')\n", *transport)
		os.Exit(1)
	}
}
