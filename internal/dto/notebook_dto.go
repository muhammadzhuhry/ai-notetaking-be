package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateNotebookRequest struct {
	Name     string     `json:"name" validate:"required"`
	ParentId *uuid.UUID `json:"parent_id"`
}

type CreateNotebookResponse struct {
	Id string `json:"id"`
}

type ShowNotebookResponse struct {
	Id        string     `json:"id"`
	Name      string     `json:"name"`
	ParentId  *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
