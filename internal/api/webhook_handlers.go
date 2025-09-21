package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/headless-pm/headless-project-management/internal/database"
	"github.com/headless-pm/headless-project-management/internal/models"
)

type WebhookHandler struct {
	db *database.Database
}

func NewWebhookHandler(db *database.Database) *WebhookHandler {
	return &WebhookHandler{db: db}
}

func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	var webhook models.Webhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook.Secret = h.generateSecret()
	webhook.IsActive = true

	if err := h.db.Create(&webhook).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook"})
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	projectID := c.Query("project_id")

	var webhooks []models.Webhook
	query := h.db.DB

	if projectID != "" {
		query = query.Where("project_id = ?", projectID)
	}

	if err := query.Find(&webhooks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list webhooks"})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}

func (h *WebhookHandler) GetWebhook(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var webhook models.Webhook
	if err := h.db.First(&webhook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

func (h *WebhookHandler) UpdateWebhook(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var webhook models.Webhook
	if err := h.db.First(&webhook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	var req struct {
		URL      string `json:"url"`
		Events   string `json:"events"`
		IsActive bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"url":        req.URL,
		"events":     req.Events,
		"is_active":  req.IsActive,
		"updated_at": time.Now(),
	}

	if err := h.db.Model(&webhook).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook"})
		return
	}

	c.JSON(http.StatusOK, webhook)
}

func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	if err := h.db.Delete(&models.Webhook{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *WebhookHandler) TestWebhook(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var webhook models.Webhook
	if err := h.db.First(&webhook, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	payload := map[string]interface{}{
		"event":     "test",
		"timestamp": time.Now().Unix(),
		"data": map[string]interface{}{
			"message": "This is a test webhook",
		},
	}

	if err := h.TriggerWebhook(&webhook, "test", payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook test sent successfully"})
}

func (h *WebhookHandler) TriggerWebhook(webhook *models.Webhook, event string, data interface{}) error {
	if !webhook.IsActive {
		return nil
	}

	events := []string{}
	json.Unmarshal([]byte(webhook.Events), &events)

	eventFound := false
	for _, e := range events {
		if e == event || e == "*" {
			eventFound = true
			break
		}
	}

	if !eventFound {
		return nil
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", event)
	req.Header.Set("X-Webhook-Signature", h.generateSignature(webhook.Secret, payload))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (h *WebhookHandler) TriggerWebhooksForProject(projectID uint, event string, data interface{}) {
	var webhooks []models.Webhook
	h.db.Where("project_id = ? AND is_active = ?", projectID, true).Find(&webhooks)

	for _, webhook := range webhooks {
		go h.TriggerWebhook(&webhook, event, data)
	}
}

func (h *WebhookHandler) generateSecret() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() & 0xff)
	}
	return hex.EncodeToString(b)
}

func (h *WebhookHandler) generateSignature(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}