package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/pkg/embedding"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
)

type IConsumerService interface {
	Consume(ctx context.Context) error
}

type consumerService struct {
	noteRepository          repository.INoteRepository
	noteEmbeddingRepository repository.INoteEmbeddingRepository
	notebookRepository      repository.INotebookRepository
	pubSub                  *gochannel.GoChannel
	topicName               string
}

func NewConsumerService(
	pubSub *gochannel.GoChannel,
	topicName string, noteRepository repository.INoteRepository,
	noteEmbeddingRepository repository.INoteEmbeddingRepository,
	notebookRepository repository.INotebookRepository,
) IConsumerService {
	return &consumerService{
		noteRepository:          noteRepository,
		noteEmbeddingRepository: noteEmbeddingRepository,
		notebookRepository:      notebookRepository,
		pubSub:                  pubSub,
		topicName:               topicName,
	}
}

// Consume implements [IConsumerService].
func (cs *consumerService) Consume(ctx context.Context) error {
	messages, err := cs.pubSub.Subscribe(ctx, cs.topicName)
	if err != nil {
		return err
	}

	go func() {
		for msg := range messages {
			cs.processMessage(ctx, msg)
		}
	}()

	return nil
}

func (cs *consumerService) processMessage(ctx context.Context, msg *message.Message) {
	defer msg.Nack() // Ensure message is acknowledged or negatively acknowledged
	defer func() {
		if e := recover(); e != nil {
			log.Error(e)
		}
	}()

	var payload dto.PublishEmbedNoteMessage
	err := json.Unmarshal(msg.Payload, &payload)
	if err != nil {
		panic(err)
	}

	note, err := cs.noteRepository.GetById(ctx, payload.NoteId)
	if err != nil {
		panic(err)
	}

	notebook, err := cs.notebookRepository.GetByID(ctx, note.NotebookId)
	if err != nil {
		panic(err)
	}

	// Enhrichment content
	noteUpdatedAt := "-"
	if note.UpdatedAt != nil {
		noteUpdatedAt = note.UpdatedAt.Format(time.RFC3339)
	}
	content := fmt.Sprintf(`
	Note Title: %s
	Notebook Title: %s

	%s

	Created At: %s
	Updated At: %s
	`,
		note.Title,
		notebook.Name,
		note.Content,
		note.CreatedAt.Format(time.RFC3339),
		noteUpdatedAt,
	)

	// Process embedding
	res, err := embedding.GetGeminiEmbedding(os.Getenv("GOOGLE_GEMINI_API_KEY"), content)
	if err != nil {
		panic(err)
	}

	entity := &entity.NoteEmbedding{
		Id:             uuid.New(),
		Document:       content,
		EmbeddingValue: res.Embedding.Values,
		NoteId:         payload.NoteId,
		CreatedAt:      time.Now(),
	}

	err = cs.noteEmbeddingRepository.Create(ctx, entity)
	if err != nil {
		panic(err)
	}

	msg.Ack()
}
