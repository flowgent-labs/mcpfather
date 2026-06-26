package mcptools

import (
	"bytes"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp-v10.7.4/internal/helpers"
	"net/http"
	"os"
	"strings"
	"time"
)

// Input Schema for the StoreTemporaryAvatarUsingMultiPart2 tool
const StoreTemporaryAvatarUsingMultiPart2InputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"contentType\": {\n          \"type\": \"string\"\n        },\n        \"formField\": {\n          \"type\": \"boolean\"\n        },\n        \"inputStream\": {\n          \"type\": \"object\"\n        },\n        \"name\": {\n          \"type\": \"string\"\n        },\n        \"size\": {\n          \"format\": \"int64\",\n          \"type\": \"integer\"\n        },\n        \"value\": {\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"owningObjectId\": {\n      \"description\": \"Entity id where to change avatar\",\n      \"type\": \"string\"\n    },\n    \"type\": {\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"owningObjectId\",\n    \"type\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the StoreTemporaryAvatarUsingMultiPart2 tool (Status: 200, Content-Type: application/json)
const StoreTemporaryAvatarUsingMultiPart2ResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns temporary avatar cropping instructions.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **cropperOffsetY** (Type: integer, int32):\n      - Example: '50'\n  - **cropperWidth** (Type: integer, int32):\n      - Example: '120'\n  - **needsCropping** (Type: boolean):\n      - Example: 'true'\n  - **url** (Type: string):\n      - Example: 'http://example.com/jira/secure/temporaryavatar?cropped=true'\n  - **cropperOffsetX** (Type: integer, int32):\n      - Example: '50'\n"

// NewStoreTemporaryAvatarUsingMultiPart2MCPTool creates the MCP Tool instance for StoreTemporaryAvatarUsingMultiPart2
func NewStoreTemporaryAvatarUsingMultiPart2MCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"StoreTemporaryAvatarUsingMultiPart2",
		"Create temporary avatar using multipart upload - Creates temporary avatar",
		[]byte(StoreTemporaryAvatarUsingMultiPart2InputSchema),
	)
}

// StoreTemporaryAvatarUsingMultiPart2Handler is the handler function for the StoreTemporaryAvatarUsingMultiPart2 tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func StoreTemporaryAvatarUsingMultiPart2Handler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	// Upload: read local file and send as raw body
	localFilePath := ""
	if fp, ok := args["local_file_path"]; ok {
		if s, ok := fp.(string); ok {
			localFilePath = s
		}
	}
	if localFilePath == "" {
		return mcp.NewToolResultError("missing required argument: local_file_path"), nil
	}

	fileData, err := os.ReadFile(localFilePath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read file %s: %v", localFilePath, err)), nil
	}

	startTime := time.Now()
	upstreamURL := upstream + "/rest/api/2/universal_avatar/type/{type}/owner/{owningObjectId}/temp"

	req, err := http.NewRequestWithContext(ctx, "POST", upstreamURL, bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to create upstream request: %w", err)
	}
	req.Header.Set("Content-Type", "multipart/form-data")

	if forwarded := mcputils.GetHTTPHeaders(ctx); forwarded != nil {
		for key, values := range forwarded {
			lowerKey := strings.ToLower(key)
			if lowerKey == "host" || lowerKey == "connection" || lowerKey == "keep-alive" || lowerKey == "proxy-authenticate" || lowerKey == "proxy-authorization" || lowerKey == "te" || lowerKey == "trailer" || lowerKey == "transfer-encoding" || lowerKey == "upgrade" || lowerKey == "authorization" || lowerKey == "cookie" || lowerKey == "content-length" {
				continue
			}
			for _, v := range values {
				req.Header.Add(key, v)
			}
		}
	}
	if req.Header.Get("Authorization") == "" {
		if token := mcputils.GetUpstreamToken(); token != "" {
			req.Header.Set("Authorization", mcputils.FormatAuthorizationHeader(token))
		}
	}

	if cookie := mcputils.GetUpstreamCookie(); cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	mcputils.LogRequest("POST", upstreamURL, nil, req.Header, nil)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "StoreTemporaryAvatarUsingMultiPart2"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
