package mcptools

import (
	"bytes"
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Input Schema for the UpdateData tool
const UpdateDataInputSchema = "{\n  \"properties\": {\n    \"attachmentId\": {\n      \"description\": \"the id of the attachment to upload the new file for.\",\n      \"type\": \"string\"\n    },\n    \"body\": {\n      \"description\": \"The attachment to be updated.\",\n      \"properties\": {\n        \"comment\": {\n          \"description\": \"(optional) a list of \\\\\\\"comments\\\\\\\" matching the list of attachment data.\\\\nIf supplied, the size of this list must match the size of the fileParts list.\",\n          \"type\": \"string\"\n        },\n        \"file\": {\n          \"description\": \"The name of the multipart/form-data parameter that contains attachments must be \\\\\\\"file\\\\\\\".\",\n          \"type\": \"string\"\n        },\n        \"hidden\": {\n          \"description\": \"(optional) form parameter indicating whether the attachments should be \\\\\\\"hidden\\\\\\\".If \\\\\\\"hidden\\\\\\\" is set to true, no notification email or activity stream will be generated for that attachment.\",\n          \"type\": \"boolean\"\n        },\n        \"minorEdit\": {\n          \"description\": \"(optional) form parameter indicating whether the attachments should be \\\\\\\"minorEdits\\\\\\\".If \\\\\\\"minorEdits\\\\\\\" is set to true, no notification email will be generated for that attachment.\",\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"id\": {\n      \"description\": \"The id of the content the attachment is on.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"attachmentId\",\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateData tool (Status: 200, Content-Type: application/json)
const UpdateDataResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> returns JSON representation of the updated attachment.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateData tool (Status: 400, Content-Type: application/json)
const UpdateDataResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n>  Returned if the attachment id is invalid.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateData tool (Status: 404, Content-Type: application/json)
const UpdateDataResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n>  Returned if no attachment is found for the attachmentId.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdateDataMCPTool creates the MCP Tool instance for UpdateData
func NewUpdateDataMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateData",
		"Update binary data of an attachment - Update the binary data of an Attachment, and optionally the comment and the minor edit field.\n\nThis adds a new version of the attachment, containing the new binary data, filename, and content-type.\n\n**When updating the binary data of an attachment**, the comment related to it together with the field that specifies if it's a minor edit can be updated as well, but are not required.\n\nIf an update is considered to be a minor edit, notifications will not be sent to the watchers of that content.\n\nThis resource expects a multipart post. The media-type multipart/form-data is defined in RFC 1867. Most client libraries have classes that make dealing with multipart posts simple. For instance, in Java the Apache HTTP Components library provides a "+"\x60"+"MultiPartEntity"+"\x60"+" that makes it simple to submit a multipart POST.\n\nIn order to protect against XSRF attacks, because this method accepts multipart/form-data, it has XSRF protection on it. This means you must submit a header of "+"\x60"+"X-Atlassian-Token: nocheck"+"\x60"+" with the request, otherwise it will be blocked.\n\nThe name of the multipart/form-data parameter that contains attachments must be 'file'.",
		[]byte(UpdateDataInputSchema),
	)
}

// UpdateDataHandler is the handler function for the UpdateData tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateDataHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	upstreamURL := upstream + "/confluence/rest/api/content/{id}/child/attachment/{attachmentId}/data"

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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateData"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
