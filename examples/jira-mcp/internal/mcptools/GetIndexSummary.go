package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"jira-mcp/internal/helpers"
	"time"
)

// Input Schema for the GetIndexSummary tool
const GetIndexSummaryInputSchema = "{\n  \"type\": \"object\"\n}"

// Response Template for the GetIndexSummary tool (Status: 200, Content-Type: application/json)
const GetIndexSummaryResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Returns an object with data about the condition of the Jira node's index\n\n## Response Structure\n\n- Structure (Type: object):\n  - **replicationQueues** (Type: object):\n    - **Additional Properties**:\n      - **property value** (Type: object):\n        - **lastConsumedOperation** (Type: object):\n          - **id** (Type: integer, int64):\n              - Example: '16822'\n          - **replicationTime** (Type: string, date-time):\n              - Example: '2017-07-08T00:49:07.842Z'\n        - **[cyclic reference]**\n        - **queueSize** (Type: integer, int64):\n            - Example: '0'\n  - **reportTime** (Type: string, date-time):\n      - Example: '2017-07-08T01:46:16.94Z'\n  - **issueIndex** (Type: object):\n    - **countInIndex** (Type: integer, int64):\n        - Example: '10072'\n    - **indexReadable** (Type: boolean):\n        - Example: 'true'\n    - **lastUpdatedInDatabase** (Type: string, date-time):\n        - Example: '2017-07-08T01:46:16.94Z'\n    - **lastUpdatedInIndex** (Type: string, date-time):\n        - Example: '2017-07-08T00:48:53Z'\n    - **countInArchive** (Type: integer, int64):\n        - Example: '2000'\n    - **countInDatabase** (Type: integer, int64):\n        - Example: '12072'\n  - **nodeId** (Type: string):\n      - Example: 'node1'\n"

// NewGetIndexSummaryMCPTool creates the MCP Tool instance for GetIndexSummary
func NewGetIndexSummaryMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"GetIndexSummary",
		"Get index condition summary - Returns a summary of the index condition of the current node.\nThe returned data consists of:\n- "+"\x60"+"nodeId"+"\x60"+" - Node identifier.\n- "+"\x60"+"reportTime"+"\x60"+" - Time of this report creation.\n- "+"\x60"+"issueIndex"+"\x60"+" - Summary of the issue index status.\n- "+"\x60"+"replicationQueues"+"\x60"+" - Map of index replication queues, where keys represent nodes from which replication operations came from.\n\n"+"\x60"+"issueIndex"+"\x60"+" can contain:\n    - "+"\x60"+"indexReadable"+"\x60"+" - If "+"\x60"+"false"+"\x60"+" the endpoint failed to read data from the issue index (check Jira logs for detailed stack trace), otherwise "+"\x60"+"true"+"\x60"+".\n    - "+"\x60"+"countInDatabase"+"\x60"+" - Count of issues found in the database.\n    - "+"\x60"+"countInIndex"+"\x60"+" - Count of issues found while querying the index.\n    - "+"\x60"+"lastUpdatedInDatabase"+"\x60"+" - Time of the last update of the issue found in the database.\n    - "+"\x60"+"lastUpdatedInIndex"+"\x60"+" - Time of the last update of the issue found while querying the index.\n"+"\x60"+"replicationQueues"+"\x60"+"'s map values can contain:\n    - "+"\x60"+"lastConsumedOperation"+"\x60"+" - Last executed index replication operation by the current node from the sending node's queue.\n    - "+"\x60"+"lastConsumedOperation.id"+"\x60"+" - Identifier of the operation.\n    - "+"\x60"+"lastConsumedOperation.replicationTime"+"\x60"+" - Time when the operation was sent to other nodes.\n    - "+"\x60"+"lastOperationInQueue"+"\x60"+" - Last index replication operation in the sending node's queue.\n    - "+"\x60"+"lastOperationInQueue.id"+"\x60"+" - Identifier of the operation.\n    - "+"\x60"+"lastOperationInQueue.replicationTime"+"\x60"+" - Time when the operation was sent to other nodes.\n    - "+"\x60"+"queueSize"+"\x60"+" - Number of operations in the queue from the sending node to the current node.",
		[]byte(GetIndexSummaryInputSchema),
	)
}

// GetIndexSummaryHandler is the handler function for the GetIndexSummary tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func GetIndexSummaryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/rest/api/2/index/summary", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "GetIndexSummary")
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
