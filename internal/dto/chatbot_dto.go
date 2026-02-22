package dto

import "github.com/google/uuid"

type CreateChatSessionResponse struct {
	Id uuid.UUID `json:"id"`
}
