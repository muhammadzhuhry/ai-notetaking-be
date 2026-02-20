package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/repository"
	"ai-notetaking-be/pkg/embedding"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v2/log"
)

type IConsumerService interface {
	Consume(ctx context.Context) error
}

type consumerService struct {
	noteRepository          repository.INoteRepository
	noteEmbeddingRepository repository.INoteEmbeddingRepository
	pubSub                  *gochannel.GoChannel
	topicName               string
}

func NewConsumerService(pubSub *gochannel.GoChannel, topicName string, noteRepository repository.INoteRepository, noteEmbeddingRepository repository.INoteEmbeddingRepository) IConsumerService {
	return &consumerService{
		noteRepository:          noteRepository,
		noteEmbeddingRepository: noteEmbeddingRepository,
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

	res, err := embedding.GetGeminiEmbedding(os.Getenv("GOOGLE_GEMINI_API_KEY"), note.Content)
	if err != nil {
		panic(err)
	}

	entity := &entity.NoteEmbedding{
		Id:             payload.NoteId,
		Document:       note.Content,
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
