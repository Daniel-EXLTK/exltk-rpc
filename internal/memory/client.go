package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ConversationContext struct {
	SessionID string                 `json:"session_id"`
	UserID    string                 `json:"user_id"`
	Messages  []Message              `json:"messages"`
	State     map[string]interface{} `json:"state"`
	UpdatedAt time.Time              `json:"updated_at"`
}

func SaveConversationContext(ctx context.Context, baseURL string, conv ConversationContext) error {
	conv.UpdatedAt = time.Now().UTC()
	body, err := json.Marshal(conv)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/memory/save", baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("save failed: %s", resp.Status)
	}
	return nil
}

func GetConversationContext(ctx context.Context, baseURL, sessionID string) (*ConversationContext, error) {
	url := fmt.Sprintf("%s/memory/get?session_id=%s", baseURL, sessionID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get failed: %s", resp.Status)
	}
	var conv ConversationContext
	if err := json.NewDecoder(resp.Body).Decode(&conv); err != nil {
		return nil, err
	}
	return &conv, nil
}

func DeleteConversationContext(ctx context.Context, baseURL, sessionID string) error {
	body, _ := json.Marshal(map[string]string{"session_id": sessionID})
	url := fmt.Sprintf("%s/memory/delete", baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: %s", resp.Status)
	}
	return nil
}
