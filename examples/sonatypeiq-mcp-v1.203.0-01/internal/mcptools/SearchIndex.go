package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SearchIndex tool
const SearchIndexInputSchema = "{\n  \"properties\": {\n    \"allComponents\": {\n      \"default\": false,\n      \"description\": \"Set to " + "\x60" + "true" + "\x60" + " to retrieve results that include components with no violations\",\n      \"type\": \"boolean\"\n    },\n    \"mode\": {\n      \"enum\": [\n        \"sbomManager\"\n      ],\n      \"type\": \"string\"\n    },\n    \"page\": {\n      \"description\": \"Enter the page no. for the page containing results\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"pageSize\": {\n      \"default\": 10,\n      \"description\": \"Enter the no. of results that should be visible per page\",\n      \"format\": \"int32\",\n      \"type\": \"integer\"\n    },\n    \"query\": {\n      \"description\": \"Enter your search query here\",\n      \"type\": \"string\"\n    }\n  },\n  \"type\": \"object\"\n}"

// Response Template for the SearchIndex tool (Status: 200, Content-Type: application/json)
const SearchIndexResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> Response JSON containing the search query sent in the API call, and other response fields as follows: \n1. searchQuery: search query sent in the request \n2. page: page number of search results requested \n3. pageSize: requested number of results per page \n4. totalNumberOfHits: total number of results returned \n5. isExactTotalNumberOfHits \n    * " + "\x60" + "true" + "\x60" + " indicates that the search results in the JSON is the same no. of search results that logically      match the search query. \n    * " + "\x60" + "false" + "\x60" + " indicates that the search results in the JSON are lower bound because fetching all results is     too expensive to compute. \n6. groupingByDTOS: array of search results grouped on a field name \n7. groupIdentifier: field name that the search results have been grouped by \n8. groupBy: field value that the search results have been grouped by \n9. additionalInfo: shared information between groups, e.g. info if grouped by a security vulnerability \n10. searchResultItemDTOS: array of search results with each element containing an itemType, field names and values \n11. resultIndex: indicating the relevance of the search result w.r.t. the query\n\n## Response Structure\n\n- Structure (Type: object):\n  - **page** (Type: integer, int32):\n  - **pageSize** (Type: integer, int32):\n  - **searchAfter** (Type: array):\n    - **Items** (Type: string):\n  - **searchQuery** (Type: string):\n  - **totalNumberOfHits** (Type: integer, int64):\n  - **groupingByDTOS** (Type: array):\n    - **Items** (Type: object):\n      - **groupIdentifier** (Type: string):\n          - Enum: ['itemType', 'organizationId', 'organizationName', 'applicationId', 'applicationName', 'applicationPublicId', 'policyEvaluationStage', 'applicationVersion', 'reportId', 'componentHash', 'componentFormat', 'componentName', 'componentCoordinate', 'vulnerabilityId', 'vulnerabilitySeverity', 'vulnerabilityStatus', 'vulnerabilityDescription', 'applicationCategoryId', 'applicationCategoryName', 'applicationCategoryColor', 'applicationCategoryDescription', 'componentLabelId', 'componentLabelName', 'componentLabelColor', 'componentLabelDescription', 'policyId', 'policyName', 'policyThreatCategory', 'policyThreatLevel', 'parentOrganizationName', 'parentOrganizationId', 'sbomSpecification']\n      - **searchResultItemDTOS** (Type: array):\n        - **Items** (Type: object):\n          - **applicationCategoryId** (Type: string):\n          - **vulnerabilityId** (Type: string):\n          - **applicationVersion** (Type: string):\n          - **componentLabelId** (Type: string):\n          - **organizationName** (Type: string):\n          - **applicationCategoryName** (Type: string):\n          - **applicationPublicId** (Type: string):\n          - **componentLabelDescription** (Type: string):\n          - **organizationId** (Type: string):\n          - **itemType** (Type: string):\n          - **vulnerabilityStatus** (Type: string):\n          - **policyName** (Type: string):\n          - **policyThreatCategory** (Type: string):\n          - **applicationCategoryColor** (Type: string):\n          - **applicationId** (Type: string):\n          - **applicationName** (Type: string):\n          - **resultIndex** (Type: integer, int32):\n          - **policyThreatLevel** (Type: integer, int32):\n          - **componentLabelColor** (Type: string):\n          - **componentHash** (Type: string):\n          - **componentName** (Type: string):\n          - **reportId** (Type: string):\n          - **policyId** (Type: string):\n          - **vulnerabilityDescription** (Type: string):\n          - **policyEvaluationStage** (Type: string):\n          - **applicationCategoryDescription** (Type: string):\n          - **componentLabelName** (Type: string):\n          - **componentIdentifier** (Type: object):\n            - **coordinates** (Type: object):\n              - **Additional Properties**:\n                - **property value** (Type: string):\n            - **format** (Type: string):\n          - **sbomSpecification** (Type: string):\n      - **additionalInfo** (Type: string):\n      - **groupBy** (Type: string):\n  - **isExactTotalNumberOfHits** (Type: boolean):\n"

// NewSearchIndexMCPTool creates the MCP Tool instance for SearchIndex
func NewSearchIndexMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SearchIndex",
		"Use this method to perform an Advanced Search. ",
		[]byte(SearchIndexInputSchema),
	)
}

// SearchIndexHandler is the handler function for the SearchIndex tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SearchIndexHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := ""
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "GET", "/api/v2/search/advanced", args, []string{}, contentType)
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
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SearchIndex")
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
