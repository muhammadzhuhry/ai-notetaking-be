package service

import (
	"ai-notetaking-be/internal/constant"
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/repository"
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatbotService interface {
	CreateSession(ctx context.Context) (*dto.CreateChatSessionResponse, error)
}

type chatbotService struct {
	db                       *pgxpool.Pool
	chatSessionRepository    repository.IChatSessionRepository
	chatMessageRepository    repository.IChatMessageRepository
	chatMessageRawRepository repository.IChatMessageRawRepository
}

func NewChatbotService(
	db *pgxpool.Pool,
	chatSessionRepository repository.IChatSessionRepository,
	chatMessageRepository repository.IChatMessageRepository,
	chatMessageRawRepository repository.IChatMessageRawRepository,
) IChatbotService {
	return &chatbotService{
		db:                       db,
		chatSessionRepository:    chatSessionRepository,
		chatMessageRepository:    chatMessageRepository,
		chatMessageRawRepository: chatMessageRawRepository,
	}
}

func (cs *chatbotService) CreateSession(ctx context.Context) (*dto.CreateChatSessionResponse, error) {

	now := time.Now()
	chatSession := entity.ChatSession{
		Id:        uuid.New(),
		Title:     "Untitled Session",
		CreatedAt: now,
	}
	chatMessage := entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          "Hi, how can i help you?",
		Role:          constant.ChatRoleModel,
		ChatSessionId: chatSession.Id,
		CreatedAt:     now,
	}
	chatMessageRawUser := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          constant.ChatRawInitialUserPromptV1,
		Role:          constant.ChatRoleUser,
		ChatSessionId: chatSession.Id,
		CreatedAt:     now,
	}
	chatMessageRawModel := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          constant.ChatRawInitialModelPromptV1,
		ChatSessionId: chatSession.Id,
		Role:          constant.ChatRoleModel,
		CreatedAt:     now.Add(1 * time.Second),
	}

	tx, err := cs.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	chatSessionRepository := cs.chatSessionRepository.UsingTx(ctx, tx)
	chatMessageRepository := cs.chatMessageRepository.UsingTx(ctx, tx)
	chatMessageRawRepository := cs.chatMessageRawRepository.UsingTx(ctx, tx)

	// TODO: Insert chat session to table chat_session
	err = chatSessionRepository.Create(ctx, &chatSession)
	if err != nil {
		return nil, err
	}

	// TODO: Insert deafult chat message to table chat_message
	err = chatMessageRepository.Create(ctx, &chatMessage)
	if err != nil {
		return nil, err
	}

	// TODO: Insert deafult raw chat message to table chat_message_raw
	err = chatMessageRawRepository.Create(ctx, &chatMessageRawUser)
	if err != nil {
		return nil, err
	}
	err = chatMessageRawRepository.Create(ctx, &chatMessageRawModel)
	if err != nil {
		return nil, err
	}

	return &dto.CreateChatSessionResponse{
		Id: chatSession.Id,
	}, nil
}
