package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the GetSuggestedRemediationForComponent tool
const GetSuggestedRemediationForComponentInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"properties\": {\n        \"componentIdentifier\": {\n          \"properties\": {\n            \"coordinates\": {\n              \"additionalProperties\": {\n                \"type\": \"string\"\n              },\n              \"type\": \"object\"\n            },\n            \"format\": {\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"displayName\": {\n          \"type\": \"string\"\n        },\n        \"hash\": {\n          \"type\": \"string\"\n        },\n        \"originalPurl\": {\n          \"type\": \"string\"\n        },\n        \"packageUrl\": {\n          \"type\": \"string\"\n        },\n        \"proprietary\": {\n          \"type\": \"boolean\"\n        },\n        \"sha256\": {\n          \"type\": \"string\"\n        },\n        \"thirdParty\": {\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"identificationSource\": {\n      \"description\": \"Enter the identification source if you want the remediation result based on third-party scan information (non-Sonatype). The identification source can be obtained from the Component Details Page in the UI.\",\n      \"type\": \"string\"\n    },\n    \"includeParentRemediation\": {\n      \"default\": false,\n      \"description\": \"Enter true if you want to include parent remediation for transitive dependency in the response based on your application policy scan.\",\n      \"type\": \"boolean\"\n    },\n    \"ownerId\": {\n      \"description\": \"Possible values: applicationId, organizationId or repositoryId.\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"Possible values: application, organization, repository. \",\n      \"enum\": [\n        \"application\",\n        \"organization\",\n        \"repository\"\n      ],\n      \"pattern\": \"application|organization|repository\",\n      \"type\": \"string\"\n    },\n    \"scanId\": {\n      \"description\": \"Enter the scanId (reportId) if you want the remediation result based on third-party scan information (non-Sonatype).\",\n      \"type\": \"string\"\n    },\n    \"stageId\": {\n      \"description\": \"Enter the stageId to obtain next-non-failing and next-non-failing-with-dependencies remediation types in the response. Possible values are develop, build, stage-release, release and operate.\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// NewGetSuggestedRemediationForComponentMCPTool creates the MCP Tool instance for GetSuggestedRemediationForComponent
func NewGetSuggestedRemediationForComponentMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetSuggestedRemediationForComponent",
		"Use this method to obtain remediation suggestions for policy violations on a component basis. Remediations obtained from this method are same as those appearing on the Component Details Page in the UI.",
		[]byte(GetSuggestedRemediationForComponentInputSchema),
	)
}

// GetSuggestedRemediationForComponentHandler is the handler function for the GetSuggestedRemediationForComponent tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetSuggestedRemediationForComponentHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "POST", "/api/v2/components/remediation/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetSuggestedRemediationForComponent")
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
