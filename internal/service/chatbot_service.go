package service

import (
	"ai-notetaking-be/internal/constant"
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/pkg/chatbot"
	"ai-notetaking-be/pkg/embedding"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IChatbotService interface {
	CreateSession(ctx context.Context) (*dto.CreateChatSessionResponse, error)
	GetAllSession(ctx context.Context) ([]*dto.GetAllSessionResponse, error)
	GetChatHistory(ctx context.Context, sessionId uuid.UUID) ([]*dto.GetChatHistoryResponse, error)
	SendChat(ctx context.Context, request dto.SendChatRequest) (*dto.SendChatResponse, error)
}

type chatbotService struct {
	db                       *pgxpool.Pool
	chatSessionRepository    repository.IChatSessionRepository
	chatMessageRepository    repository.IChatMessageRepository
	chatMessageRawRepository repository.IChatMessageRawRepository
	noteEmbeddingRepository  repository.INoteEmbeddingRepository
}

func NewChatbotService(
	db *pgxpool.Pool,
	chatSessionRepository repository.IChatSessionRepository,
	chatMessageRepository repository.IChatMessageRepository,
	chatMessageRawRepository repository.IChatMessageRawRepository,
	noteEmbeddingRepository repository.INoteEmbeddingRepository,
) IChatbotService {
	return &chatbotService{
		db:                       db,
		chatSessionRepository:    chatSessionRepository,
		chatMessageRepository:    chatMessageRepository,
		chatMessageRawRepository: chatMessageRawRepository,
		noteEmbeddingRepository:  noteEmbeddingRepository,
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

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.CreateChatSessionResponse{
		Id: chatSession.Id,
	}, nil
}

func (cs *chatbotService) GetAllSession(ctx context.Context) ([]*dto.GetAllSessionResponse, error) {
	chats, err := cs.chatSessionRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.GetAllSessionResponse, 0)

	for _, chat := range chats {
		response = append(response, &dto.GetAllSessionResponse{
			Id:        chat.Id,
			Title:     chat.Title,
			CreatedAt: chat.CreatedAt,
			UpdatedAt: chat.UpdatedAt,
		})
	}
	return response, nil
}

func (cs *chatbotService) GetChatHistory(ctx context.Context, sessionId uuid.UUID) ([]*dto.GetChatHistoryResponse, error) {
	_, err := cs.chatSessionRepository.GetById(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	chatMessages, err := cs.chatMessageRepository.GetChatBySessionId(ctx, sessionId)
	if err != nil {
		return nil, err
	}

	response := make([]*dto.GetChatHistoryResponse, 0)
	for _, message := range chatMessages {
		response = append(response, &dto.GetChatHistoryResponse{
			Id:        message.Id,
			Role:      message.Role,
			Chat:      message.Chat,
			CreatedAt: message.CreatedAt,
		})
	}
	return response, nil
}

func (cs *chatbotService) SendChat(ctx context.Context, request dto.SendChatRequest) (*dto.SendChatResponse, error) {

	tx, err := cs.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	chatSessionRepo := cs.chatSessionRepository.UsingTx(ctx, tx)
	chatMessageRepo := cs.chatMessageRepository.UsingTx(ctx, tx)
	chatMessageRawRepo := cs.chatMessageRawRepository.UsingTx(ctx, tx)
	noteEmbeddingRepo := cs.noteEmbeddingRepository.UsingTx(ctx, tx)

	chatSession, err := chatSessionRepo.GetById(ctx, request.ChatSessionId)
	if err != nil {
		return nil, err
	}

	existingRawChats, err := chatMessageRawRepo.GetChatBySessionId(ctx, request.ChatSessionId)
	if err != nil {
		return nil, err
	}

	updateSessionTitle := len(existingRawChats) == 2

	now := time.Now()

	chatMessage := entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          request.Chat,
		Role:          constant.ChatRoleUser,
		ChatSessionId: chatSession.Id,
		CreatedAt:     now,
	}

	embeddingRes, err := embedding.GetGeminiEmbedding(
		os.Getenv("GOOGLE_GEMINI_API_KEY"),
		request.Chat,
		"RETRIEVAL_QUERY",
	)
	if err != nil {
		return nil, err
	}

	noteEmbeddings, err := noteEmbeddingRepo.SearchSimilarity(ctx, embeddingRes.Embedding.Values)
	if err != nil {
		return nil, err
	}

	strBuilder := strings.Builder{}

	for i, noteEmbedding := range noteEmbeddings {
		strBuilder.WriteString(fmt.Sprintf("Reference %d\n", i+1))
		strBuilder.WriteString(noteEmbedding.Document)
		strBuilder.WriteString("\n\n")
	}

	strBuilder.WriteString("User next question: ")
	strBuilder.WriteString(request.Chat)
	strBuilder.WriteString("\n\n")
	strBuilder.WriteString("Your answer ?")
	chatMessageRaw := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          strBuilder.String(),
		Role:          constant.ChatRoleUser,
		ChatSessionId: chatSession.Id,
		CreatedAt:     now.Add(1 * time.Millisecond),
	}

	existingRawChats = append(
		existingRawChats,
		&chatMessageRaw,
	)

	geminiReq := make([]*chatbot.ChatHistory, 0)
	for _, existing := range existingRawChats {
		geminiReq = append(geminiReq, &chatbot.ChatHistory{
			Chat: existing.Chat,
			Role: existing.Role,
		})
	}

	reply, err := chatbot.GetGeminiResponse(
		ctx,
		os.Getenv("GOOGLE_GEMINI_API_KEY"),
		geminiReq,
	)
	if err != nil {
		return nil, err
	}

	chatMessageModel := entity.ChatMessage{
		Id:            uuid.New(),
		Chat:          reply,
		Role:          constant.ChatRoleModel,
		ChatSessionId: chatSession.Id,
		CreatedAt:     now.Add(1 * time.Millisecond),
	}

	chatMessageModelRaw := entity.ChatMessageRaw{
		Id:            uuid.New(),
		Chat:          reply,
		Role:          constant.ChatRoleModel,
		ChatSessionId: chatSession.Id,
		CreatedAt:     now.Add(1 * time.Millisecond),
	}

	chatMessageRepo.Create(ctx, &chatMessage)
	chatMessageRepo.Create(ctx, &chatMessageModel)
	chatMessageRawRepo.Create(ctx, &chatMessageRaw)
	chatMessageRawRepo.Create(ctx, &chatMessageModelRaw)

	if updateSessionTitle {
		chatSession.Title = request.Chat
		chatSession.UpdatedAt = &now
		// TODO: update chat session
		err = chatSessionRepo.Update(ctx, chatSession)
		if err != nil {
			return nil, err
		}

		err = tx.Commit(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &dto.SendChatResponse{
		ChatSessionId: request.ChatSessionId,
		Title:         chatSession.Title,
		Sent: &dto.SendChatResponseChat{
			Id:        chatMessage.Id,
			Role:      chatMessage.Role,
			Chat:      chatMessage.Chat,
			CreatedAt: chatMessage.CreatedAt,
		},
		Reply: &dto.SendChatResponseChat{
			Id:        chatMessageModel.Id,
			Role:      chatMessageModel.Role,
			Chat:      chatMessageModel.Chat,
			CreatedAt: chatMessageModel.CreatedAt,
		},
	}, nil
}
