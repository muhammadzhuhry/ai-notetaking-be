package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/pkg/database"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatMessageRawRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRawRepository
	Create(ctx context.Context, chatMessage *entity.ChatMessageRaw) error
	GetChatBySessionId(ctx context.Context, sessionId uuid.UUID) ([]*entity.ChatMessageRaw, error)
}

type chatMessageRawRepository struct {
	db database.DatabaseQueryer
}

func NewChatMessageRawRepository(db *pgxpool.Pool) IChatMessageRawRepository {
	return &chatMessageRawRepository{
		db: db,
	}
}

func (cm *chatMessageRawRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IChatMessageRawRepository {
	return &chatMessageRawRepository{
		db: tx,
	}
}

func (cm *chatMessageRawRepository) Create(ctx context.Context, chatMessage *entity.ChatMessageRaw) error {
	_, err := cm.db.Exec(
		ctx,
		`INSERT INTO chat_message_raw (id, role, chat, chat_session_id, created_at, updated_at, deleted_at, is_deleted) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
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

func (cm *chatMessageRawRepository) GetChatBySessionId(ctx context.Context, sessionId uuid.UUID) ([]*entity.ChatMessageRaw, error) {
	rows, err := cm.db.Query(
		ctx,
		`SELECT id, role, chat, chat_session_id, created_at, updated_at, deleted_at, is_deleted FROM chat_message_raw WHERE chat_session_id = $1 AND is_deleted = false ORDER BY created_at ASC`,
		sessionId,
	)
	if err != nil {
		return nil, err
	}

	res := make([]*entity.ChatMessageRaw, 0)
	for rows.Next() {
		var chatMessageRaw entity.ChatMessageRaw
		if err := rows.Scan(
			&chatMessageRaw.Id,
			&chatMessageRaw.Role,
			&chatMessageRaw.Chat,
			&chatMessageRaw.ChatSessionId,
			&chatMessageRaw.CreatedAt,
			&chatMessageRaw.UpdatedAt,
			&chatMessageRaw.DeletedAt,
			&chatMessageRaw.IsDeleted,
		); err != nil {
			return nil, err
		}
		res = append(res, &chatMessageRaw)
	}
	return res, nil
}
