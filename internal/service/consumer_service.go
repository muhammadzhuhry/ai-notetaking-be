package service

import (
	"ai-notetaking-be/internal/dto"
	"ai-notetaking-be/internal/repository"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v2/log"
)

type EmbeddingRequestContentPart struct {
	Text string `json:"text"`
}

type EmbeddingRequestContent struct {
	Parts []EmbeddingRequestContentPart `json:"parts"`
}

type EmbeddingRequest struct {
	Model    string                  `json:"model"`
	Content  EmbeddingRequestContent `json:"content"`
	TaskType string                  `json:"task_type"`
}

type EmbeddingResponseEmbedding struct {
	Values []float32 `json:"values"`
}

type EmbeddingResponse struct {
	Embedding EmbeddingResponseEmbedding `json:"embedding"`
}

type IConsumerService interface {
	Consume(ctx context.Context) error
}

type consumerService struct {
	noteRepository repository.INoteRepository
	pubSub         *gochannel.GoChannel
	topicName      string
}

func NewConsumerService(pubSub *gochannel.GoChannel, topicName string, noteRepository repository.INoteRepository) IConsumerService {
	return &consumerService{
		noteRepository: noteRepository,
		pubSub:         pubSub,
		topicName:      topicName,
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

	geminiReq := EmbeddingRequest{
		Model: "models/gemini-embedding-001",
		Content: EmbeddingRequestContent{
			Parts: []EmbeddingRequestContentPart{
				{
					Text: note.Content,
				},
			},
		},
		TaskType: "RETRIEVAL_DOCUMENT",
	}

	geminiReqJson, err := json.Marshal(geminiReq)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent",
		bytes.NewBuffer(geminiReqJson),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("x-goog-api-key", os.Getenv("GOOGLE_GEMINI_API_KEY"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	resByte, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	if res.StatusCode != http.StatusOK {
		panic(fmt.Errorf("error from response, code %d body %s", res.StatusCode, string(resByte)))
	}

	var resEmbedding EmbeddingResponse
	err = json.Unmarshal(resByte, &resEmbedding)
	if err != nil {
		panic(err)
	}

	fmt.Println(resEmbedding.Embedding.Values)
	msg.Ack()
}
