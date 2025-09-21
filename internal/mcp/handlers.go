package mcp

import (
	"encoding/json"
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

	// MCP Info endpoint
	router.GET("/", s.handleMCPInfo)
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

// Legacy MCP Server Routes (for backward compatibility)
func (s *MCPServer) HandleListTools(c *gin.Context) {
	tools := s.ListTools()
	c.JSON(http.StatusOK, gin.H{
		"tools": tools,
	})
}

func (s *MCPServer) HandleToolCall(c *gin.Context) {
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

func (s *MCPServer) RegisterRoutes(router *gin.RouterGroup) {
	mcp := router.Group("/mcp")
	{
		mcp.GET("/tools", s.HandleListTools)
		mcp.POST("/tools/call", s.HandleToolCall)
	}
}