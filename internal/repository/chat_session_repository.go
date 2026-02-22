package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatSessionRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatSessionRepository
	Create(ctx context.Context, chatSession *entity.ChatSession) error
}

type chatSessionRepository struct {
	db database.DatabaseQueryer
}

func NewChatSessionRepository(db *pgxpool.Pool) IChatSessionRepository {
	return &chatSessionRepository{
		db: db,
	}
}

func (cs *chatSessionRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatSessionRepository {
	return &chatSessionRepository{
		db: tx,
	}
}

func (cs *chatSessionRepository) Create(ctx context.Context, chatSession *entity.ChatSession) error {
	_, err := cs.db.Exec(
		ctx,
		`INSERT INTO chat_session (id, title, created_at, updated_at, deleted_at, is_deleted) VALUES ($1, $2, $3, $4, $5, $6)`,
		chatSession.Id,
		chatSession.Title,
		chatSession.CreatedAt,
		chatSession.UpdatedAt,
		chatSession.DeletedAt,
		chatSession.IsDeleted,
	)
	if err != nil {
		return err
	}
	return nil
}
