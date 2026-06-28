package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the ValidateUserAnonymizationRerun tool
const ValidateUserAnonymizationRerunInputSchema = "{\n  \"properties\": {\n    \"expand\": {\n      \"description\": \"Parameter used to include parts of the response.\",\n      \"type\": \"string\"\n    },\n    \"oldUserKey\": {\n      \"description\": \"User key before anonymization, only needed when current value is anonymized. If there is no old key, e.g. because the user was already created using the new key generation strategy, provide a value equal to the current key.\",\n      \"type\": \"string\"\n    },\n    \"oldUserName\": {\n      \"description\": \"User name before anonymization, only needed when the current value is anonymized. If there is no old name, provide a value equal to the current name.\",\n      \"type\": \"string\"\n    },\n    \"userKey\": {\n      \"description\": \"The key of the user to validate anonymization for.\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the ValidateUserAnonymizationRerun tool (Status: 200, Content-Type: application/json)
const ValidateUserAnonymizationRerunResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returned when validation succeeded.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **operations** (Type: array):\n      - Unique Items: true\n      - Example: '[\"USER_TRANSFER_OWNERSHIP_PLUGIN_POINTS\",\"USER_DISABLE\",\"USER_KEY_CHANGE_PLUGIN_POINTS\",\"USER_KEY_CHANGE\",\"USER_NAME_CHANGE_PLUGIN_POINTS\",\"USER_NAME_CHANGE\",\"USER_EXTERNAL_ID_CHANGE\",\"USER_ANONYMIZE_PLUGIN_POINTS\"]'\n    - **Items** (Type: string):\n        - Example: '[\"USER_TRANSFER_OWNERSHIP_PLUGIN_POINTS\",\"USER_DISABLE\",\"USER_KEY_CHANGE_PLUGIN_POINTS\",\"USER_KEY_CHANGE\",\"USER_NAME_CHANGE_PLUGIN_POINTS\",\"USER_NAME_CHANGE\",\"USER_EXTERNAL_ID_CHANGE\",\"USER_ANONYMIZE_PLUGIN_POINTS\"]'\n  - **email** (Type: string):\n      - Example: 'fred@example.com'\n  - **errors** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: object):\n        - **errors** (Type: object):\n          - **Additional Properties**:\n            - **property value** (Type: string):\n        - **errorMessages** (Type: array):\n          - **Items** (Type: string):\n  - **userKey** (Type: string):\n      - Example: 'JIRAUSER10100'\n  - **expand** (Type: string):\n  - **success** (Type: boolean):\n      - Example: 'true'\n  - **businessLogicValidationFailed** (Type: boolean):\n      - Example: 'true'\n  - **userName** (Type: string):\n      - Example: 'fred'\n  - **warnings** (Type: object):\n    - **Additional Properties**:\n      - **[cyclic reference]**\n  - **affectedEntities** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: array):\n        - **Items** (Type: object):\n          - **type** (Type: string):\n              - Example: 'ANONYMIZE'\n              - Enum: ['ANONYMIZE', 'TRANSFER_OWNERSHIP', 'REMOVE', 'MANUAL']\n          - **uri** (Type: string):\n              - Example: '/jira/secure/ViewProfile.jspa?name=fred'\n          - **uriDisplayName** (Type: string):\n              - Example: 'User Profile'\n          - **description** (Type: string):\n              - Example: 'User Profile'\n          - **numberOfOccurrences** (Type: integer, int64):\n              - Example: '1'\n  - **deleted** (Type: boolean):\n      - Example: 'false'\n  - **displayName** (Type: string):\n      - Example: 'Fred Flinston'\n"

// NewValidateUserAnonymizationRerunMCPTool creates the MCP Tool instance for ValidateUserAnonymizationRerun
func NewValidateUserAnonymizationRerunMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"ValidateUserAnonymizationRerun",
		"Get validation for user anonymization rerun - Validates user anonymization re-run process.",
		[]byte(ValidateUserAnonymizationRerunInputSchema),
	)
}

// ValidateUserAnonymizationRerunHandler is the handler function for the ValidateUserAnonymizationRerun tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func ValidateUserAnonymizationRerunHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/user/anonymization/rerun", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "ValidateUserAnonymizationRerun")
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
