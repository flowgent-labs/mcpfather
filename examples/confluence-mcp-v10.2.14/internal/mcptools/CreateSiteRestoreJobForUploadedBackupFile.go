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

// Input Schema for the CreateSiteRestoreJobForUploadedBackupFile tool
const CreateSiteRestoreJobForUploadedBackupFileInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Backup file to be uploaded. Has to be a zip file.\",\n      \"properties\": {\n        \"file\": {\n          \"description\": \"backup file uploaded. Has to be a zip file.\",\n          \"type\": \"string\"\n        }\n      },\n      \"type\": \"object\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the CreateSiteRestoreJobForUploadedBackupFile tool (Status: 200, Content-Type: application/json)
const CreateSiteRestoreJobForUploadedBackupFileResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a JSON representation of the site restore job details.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateSiteRestoreJobForUploadedBackupFile tool (Status: 400, Content-Type: application/json)
const CreateSiteRestoreJobForUploadedBackupFileResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n>  Returned if the uploaded file is not a zip file\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the CreateSiteRestoreJobForUploadedBackupFile tool (Status: 403, Content-Type: application/json)
const CreateSiteRestoreJobForUploadedBackupFileResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n>  Returned if user doesn't have permission to restore space\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewCreateSiteRestoreJobForUploadedBackupFileMCPTool creates the MCP Tool instance for CreateSiteRestoreJobForUploadedBackupFile
func NewCreateSiteRestoreJobForUploadedBackupFileMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"CreateSiteRestoreJobForUploadedBackupFile",
		"Create site restore job for upload backup file - This resource expects a multipart post. The media-type multipart/form-data is defined in RFC 1867. \n\nMost client libraries have classes that make dealing with multipart posts simple. \n\nFor instance, in Java the Apache HTTP Components library provides a MultiPartEntity that makes it simple to submit a multipart POST. \n\n In order to protect against XSRF attacks, because this method accepts multipart/form-data, it has XSRF protection on it.  This means you must submit a header of X-Atlassian-Token: nocheck with the request, otherwise it will be blocked. \n\n The name of the multipart/form-data parameter that contains attachments must be \"file\". \n\n An example to attach the file: \n\n curl -D- -u admin:admin -X POST -H \"X-Atlassian-Token: nocheck\" -F  file=@myfile.zip http://myhost/rest/api/backup-restore/restore/space/upload \n\n.",
		[]byte(CreateSiteRestoreJobForUploadedBackupFileInputSchema),
	)
}

// CreateSiteRestoreJobForUploadedBackupFileHandler is the handler function for the CreateSiteRestoreJobForUploadedBackupFile tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func CreateSiteRestoreJobForUploadedBackupFileHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	upstreamURL := upstream + "/confluence/rest/api/backup-restore/restore/site/upload"

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

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "CreateSiteRestoreJobForUploadedBackupFile"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
