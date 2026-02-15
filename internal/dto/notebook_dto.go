package dto

import "github.com/google/uuid"

type CreateNotebookRequest struct {
	Name     string     `json:"name" validate:"required"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type CreateNotebookResponse struct {
	Id string `json:"id"`
}
