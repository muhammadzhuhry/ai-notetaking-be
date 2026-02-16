package repository

import (
	"ai-notetaking-be/internal/entity"
	"ai-notetaking-be/internal/pkg/serverutils"
	"ai-notetaking-be/pkg/database"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INotebookRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) INotebookRepository
	Create(ctx context.Context, notebook *entity.Notebook) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Notebook, error)
}

// Concrete implementation of INotebookRepository
type notebookRepository struct {
	db database.DatabaseQueryer
}

func (n *notebookRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) INotebookRepository {
	return &notebookRepository{
		db: tx,
	}
}

func (n *notebookRepository) Create(ctx context.Context, notebook *entity.Notebook) error {
	_, err := n.db.Exec(
		ctx,
		`INSERT INTO notebooks (id, name, parent_id, created_at, updated_at, deleted_at, is_deleted) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		notebook.Id,
		notebook.Name,
		notebook.ParentId,
		notebook.CreatedAt,
		notebook.UpdatedAt,
		notebook.DeletedAt,
		notebook.IsDeleted,
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *notebookRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notebook, error) {
	row := n.db.QueryRow(
		ctx,
		`SELECT id, name, parent_id, created_at, updated_at FROM notebooks n WHERE n.is_deleted = false AND n.id = $1`,
		id,
	)

	var notebook entity.Notebook
	err := row.Scan(&notebook.Id, &notebook.Name, &notebook.ParentId, &notebook.CreatedAt, &notebook.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, serverutils.ErrNotFound
		}
		return nil, err
	}

	return &notebook, nil
}

// Factory function
func NewNotebookRepository(db *pgxpool.Pool) INotebookRepository {
	return &notebookRepository{
		db: db,
	}
}
