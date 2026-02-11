package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/model"
)

// SkillRepository управляет скиллами персонажей в БД.
//
// Phase 5.9.2: Skill Trees & Player Skills.
type SkillRepository struct {
	db *pgxpool.Pool
}

// NewSkillRepository создаёт новый SkillRepository.
func NewSkillRepository(db *pgxpool.Pool) *SkillRepository {
	return &SkillRepository{db: db}
}

// LoadByCharacterID загружает все скиллы персонажа.
func (r *SkillRepository) LoadByCharacterID(ctx context.Context, charID int64) ([]*model.SkillInfo, error) {
	query := `
		SELECT skill_id, skill_level
		FROM character_skills
		WHERE character_id = $1 AND class_index = 0
		ORDER BY skill_id
	`

	rows, err := r.db.Query(ctx, query, charID)
	if err != nil {
		return nil, fmt.Errorf("querying skills for character %d: %w", charID, err)
	}
	defer rows.Close()

	skills := make([]*model.SkillInfo, 0, 32)
	for rows.Next() {
		var skillID, skillLevel int32
		if err := rows.Scan(&skillID, &skillLevel); err != nil {
			return nil, fmt.Errorf("scanning skill row: %w", err)
		}

		skills = append(skills, &model.SkillInfo{
			SkillID: skillID,
			Level:   skillLevel,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating skill rows: %w", err)
	}

	return skills, nil
}

// Save сохраняет все скиллы персонажа (полная перезапись).
// Удаляет старые, вставляет новые в одной транзакции.
func (r *SkillRepository) Save(ctx context.Context, charID int64, skills []*model.SkillInfo) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			// Rollback after commit is expected to fail
		}
	}()

	// Delete existing skills
	if _, err := tx.Exec(ctx, `DELETE FROM character_skills WHERE character_id = $1 AND class_index = 0`, charID); err != nil {
		return fmt.Errorf("deleting existing skills: %w", err)
	}

	// Insert new skills
	for _, s := range skills {
		if _, err := tx.Exec(ctx,
			`INSERT INTO character_skills (character_id, skill_id, skill_level, class_index) VALUES ($1, $2, $3, 0)`,
			charID, s.SkillID, s.Level,
		); err != nil {
			return fmt.Errorf("inserting skill %d: %w", s.SkillID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing skills save: %w", err)
	}

	return nil
}

// AddSkill добавляет один скилл (UPSERT).
func (r *SkillRepository) AddSkill(ctx context.Context, charID int64, skill *model.SkillInfo) error {
	query := `
		INSERT INTO character_skills (character_id, skill_id, skill_level, class_index)
		VALUES ($1, $2, $3, 0)
		ON CONFLICT (character_id, skill_id, class_index)
		DO UPDATE SET skill_level = $3
	`

	if _, err := r.db.Exec(ctx, query, charID, skill.SkillID, skill.Level); err != nil {
		return fmt.Errorf("upserting skill %d for character %d: %w", skill.SkillID, charID, err)
	}

	return nil
}

// DeleteSkill удаляет один скилл.
func (r *SkillRepository) DeleteSkill(ctx context.Context, charID int64, skillID int32) error {
	query := `DELETE FROM character_skills WHERE character_id = $1 AND skill_id = $2 AND class_index = 0`

	if _, err := r.db.Exec(ctx, query, charID, skillID); err != nil {
		return fmt.Errorf("deleting skill %d for character %d: %w", skillID, charID, err)
	}

	return nil
}
