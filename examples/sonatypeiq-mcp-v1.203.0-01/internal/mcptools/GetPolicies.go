package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetPolicies tool
const GetPoliciesInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetPolicies tool (Status: 200, Content-Type: application/json)
const GetPoliciesResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> The response contains a " + "\x60" + "policies" + "\x60" + " object which contains a list of:<ul><li>" + "\x60" + "id" + "\x60" + " is the policyId. It can be used in the GET method for endpoint /api/v2/policyViolations to retrieve policy violations for the policy, and other similar operations.</li><li>" + "\x60" + "name" + "\x60" + " is the name of the policy.</li><li>" + "\x60" + "ownerType" + "\x60" + " is the ownerType.</li><li>" + "\x60" + "ownerId" + "\x60" + " is the internal id associated with the ownerType.</li><li>" + "\x60" + "threatLevel" + "\x60" + " is the threat level that is set for this policy.</li><li>" + "\x60" + "policyType" + "\x60" + " indicates the type for the policy. Values can be " + "\x60" + "Security" + "\x60" + ", " + "\x60" + "License" + "\x60" + ", " + "\x60" + "Quality" + "\x60" + " or " + "\x60" + "Other" + "\x60" + ".</li>\n\n## Response Structure\n\n- Structure (Type: object):\n  - **policies** (Type: array):\n    - **Items** (Type: object):\n      - **id** (Type: string):\n      - **name** (Type: string):\n      - **ownerId** (Type: string):\n      - **ownerType** (Type: string):\n          - Enum: ['APPLICATION', 'ORGANIZATION', 'REPOSITORY_CONTAINER', 'REPOSITORY_MANAGER', 'REPOSITORY']\n      - **policyType** (Type: string):\n      - **threatLevel** (Type: integer, int32):\n"

// NewGetPoliciesMCPTool creates the MCP Tool instance for GetPolicies
func NewGetPoliciesMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetPolicies",
		"Use this method to retrieve all existing policies.\n\nPermissions required: View IQ Elements",
		[]byte(GetPoliciesInputSchema),
	)
}

// GetPoliciesHandler is the handler function for the GetPolicies tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetPoliciesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/policies", args, []string{}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetPolicies")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "GET", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
