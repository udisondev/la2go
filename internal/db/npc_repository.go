package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/model"
)

// NpcRepository handles NPC template CRUD operations
type NpcRepository struct {
	pool *pgxpool.Pool
}

// NewNpcRepository creates a new NPC repository
func NewNpcRepository(pool *pgxpool.Pool) *NpcRepository {
	return &NpcRepository{pool: pool}
}

// LoadTemplate loads NPC template by ID
func (r *NpcRepository) LoadTemplate(ctx context.Context, id int32) (*model.NpcTemplate, error) {
	query := `
		SELECT template_id, name, title, level, max_hp, max_mp,
		       p_atk, p_def, m_atk, m_def, aggro_range, move_speed, atk_speed,
		       respawn_min, respawn_max, base_exp, base_sp
		FROM npc_templates
		WHERE template_id = $1
	`

	var (
		templateID  int32
		name, title string
		level       int32
		maxHP       int32
		maxMP       int32
		pAtk        int32
		pDef        int32
		mAtk        int32
		mDef        int32
		aggroRange  int32
		moveSpeed   int32
		atkSpeed    int32
		respawnMin  int32
		respawnMax  int32
		baseExp     int64
		baseSP      int32
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&templateID, &name, &title, &level, &maxHP, &maxMP,
		&pAtk, &pDef, &mAtk, &mDef, &aggroRange, &moveSpeed, &atkSpeed,
		&respawnMin, &respawnMax, &baseExp, &baseSP,
	)
	if err != nil {
		return nil, fmt.Errorf("loading npc template %d: %w", id, err)
	}

	return model.NewNpcTemplate(
		templateID, name, title, level, maxHP, maxMP,
		pAtk, pDef, mAtk, mDef, aggroRange, moveSpeed, atkSpeed,
		respawnMin, respawnMax, baseExp, baseSP,
	), nil
}

// LoadAllTemplates loads all NPC templates
func (r *NpcRepository) LoadAllTemplates(ctx context.Context) ([]*model.NpcTemplate, error) {
	query := `
		SELECT template_id, name, title, level, max_hp, max_mp,
		       p_atk, p_def, m_atk, m_def, aggro_range, move_speed, atk_speed,
		       respawn_min, respawn_max, base_exp, base_sp
		FROM npc_templates
		ORDER BY template_id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("loading all npc templates: %w", err)
	}
	defer rows.Close()

	templates := make([]*model.NpcTemplate, 0, 100) // pre-allocate for typical count

	for rows.Next() {
		var (
			templateID  int32
			name, title string
			level       int32
			maxHP       int32
			maxMP       int32
			pAtk        int32
			pDef        int32
			mAtk        int32
			mDef        int32
			aggroRange  int32
			moveSpeed   int32
			atkSpeed    int32
			respawnMin  int32
			respawnMax  int32
			baseExp     int64
			baseSP      int32
		)

		if err := rows.Scan(
			&templateID, &name, &title, &level, &maxHP, &maxMP,
			&pAtk, &pDef, &mAtk, &mDef, &aggroRange, &moveSpeed, &atkSpeed,
			&respawnMin, &respawnMax, &baseExp, &baseSP,
		); err != nil {
			return nil, fmt.Errorf("scanning npc template row: %w", err)
		}

		template := model.NewNpcTemplate(
			templateID, name, title, level, maxHP, maxMP,
			pAtk, pDef, mAtk, mDef, aggroRange, moveSpeed, atkSpeed,
			respawnMin, respawnMax, baseExp, baseSP,
		)

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating npc template rows: %w", err)
	}

	return templates, nil
}

// Create creates new NPC template
func (r *NpcRepository) Create(ctx context.Context, template *model.NpcTemplate) error {
	query := `
		INSERT INTO npc_templates (
			template_id, name, title, level, max_hp, max_mp,
			p_atk, p_def, m_atk, m_def, aggro_range, move_speed, atk_speed,
			respawn_min, respawn_max, base_exp, base_sp
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
	`

	_, err := r.pool.Exec(ctx, query,
		template.TemplateID(),
		template.Name(),
		template.Title(),
		template.Level(),
		template.MaxHP(),
		template.MaxMP(),
		template.PAtk(),
		template.PDef(),
		template.MAtk(),
		template.MDef(),
		template.AggroRange(),
		template.MoveSpeed(),
		template.AtkSpeed(),
		template.RespawnMin(),
		template.RespawnMax(),
		template.BaseExp(),
		template.BaseSP(),
	)
	if err != nil {
		return fmt.Errorf("creating npc template %d: %w", template.TemplateID(), err)
	}

	return nil
}
