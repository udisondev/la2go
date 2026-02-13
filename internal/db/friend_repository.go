package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// FriendRow represents a row in character_friends table.
// Relation: 0 = friend, 1 = blocked.
type FriendRow struct {
	FriendID int32
	Relation int16 // 0=friend, 1=blocked
}

// FriendRepository manages character friends and blocks in the database.
// Phase 35: Friend/Block persistence.
type FriendRepository struct {
	db *pgxpool.Pool
}

// NewFriendRepository creates a new FriendRepository.
func NewFriendRepository(db *pgxpool.Pool) *FriendRepository {
	return &FriendRepository{db: db}
}

// LoadByCharacterID loads all friend and block entries for a character.
func (r *FriendRepository) LoadByCharacterID(ctx context.Context, charID int64) ([]FriendRow, error) {
	query := `
		SELECT friend_id, relation
		FROM character_friends
		WHERE char_id = $1
		ORDER BY friend_id
	`

	rows, err := r.db.Query(ctx, query, charID)
	if err != nil {
		return nil, fmt.Errorf("querying friends for character %d: %w", charID, err)
	}
	defer rows.Close()

	result := make([]FriendRow, 0, 16)
	for rows.Next() {
		var fr FriendRow
		if err := rows.Scan(&fr.FriendID, &fr.Relation); err != nil {
			return nil, fmt.Errorf("scanning friend row: %w", err)
		}
		result = append(result, fr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating friend rows: %w", err)
	}

	return result, nil
}

// SaveTx saves all friend/block entries within a transaction (full replace).
func (r *FriendRepository) SaveTx(ctx context.Context, tx pgx.Tx, charID int64, entries []FriendRow) error {
	if _, err := tx.Exec(ctx, `DELETE FROM character_friends WHERE char_id = $1`, charID); err != nil {
		return fmt.Errorf("deleting old friends for character %d: %w", charID, err)
	}

	if len(entries) == 0 {
		return nil
	}

	rows := make([][]any, 0, len(entries))
	for _, e := range entries {
		rows = append(rows, []any{charID, int64(e.FriendID), e.Relation})
	}

	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"character_friends"},
		[]string{"char_id", "friend_id", "relation"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("inserting friends for character %d: %w", charID, err)
	}

	slog.Debug("saved character friends",
		"characterID", charID,
		"count", len(entries))

	return nil
}

// InsertFriend adds a single friend relation (relation=0) immediately.
func (r *FriendRepository) InsertFriend(ctx context.Context, charID int64, friendID int32) error {
	query := `
		INSERT INTO character_friends (char_id, friend_id, relation)
		VALUES ($1, $2, 0)
		ON CONFLICT (char_id, friend_id) DO NOTHING
	`
	if _, err := r.db.Exec(ctx, query, charID, int64(friendID)); err != nil {
		return fmt.Errorf("inserting friend %d for character %d: %w", friendID, charID, err)
	}
	return nil
}

// DeleteFriend removes a single friend relation immediately.
func (r *FriendRepository) DeleteFriend(ctx context.Context, charID int64, friendID int32) error {
	query := `DELETE FROM character_friends WHERE char_id = $1 AND friend_id = $2 AND relation = 0`
	if _, err := r.db.Exec(ctx, query, charID, int64(friendID)); err != nil {
		return fmt.Errorf("deleting friend %d for character %d: %w", friendID, charID, err)
	}
	return nil
}

// InsertBlock adds a single block relation (relation=1) immediately.
func (r *FriendRepository) InsertBlock(ctx context.Context, charID int64, blockedID int32) error {
	query := `
		INSERT INTO character_friends (char_id, friend_id, relation)
		VALUES ($1, $2, 1)
		ON CONFLICT (char_id, friend_id) DO UPDATE SET relation = 1
	`
	if _, err := r.db.Exec(ctx, query, charID, int64(blockedID)); err != nil {
		return fmt.Errorf("inserting block %d for character %d: %w", blockedID, charID, err)
	}
	return nil
}

// DeleteBlock removes a single block relation immediately.
func (r *FriendRepository) DeleteBlock(ctx context.Context, charID int64, blockedID int32) error {
	query := `DELETE FROM character_friends WHERE char_id = $1 AND friend_id = $2 AND relation = 1`
	if _, err := r.db.Exec(ctx, query, charID, int64(blockedID)); err != nil {
		return fmt.Errorf("deleting block %d for character %d: %w", blockedID, charID, err)
	}
	return nil
}
