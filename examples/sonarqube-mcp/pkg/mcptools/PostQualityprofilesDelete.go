package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonarqube-mcp/pkg/helpers"
	"time"
)

// Input Schema for the PostQualityprofilesDelete tool
const PostQualityprofilesDeleteInputSchema = "{\n  \"properties\": {\n    \"language\": {\n      \"description\": \"Quality profile language.\",\n      \"enum\": [\n        \"kubernetes\",\n        \"css\",\n        \"py\",\n        \"plsql\",\n        \"apex\",\n        \"java\",\n        \"xml\",\n        \"powershell\",\n        \"ipynb\",\n        \"text\",\n        \"azurepipelines\",\n        \"swift\",\n        \"yaml\",\n        \"cpp\",\n        \"vb\",\n        \"secrets\",\n        \"groovy\",\n        \"shell\",\n        \"terraform\",\n        \"azureresourcemanager\",\n        \"abap\",\n        \"scala\",\n        \"jsp\",\n        \"js\",\n        \"sql\",\n        \"docker\",\n        \"ansible\",\n        \"dart\",\n        \"rust\",\n        \"jcl\",\n        \"web\",\n        \"flex\",\n        \"json\",\n        \"vbnet\",\n        \"cloudformation\",\n        \"c\",\n        \"go\",\n        \"kotlin\",\n        \"rpg\",\n        \"pli\",\n        \"tsql\",\n        \"ruby\",\n        \"cs\",\n        \"cobol\",\n        \"php\",\n        \"gosu\",\n        \"objc\",\n        \"ts\",\n        \"githubactions\"\n      ],\n      \"type\": \"string\"\n    },\n    \"qualityProfile\": {\n      \"description\": \"Quality profile name.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"language\",\n    \"qualityProfile\"\n  ],\n  \"type\": \"object\"\n}"

// NewPostQualityprofilesDeleteMCPTool creates the MCP Tool instance for PostQualityprofilesDelete
func NewPostQualityprofilesDeleteMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"PostQualityprofilesDelete",
		"Delete a quality profile and all its descendants. - Delete a quality profile and all its descendants. The default quality profile cannot be deleted. Requires one of the following permissions: - 'Administer Quality Profiles' - Edit right on the specified quality profile",
		[]byte(PostQualityprofilesDeleteInputSchema),
	)
}

// PostQualityprofilesDeleteHandler is the handler function for the PostQualityprofilesDelete tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func PostQualityprofilesDeleteHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/qualityprofiles/delete", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "PostQualityprofilesDelete")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "POST", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
