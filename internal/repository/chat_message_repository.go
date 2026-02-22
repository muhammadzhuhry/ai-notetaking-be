package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatMessageRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRepository
	Create(ctx context.Context, chatMessage *entity.ChatMessage) error
}

type chatMessageRepository struct {
	db database.DatabaseQueryer
}

func NewChatMessageRepository(db *pgxpool.Pool) IChatMessageRepository {
	return &chatMessageRepository{
		db: db,
	}
}

func (cm *chatMessageRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRepository {
	return &chatMessageRepository{
		db: tx,
	}
}

func (cm *chatMessageRepository) Create(ctx context.Context, chatMessage *entity.ChatMessage) error {
	_, err := cm.db.Exec(
		ctx,
		`INSERT INTO chat_message (id, role, chat, chat_session_id, created_at, updated_at, deleted_at, is_deleted) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		chatMessage.Id,
		chatMessage.Role,
		chatMessage.Chat,
		chatMessage.ChatSessionId,
		chatMessage.CreatedAt,
		chatMessage.UpdatedAt,
		chatMessage.DeletedAt,
		chatMessage.IsDeleted,
	)
	if err != nil {
		return err
	}
	return nil
}
