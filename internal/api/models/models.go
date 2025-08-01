package models

import (
	"encoding/json"
	"time"
)

// APIResponse - это универсальная обертка для всех ответов API.
type APIResponse struct {
	Error    *APIError   `json:"error,omitempty"`
	Response interface{} `json:"response,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

type APIError struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type RegisterRequestDTO struct {
	Token    string `json:"token"`
	Login    string `json:"login"`
	Password string `json:"pswd"`
}

type RegisterResponseDTO struct {
	Login string `json:"login"`
}

type AuthRequestDTO struct {
	Login    string `json:"login"`
	Password string `json:"pswd"`
}

type AuthResponseDTO struct {
	Token string `json:"token"`
}

type DocumentUploadMetaDTO struct {
	Name   string   `json:"name"`
	File   bool     `json:"file"`
	Public bool     `json:"public"`
	Token  string   `json:"token"`
	Mime   string   `json:"mime"`
	Grant  []string `json:"grant"`
}

type DocumentResponseDTO struct {
	JSON json.RawMessage `json:"json,omitempty"`
	File string          `json:"file,omitempty"`
}

type DocumentListItemDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Mime      string    `json:"mime"`
	File      bool      `json:"file"`
	Public    bool      `json:"public"`
	CreatedAt time.Time `json:"created_at"`
	Grant     []string  `json:"grant,omitempty"`
}
