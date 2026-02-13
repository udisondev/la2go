package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/game/quest"
)

// QuestRepository manages character quest data in the database.
// Phase 16: Quest System Framework.
type QuestRepository struct {
	db *pgxpool.Pool
}

// NewQuestRepository creates a new QuestRepository.
func NewQuestRepository(db *pgxpool.Pool) *QuestRepository {
	return &QuestRepository{db: db}
}

// LoadByCharacterID loads all quest variables for a character.
func (r *QuestRepository) LoadByCharacterID(ctx context.Context, charID int64) ([]quest.QuestVar, error) {
	query := `
		SELECT quest_name, variable, value
		FROM character_quests
		WHERE character_id = $1
		ORDER BY quest_name, variable
	`

	rows, err := r.db.Query(ctx, query, charID)
	if err != nil {
		return nil, fmt.Errorf("querying quests for character %d: %w", charID, err)
	}
	defer rows.Close()

	vars := make([]quest.QuestVar, 0, 32)
	for rows.Next() {
		var v quest.QuestVar
		if err := rows.Scan(&v.QuestName, &v.Variable, &v.Value); err != nil {
			return nil, fmt.Errorf("scanning quest var row: %w", err)
		}
		vars = append(vars, v)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating quest var rows: %w", err)
	}

	return vars, nil
}

// SaveQuestState saves all variables for a single quest.
// Performs full replace: deletes all existing vars, then inserts new ones.
func (r *QuestRepository) SaveQuestState(ctx context.Context, charID int64, questName string, vars map[string]string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err.Error() != "tx is closed" {
			slog.Error("rollback failed", "characterID", charID, "quest", questName, "error", err)
		}
	}()

	if err := r.SaveQuestStateTx(ctx, tx, charID, questName, vars); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// SaveQuestStateTx saves quest state within an existing transaction.
func (r *QuestRepository) SaveQuestStateTx(ctx context.Context, tx pgx.Tx, charID int64, questName string, vars map[string]string) error {
	// Удаляем существующие переменные
	if _, err := tx.Exec(ctx,
		`DELETE FROM character_quests WHERE character_id = $1 AND quest_name = $2`,
		charID, questName,
	); err != nil {
		return fmt.Errorf("deleting old quest vars for character %d quest %q: %w", charID, questName, err)
	}

	if len(vars) == 0 {
		return nil
	}

	// Вставляем новые переменные через COPY
	rows := make([][]any, 0, len(vars))
	for k, v := range vars {
		rows = append(rows, []any{charID, questName, k, v})
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"character_quests"},
		[]string{"character_id", "quest_name", "variable", "value"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("inserting quest vars for character %d quest %q: %w", charID, questName, err)
	}

	return nil
}

// DeleteQuest removes all data for a quest from the database.
func (r *QuestRepository) DeleteQuest(ctx context.Context, charID int64, questName string) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM character_quests WHERE character_id = $1 AND quest_name = $2`,
		charID, questName,
	)
	if err != nil {
		return fmt.Errorf("deleting quest %q for character %d: %w", questName, charID, err)
	}
	return nil
}

// SaveAllTx saves all quest states for a character within a transaction.
// Used by PlayerPersistenceService for atomic saves.
func (r *QuestRepository) SaveAllTx(ctx context.Context, tx pgx.Tx, charID int64, states []*quest.QuestState) error {
	// Удаляем все квесты персонажа
	if _, err := tx.Exec(ctx,
		`DELETE FROM character_quests WHERE character_id = $1`,
		charID,
	); err != nil {
		return fmt.Errorf("deleting all quests for character %d: %w", charID, err)
	}

	// Собираем все переменные из всех квестов
	var allRows [][]any
	for _, qs := range states {
		vars := qs.Vars()
		for k, v := range vars {
			allRows = append(allRows, []any{charID, qs.QuestName(), k, v})
		}
	}

	if len(allRows) == 0 {
		return nil
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"character_quests"},
		[]string{"character_id", "quest_name", "variable", "value"},
		pgx.CopyFromRows(allRows),
	)
	if err != nil {
		return fmt.Errorf("inserting quest vars for character %d: %w", charID, err)
	}

	slog.Debug("saved character quests",
		"characterID", charID,
		"questCount", len(states))

	return nil
}
