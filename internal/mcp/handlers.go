package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Enhanced MCP Server Routes
func (s *EnhancedMCPServer) RegisterRoutes(router *gin.RouterGroup) {
	// MCP Tool endpoints
	router.GET("/tools", s.handleListTools)
	router.POST("/tools/call", s.handleExecuteTool)
	router.POST("/tools/execute", s.handleExecuteTool)

	// MCP Resource endpoints
	router.GET("/resources", s.handleListResources)
	router.GET("/resources/get", s.handleGetResource)
	router.POST("/resources/subscribe", s.handleSubscribeResource)

	// MCP Info endpoint - handle both with and without trailing slash
	router.GET("", s.handleMCPInfo)
	router.GET("/", s.handleMCPInfo)

	// JSON-RPC endpoint for MCP protocol
	router.POST("", s.handleJSONRPC)
	router.POST("/", s.handleJSONRPC)
}

// JSON-RPC 2.0 structures
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      interface{}     `json:"id,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (s *EnhancedMCPServer) handleJSONRPC(c *gin.Context) {
	var request JSONRPCRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32700,
				Message: "Parse error",
			},
			ID: nil,
		})
		return
	}

	// Handle different MCP methods
	var result interface{}
	var rpcErr *JSONRPCError

	switch request.Method {
	case "initialize":
		result = gin.H{
			"protocolVersion": "2024-11-05",
			"serverInfo": gin.H{
				"name":    "Headless PM MCP Server",
				"version": "2.0.0",
			},
			"capabilities": gin.H{
				"tools": gin.H{
					"listChanged": true,
				},
				"resources": gin.H{
					"subscribe": true,
					"listChanged": true,
				},
			},
		}

	case "tools/list":
		tools := s.ListTools()
		result = gin.H{
			"tools": tools,
		}

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			rpcErr = &JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
			}
		} else {
			args, _ := json.Marshal(params.Arguments)
			call := ToolCall{
				Name:      params.Name,
				Arguments: args,
			}
			toolResult, err := s.ExecuteTool(c.Request.Context(), call)
			if err != nil {
				rpcErr = &JSONRPCError{
					Code:    -32603,
					Message: fmt.Sprintf("Tool execution error: %v", err),
				}
			} else if toolResult.IsError {
				rpcErr = &JSONRPCError{
					Code:    -32603,
					Message: fmt.Sprintf("%v", toolResult.Content),
				}
			} else {
				// Convert content to string
				var contentStr string
				switch v := toolResult.Content.(type) {
				case string:
					contentStr = v
				case []byte:
					contentStr = string(v)
				default:
					// Marshal to JSON if not a string
					if bytes, err := json.Marshal(v); err == nil {
						contentStr = string(bytes)
					} else {
						contentStr = fmt.Sprintf("%v", v)
					}
				}

				result = gin.H{
					"content": []gin.H{
						{
							"type": "text",
							"text": contentStr,
						},
					},
				}
			}
		}

	case "resources/list":
		resources := s.ListResources()
		result = gin.H{
			"resources": resources,
		}

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(request.Params, &params); err != nil {
			rpcErr = &JSONRPCError{
				Code:    -32602,
				Message: "Invalid params",
			}
		} else {
			content, err := s.GetResource(c.Request.Context(), params.URI)
			if err != nil {
				rpcErr = &JSONRPCError{
					Code:    -32603,
					Message: fmt.Sprintf("Resource error: %v", err),
				}
			} else {
				result = gin.H{
					"contents": []interface{}{content},
				}
			}
		}

	default:
		rpcErr = &JSONRPCError{
			Code:    -32601,
			Message: "Method not found",
		}
	}

	// Send response
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
	}

	if rpcErr != nil {
		response.Error = rpcErr
	} else {
		response.Result = result
	}

	c.JSON(http.StatusOK, response)
}

func (s *EnhancedMCPServer) handleMCPInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"name":        "Headless PM MCP Server",
		"version":     "2.0.0",
		"description": "Enhanced MCP server with comprehensive project management capabilities",
		"capabilities": gin.H{
			"tools":     len(s.ListTools()),
			"resources": len(s.ListResources()),
			"features": []string{
				"project_management",
				"task_tracking",
				"team_collaboration",
				"sprint_management",
				"milestone_tracking",
				"workflow_automation",
				"semantic_search",
				"ai_recommendations",
				"analytics",
				"webhooks",
				"notifications",
				"time_tracking",
			},
		},
	})
}

func (s *EnhancedMCPServer) handleListTools(c *gin.Context) {
	tools := s.ListTools()
	c.JSON(http.StatusOK, gin.H{
		"tools": tools,
	})
}

func (s *EnhancedMCPServer) handleExecuteTool(c *gin.Context) {
	var request struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	args, err := json.Marshal(request.Arguments)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid arguments",
		})
		return
	}

	call := ToolCall{
		Name:      request.Name,
		Arguments: args,
	}

	result, err := s.ExecuteTool(c.Request.Context(), call)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if result.IsError {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": result.Content,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": result.Content,
	})
}

func (s *EnhancedMCPServer) handleListResources(c *gin.Context) {
	resources := s.ListResources()
	c.JSON(http.StatusOK, gin.H{
		"resources": resources,
	})
}

func (s *EnhancedMCPServer) handleGetResource(c *gin.Context) {
	uri := c.Query("uri")
	if uri == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "uri parameter is required",
		})
		return
	}

	content, err := s.GetResource(c.Request.Context(), uri)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, content)
}

func (s *EnhancedMCPServer) handleSubscribeResource(c *gin.Context) {
	var req struct {
		URI string `json:"uri"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// For HTTP-based subscriptions, we'll return the current state
	// In a production system, this could use WebSockets or Server-Sent Events
	content, err := s.GetResource(c.Request.Context(), req.URI)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribed": true,
		"uri":        req.URI,
		"initial":    content,
	})
}