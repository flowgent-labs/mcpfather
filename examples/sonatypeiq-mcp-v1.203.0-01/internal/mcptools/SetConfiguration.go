package mcptools

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"io"
	"sonatypeiq-mcp-v1.203.0-01/internal/helpers"
	"time"
)

// Input Schema for the SetConfiguration tool
const SetConfigurationInputSchema = "{\n  \"properties\": {\n    \"body\": {\n      \"description\": \"Provide the CI integration configuration as a JSON object. The structure supports different CI systems.\",\n      \"properties\": {\n        \"advancedProperties\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"download\": {\n          \"properties\": {\n            \"iqCliUrl\": {\n              \"type\": \"string\"\n            },\n            \"iqCliVersion\": {\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"enableDebugLogging\": {\n          \"type\": \"boolean\"\n        },\n        \"failBuildOnNetworkError\": {\n          \"type\": \"boolean\"\n        },\n        \"failBuildOnPolicyWarnings\": {\n          \"type\": \"boolean\"\n        },\n        \"failBuildOnReachabilityErrors\": {\n          \"type\": \"boolean\"\n        },\n        \"failBuildOnScanningErrors\": {\n          \"type\": \"boolean\"\n        },\n        \"moduleExcludes\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"parameterPriority\": {\n          \"type\": \"string\"\n        },\n        \"proxy\": {\n          \"properties\": {\n            \"host\": {\n              \"type\": \"string\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"reachability\": {\n          \"properties\": {\n            \"failOnError\": {\n              \"type\": \"boolean\"\n            },\n            \"javaAnalysis\": {\n              \"properties\": {\n                \"enabled\": {\n                  \"type\": \"boolean\"\n                },\n                \"entrypointStrategy\": {\n                  \"type\": \"string\"\n                },\n                \"namespaces\": {\n                  \"items\": {\n                    \"type\": \"string\"\n                  },\n                  \"type\": \"array\"\n                }\n              },\n              \"type\": \"object\"\n            },\n            \"javaScriptAnalysis\": {\n              \"properties\": {\n                \"enabled\": {\n                  \"type\": \"boolean\"\n                },\n                \"jsExcludes\": {\n                  \"items\": {\n                    \"type\": \"string\"\n                  },\n                  \"type\": \"array\"\n                },\n                \"jsSources\": {\n                  \"items\": {\n                    \"type\": \"string\"\n                  },\n                  \"type\": \"array\"\n                },\n                \"nodeJsExecutable\": {\n                  \"type\": \"string\"\n                },\n                \"projectRoot\": {\n                  \"type\": \"string\"\n                }\n              },\n              \"type\": \"object\"\n            }\n          },\n          \"type\": \"object\"\n        },\n        \"resultFile\": {\n          \"type\": \"string\"\n        },\n        \"sarifFile\": {\n          \"type\": \"string\"\n        },\n        \"scanPatterns\": {\n          \"items\": {\n            \"type\": \"string\"\n          },\n          \"type\": \"array\"\n        },\n        \"unstableBuildOnPolicyWarnings\": {\n          \"type\": \"boolean\"\n        }\n      },\n      \"type\": \"object\"\n    },\n    \"ownerId\": {\n      \"description\": \"The internal ID of the owner\",\n      \"type\": \"string\"\n    },\n    \"ownerType\": {\n      \"description\": \"The owner type (application or organization)\",\n      \"enum\": [\n        \"application\",\n        \"organization\"\n      ],\n      \"pattern\": \"application|organization\",\n      \"type\": \"string\"\n    }\n  },\n  \"required\": [\n    \"ownerId\",\n    \"ownerType\"\n  ],\n  \"type\": \"object\"\n}"

// Response Template for the SetConfiguration tool (Status: 200, Content-Type: application/json)
const SetConfigurationResponseTemplate_A = "# API Response Information\n\nBelow is the response template for this API endpoint.\n\nThe template shows a possible response, including its status code and content type, to help you understand and generate correct outputs.\n\n**Status Code:** 200\n\n**Content-Type:** application/json\n\n> CI configuration was saved successfully.\n\n## Response Structure\n\n- Structure (Type: object):\n  - **moduleExcludes** (Type: array):\n    - **Items** (Type: string):\n  - **sarifFile** (Type: string):\n  - **failBuildOnNetworkError** (Type: boolean):\n  - **proxy** (Type: object):\n    - **host** (Type: string):\n  - **failBuildOnPolicyWarnings** (Type: boolean):\n  - **enableDebugLogging** (Type: boolean):\n  - **parameterPriority** (Type: string):\n  - **failBuildOnScanningErrors** (Type: boolean):\n  - **resultFile** (Type: string):\n  - **reachability** (Type: object):\n    - **failOnError** (Type: boolean):\n    - **javaAnalysis** (Type: object):\n      - **enabled** (Type: boolean):\n      - **entrypointStrategy** (Type: string):\n      - **namespaces** (Type: array):\n        - **Items** (Type: string):\n    - **javaScriptAnalysis** (Type: object):\n      - **jsExcludes** (Type: array):\n        - **Items** (Type: string):\n      - **jsSources** (Type: array):\n        - **Items** (Type: string):\n      - **nodeJsExecutable** (Type: string):\n      - **projectRoot** (Type: string):\n      - **enabled** (Type: boolean):\n  - **download** (Type: object):\n    - **iqCliUrl** (Type: string):\n    - **iqCliVersion** (Type: string):\n  - **scanPatterns** (Type: array):\n    - **Items** (Type: string):\n  - **unstableBuildOnPolicyWarnings** (Type: boolean):\n  - **failBuildOnReachabilityErrors** (Type: boolean):\n  - **advancedProperties** (Type: array):\n    - **Items** (Type: string):\n"

// NewSetConfigurationMCPTool creates the MCP Tool instance for SetConfiguration
func NewSetConfigurationMCPTool() mcp.Tool {
	return mcp.NewToolWithRawSchema(
		"SetConfiguration",
		"Use this method to create or update CI integration configuration for the specified owner.\n\nThe configuration is stored as a JSON object to support various CI systems (GitHub Actions, GitLab CI, etc.). String values must be non-empty.\n\nPermissions required: Edit IQ Elements",
		[]byte(SetConfigurationInputSchema),
	)
}

// SetConfigurationHandler is the handler function for the SetConfiguration tool.
// It reads tool arguments, forwards the request to the upstream service, and returns the response.
func SetConfigurationHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	upstream := mcputils.GetUpstreamEndpoint()

	args := request.GetArguments()
	if args == nil {
		args = make(map[string]interface{})
	}
	contentType := "application/json"
	startTime := time.Now()
	resp, err := mcputils.ForwardRequest(ctx, upstream, "PUT", "/api/v2/config/ci/{ownerType}/{ownerId}", args, []string{"ownerId", "ownerType"}, contentType)
	if err != nil {
		return nil, fmt.Errorf("upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), nil)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return mcp.NewToolResultError(fmt.Sprintf("upstream error: status %d, body: %s", resp.StatusCode, string(body))), nil
	}

	if mcputils.IsBinaryDownload(resp) {
		filePath, written, err := mcputils.SaveBinaryStream(resp, "SetConfiguration")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Saved to: %s (%d bytes)", filePath, written)), nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read upstream response: %w", err)
	}

	mcputils.LogResponse(ctx, resp.StatusCode, "PUT", resp.Request.URL.String(), time.Since(startTime), body)

	return mcp.NewToolResultText(string(body)), nil
}
