package mcptools

import (
	"confluence-mcp-v10.2.14/internal/helpers"
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"time"
)

// Input Schema for the UpdateRestrictions tool
const UpdateRestrictionsInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"items\": {},\n      \"type\": \"array\"\n    },\n    \"expand\": {\n      \"description\": \"A comma separated list of properties to expand in the response. Default is \\u003ccode\\u003erestrictions.user, restrictions.group\\u003c/code\\u003e Default value: group.\",\n      \"type\": \"string\"\n    },\n    \"id\": {\n      \"type\": \"string\"\n    },\n    \"limit\": {\n      \"description\": \"pagination limit.\",\n      \"type\": \"string\"\n    },\n    \"start\": {\n      \"description\": \"pagination start.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"id\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the UpdateRestrictions tool (Status: 200, Content-Type: application/json)
const UpdateRestrictionsResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns a full JSON representation of the restrictions present directly on piece of content after the update operation.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **_links** (Type: object):\n    - **base** (Type: string):\n        - Example: 'http://localhost:8085/confluence'\n    - **context** (Type: string):\n        - Example: 'confluence'\n    - **next** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=50'\n    - **prev** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=0'\n    - **self** (Type: string):\n        - Example: 'http://localhost:8085/rest/api/latest/..?limit=25&start=25'\n  - **limit** (Type: number):\n      - Example: '25'\n  - **results** (Type: array):\n    - **Items** (Type: unknown):\n  - **size** (Type: number):\n      - Example: '25'\n  - **start** (Type: number):\n      - Example: '25'\n  - **totalCount** (Type: integer, int64):\n"

// Response Template for the UpdateRestrictions tool (Status: 400, Content-Type: application/json)
const UpdateRestrictionsResponseTemplate_B = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 400\n\n**Content-Type:** application/json\n\n> Returned if any of the above validation rules are violated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateRestrictions tool (Status: 401, Content-Type: application/json)
const UpdateRestrictionsResponseTemplate_C = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 401\n\n**Content-Type:** application/json\n\n> Returned if the calling user is not authenticated.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateRestrictions tool (Status: 403, Content-Type: application/json)
const UpdateRestrictionsResponseTemplate_D = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 403\n\n**Content-Type:** application/json\n\n> Returned if the calling user does not have permission to edit the restrictions.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// Response Template for the UpdateRestrictions tool (Status: 404, Content-Type: application/json)
const UpdateRestrictionsResponseTemplate_E = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 404\n\n**Content-Type:** application/json\n\n> Returned if there is no content with the given id, or if the calling user does not have permission to view the content.\n\n## Response Structure\n\n- Structure (Type: unknown):\n"

// NewUpdateRestrictionsMCPTool creates the MCP Tool instance for UpdateRestrictions
func NewUpdateRestrictionsMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"UpdateRestrictions",
		"Update restrictions - Sets all the restrictions specified to a piece of content identified by "+"\x60"+"contentId"+"\x60"+". Setting per-content restrictions is currently allowed for Pages or BlogPosts only. \n\nExample request URI: \n\n"+"\x60"+"http://example.com/confluence/rest/content/1234567/restriction?expand="+"\x60"+"\n\nThe payload uses the same schema as returned by the GET requests from "+"\x60"+"/rest/api/content/{id}/restriction/byOperation*"+"\x60"+" which can be used as a template but is not necessary. \n\nExample request for a single content restriction: \n\n"+"\x60"+""+"\x60"+""+"\x60"+"json\n[ { \"operation\": \"read\", \"restrictions\": { \"user\": [ { \"type\": \"known\", \"username\": \"admin\" } ] } }, { \"operation\": \"update\", \"restrictions\": { \"user\": [ { \"type\": \"known\", \"username\": \"admin\" } ] } } ]\n"+"\x60"+""+"\x60"+""+"\x60"+"\n\nExample request for updating two ContentRestrictions: \n\n"+"\x60"+""+"\x60"+""+"\x60"+"json\n[ { \"operation\": \"update\", \"restrictions\": { \"user\": [ { \"type\": \"known\", \"username\": \"admin\" } ] } }, { \"operation\": \"read\", \"restrictions\": { \"user\": [ { \"type\": \"known\", \"username\": \"fred\" }, { \"type\": \"known\", \"username\": \"admin\" } ] } } ]\n"+"\x60"+""+"\x60"+""+"\x60"+"\n\nRules for using this method: \n\n- The provided ContentRestrictions will overwrite any existing restrictions on the Content for the corresponding operations. \n- If the provided "+"\x60"+"ContentRestriction"+"\x60"+" lacks any supported operations, the restrictions for the operations will not be altered. \n- Setting "+"\x60"+"users"+"\x60"+" and/or "+"\x60"+"groups"+"\x60"+" map entries as empty arrays will remove the corresponding content restrictions. \n- Missing "+"\x60"+"users"+"\x60"+" and/or "+"\x60"+"groups"+"\x60"+" map entries means the corresponding operation's user/group content restrictions won't be changed. \n- Modifying restrictions to revoke the requesting user's access is prohibited. \n- 'update' restrictions requires 'read' restrictions for same user/group.",
		[]byte(UpdateRestrictionsInputSchema),
	)
}

// UpdateRestrictionsHandler is the handler function for the UpdateRestrictions tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func UpdateRestrictionsHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/confluence/rest/api/content/{id}/restriction", args, []string{"id"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if filePath, err := mcputils.SaveBinaryResponse(resp, body, "UpdateRestrictions"); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	} else if filePath != "" {
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, len(body))), nil
	}

	return mcp.NewToolResultText(string(body)), nil
}
