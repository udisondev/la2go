package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RecipeRow represents a stored recipe entry.
type RecipeRow struct {
	RecipeID  int32
	IsDwarven bool
}

// RecipeRepository manages character recipes in the database.
// Phase 15: Recipe/Craft System.
type RecipeRepository struct {
	db *pgxpool.Pool
}

// NewRecipeRepository creates a new RecipeRepository.
func NewRecipeRepository(db *pgxpool.Pool) *RecipeRepository {
	return &RecipeRepository{db: db}
}

// LoadByCharacterID loads all recipes for a character.
func (r *RecipeRepository) LoadByCharacterID(ctx context.Context, charID int64) ([]RecipeRow, error) {
	query := `
		SELECT recipe_id, type
		FROM character_recipes
		WHERE character_id = $1
		ORDER BY recipe_id
	`

	rows, err := r.db.Query(ctx, query, charID)
	if err != nil {
		return nil, fmt.Errorf("querying recipes for character %d: %w", charID, err)
	}
	defer rows.Close()

	recipes := make([]RecipeRow, 0, 16)
	for rows.Next() {
		var recipeID int32
		var recipeType int16
		if err := rows.Scan(&recipeID, &recipeType); err != nil {
			return nil, fmt.Errorf("scanning recipe row: %w", err)
		}

		recipes = append(recipes, RecipeRow{
			RecipeID:  recipeID,
			IsDwarven: recipeType == 1,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating recipe rows: %w", err)
	}

	return recipes, nil
}

// SaveTx saves all recipes for a character within a transaction.
// Performs full replace: deletes all existing, then inserts.
func (r *RecipeRepository) SaveTx(ctx context.Context, tx pgx.Tx, charID int64, recipes []RecipeRow) error {
	// Delete existing recipes
	if _, err := tx.Exec(ctx, `DELETE FROM character_recipes WHERE character_id = $1`, charID); err != nil {
		return fmt.Errorf("deleting old recipes for character %d: %w", charID, err)
	}

	if len(recipes) == 0 {
		return nil
	}

	// Bulk insert using COPY
	rows := make([][]any, 0, len(recipes))
	for _, rec := range recipes {
		recipeType := int16(0) // common
		if rec.IsDwarven {
			recipeType = 1 // dwarven
		}
		rows = append(rows, []any{charID, rec.RecipeID, recipeType})
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"character_recipes"},
		[]string{"character_id", "recipe_id", "type"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("inserting recipes for character %d: %w", charID, err)
	}

	slog.Debug("saved character recipes",
		"characterID", charID,
		"count", len(recipes))

	return nil
}

// Save saves all recipes for a character (standalone, creates own transaction).
func (r *RecipeRepository) Save(ctx context.Context, charID int64, recipes []RecipeRow) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			slog.Error("rollback failed", "characterID", charID, "error", err)
		}
	}()

	if err := r.SaveTx(ctx, tx, charID, recipes); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
