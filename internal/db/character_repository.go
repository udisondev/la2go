package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/model"
	"github.com/udisondev/la2go/internal/world"
)

// CharacterRepository управляет персонажами в БД.
type CharacterRepository struct {
	db *pgxpool.Pool
}

// NewCharacterRepository создаёт новый CharacterRepository.
func NewCharacterRepository(db *pgxpool.Pool) *CharacterRepository {
	return &CharacterRepository{db: db}
}

// LoadByID загружает персонажа по ID.
// Возвращает nil если персонаж не найден (не ошибка).
func (r *CharacterRepository) LoadByID(ctx context.Context, characterID int64) (*model.Player, error) {
	query := `
		SELECT character_id, account_id, name, level, race_id, class_id,
		       x, y, z, heading,
		       current_hp, max_hp, current_mp, max_mp, current_cp, max_cp,
		       experience, created_at, last_login
		FROM characters
		WHERE character_id = $1
	`

	var characterIDDB int64
	var accountIDDB int64
	var name string
	var level int32
	var raceID int32
	var classID int32
	var x int32
	var y int32
	var z int32
	var heading uint16
	var currentHP int32
	var maxHP int32
	var currentMP int32
	var maxMP int32
	var currentCP int32
	var maxCP int32
	var experience int64
	var createdAt time.Time
	var lastLogin *time.Time // nullable

	err := r.db.QueryRow(ctx, query, characterID).Scan(
		&characterIDDB, &accountIDDB, &name, &level, &raceID, &classID,
		&x, &y, &z, &heading,
		&currentHP, &maxHP, &currentMP, &maxMP, &currentCP, &maxCP,
		&experience, &createdAt, &lastLogin,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // NOT ERROR, just not found
	}
	if err != nil {
		return nil, fmt.Errorf("querying character %d: %w", characterID, err)
	}

	// Создаём Player через NewPlayer (с валидацией)
	// Phase 4.15: Generate unique objectID for player
	objectID := world.IDGenerator().NextPlayerID()
	player, err := model.NewPlayer(objectID, characterIDDB, accountIDDB, name, level, raceID, classID)
	if err != nil {
		return nil, fmt.Errorf("creating player model: %w", err)
	}

	// Устанавливаем Location
	loc := model.NewLocation(x, y, z, heading)
	player.SetLocation(loc)

	// Устанавливаем Stats
	player.SetMaxHP(maxHP)
	player.SetMaxMP(maxMP)
	player.SetMaxCP(maxCP)
	player.SetCurrentHP(currentHP)
	player.SetCurrentMP(currentMP)
	player.SetCurrentCP(currentCP)

	// Устанавливаем Experience
	player.SetExperience(experience)

	// Устанавливаем timestamps
	player.SetCreatedAt(createdAt)
	if lastLogin != nil {
		player.SetLastLogin(*lastLogin)
	}

	return player, nil
}

// LoadByAccountName загружает всех персонажей аккаунта по имени аккаунта.
// NOTE: Requires migration 00005_fix_character_account_reference.sql to be applied.
func (r *CharacterRepository) LoadByAccountName(ctx context.Context, accountName string) ([]*model.Player, error) {
	query := `
		SELECT character_id, account_name, name, level, race_id, class_id,
		       x, y, z, heading,
		       current_hp, max_hp, current_mp, max_mp, current_cp, max_cp,
		       experience, created_at, last_login
		FROM characters
		WHERE account_name = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, accountName)
	if err != nil {
		return nil, fmt.Errorf("querying characters for account %s: %w", accountName, err)
	}
	defer rows.Close()

	// Pre-allocate для типичного аккаунта (3-7 персонажей).
	// Capacity 8 покрывает большинство случаев.
	players := make([]*model.Player, 0, 8)

	for rows.Next() {
		var characterIDDB int64
		var accountNameDB string
		var name string
		var level int32
		var raceID int32
		var classID int32
		var x int32
		var y int32
		var z int32
		var heading uint16
		var currentHP int32
		var maxHP int32
		var currentMP int32
		var maxMP int32
		var currentCP int32
		var maxCP int32
		var experience int64
		var createdAt time.Time
		var lastLogin *time.Time // nullable

		err := rows.Scan(
			&characterIDDB, &accountNameDB, &name, &level, &raceID, &classID,
			&x, &y, &z, &heading,
			&currentHP, &maxHP, &currentMP, &maxMP, &currentCP, &maxCP,
			&experience, &createdAt, &lastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning character row: %w", err)
		}

		// Создаём Player (accountID=0 placeholder, будет refactored в Phase 4.7)
		// Phase 4.15: Generate unique objectID for player
		objectID := world.IDGenerator().NextPlayerID()
		player, err := model.NewPlayer(objectID, characterIDDB, 0, name, level, raceID, classID)
		if err != nil {
			return nil, fmt.Errorf("creating player model: %w", err)
		}

		// Устанавливаем Location
		loc := model.NewLocation(x, y, z, heading)
		player.SetLocation(loc)

		// Устанавливаем Stats
		player.SetMaxHP(maxHP)
		player.SetMaxMP(maxMP)
		player.SetMaxCP(maxCP)
		player.SetCurrentHP(currentHP)
		player.SetCurrentMP(currentMP)
		player.SetCurrentCP(currentCP)

		// Устанавливаем Experience
		player.SetExperience(experience)

		// Устанавливаем timestamps
		player.SetCreatedAt(createdAt)
		if lastLogin != nil {
			player.SetLastLogin(*lastLogin)
		}

		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating character rows: %w", err)
	}

	return players, nil
}

// LoadByAccountID загружает всех персонажей аккаунта.
// DEPRECATED: Use LoadByAccountName after migration 00005 is applied.
func (r *CharacterRepository) LoadByAccountID(ctx context.Context, accountID int64) ([]*model.Player, error) {
	query := `
		SELECT character_id, account_id, name, level, race_id, class_id,
		       x, y, z, heading,
		       current_hp, max_hp, current_mp, max_mp, current_cp, max_cp,
		       experience, created_at, last_login
		FROM characters
		WHERE account_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("querying characters for account %d: %w", accountID, err)
	}
	defer rows.Close()

	// Pre-allocate для типичного аккаунта (3-7 персонажей).
	// Capacity 8 покрывает большинство случаев.
	players := make([]*model.Player, 0, 8)

	for rows.Next() {
		var characterIDDB int64
		var accountIDDB int64
		var name string
		var level int32
		var raceID int32
		var classID int32
		var x int32
		var y int32
		var z int32
		var heading uint16
		var currentHP int32
		var maxHP int32
		var currentMP int32
		var maxMP int32
		var currentCP int32
		var maxCP int32
		var experience int64
		var createdAt time.Time
		var lastLogin *time.Time // nullable

		err := rows.Scan(
			&characterIDDB, &accountIDDB, &name, &level, &raceID, &classID,
			&x, &y, &z, &heading,
			&currentHP, &maxHP, &currentMP, &maxMP, &currentCP, &maxCP,
			&experience, &createdAt, &lastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning character row: %w", err)
		}

		// Создаём Player
		// Phase 4.15: Generate unique objectID for player
		objectID := world.IDGenerator().NextPlayerID()
		player, err := model.NewPlayer(objectID, characterIDDB, accountIDDB, name, level, raceID, classID)
		if err != nil {
			return nil, fmt.Errorf("creating player model: %w", err)
		}

		// Устанавливаем Location
		loc := model.NewLocation(x, y, z, heading)
		player.SetLocation(loc)

		// Устанавливаем Stats
		player.SetMaxHP(maxHP)
		player.SetMaxMP(maxMP)
		player.SetMaxCP(maxCP)
		player.SetCurrentHP(currentHP)
		player.SetCurrentMP(currentMP)
		player.SetCurrentCP(currentCP)

		// Устанавливаем Experience
		player.SetExperience(experience)

		// Устанавливаем timestamps
		player.SetCreatedAt(createdAt)
		if lastLogin != nil {
			player.SetLastLogin(*lastLogin)
		}

		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating character rows: %w", err)
	}

	return players, nil
}

// Create создаёт нового персонажа в БД.
func (r *CharacterRepository) Create(ctx context.Context, p *model.Player) error {
	query := `
		INSERT INTO characters (
			account_id, name, level, race_id, class_id,
			x, y, z, heading,
			current_hp, max_hp, current_mp, max_mp, current_cp, max_cp,
			experience
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING character_id, created_at
	`

	loc := p.Location()

	var characterID int64
	var createdAt time.Time

	err := r.db.QueryRow(ctx, query,
		p.AccountID(), p.Name(), p.Level(), p.RaceID(), p.ClassID(),
		loc.X, loc.Y, loc.Z, loc.Heading,
		p.CurrentHP(), p.MaxHP(), p.CurrentMP(), p.MaxMP(), p.CurrentCP(), p.MaxCP(),
		p.Experience(),
	).Scan(&characterID, &createdAt)

	if err != nil {
		return fmt.Errorf("creating character: %w", err)
	}

	// Устанавливаем ID и createdAt который вернула БД
	p.SetCharacterID(characterID)
	p.SetCreatedAt(createdAt)

	return nil
}

// Update сохраняет изменения персонажа в БД.
func (r *CharacterRepository) Update(ctx context.Context, p *model.Player) error {
	query := `
		UPDATE characters
		SET level = $2, x = $3, y = $4, z = $5, heading = $6,
		    current_hp = $7, max_hp = $8, current_mp = $9, max_mp = $10,
		    current_cp = $11, max_cp = $12, experience = $13, last_login = $14
		WHERE character_id = $1
	`

	loc := p.Location()

	// lastLogin может быть zero value — нужно передать NULL
	var lastLogin any = p.LastLogin()
	if p.LastLogin().IsZero() {
		lastLogin = nil
	}

	_, err := r.db.Exec(ctx, query,
		p.CharacterID(), p.Level(),
		loc.X, loc.Y, loc.Z, loc.Heading,
		p.CurrentHP(), p.MaxHP(), p.CurrentMP(), p.MaxMP(),
		p.CurrentCP(), p.MaxCP(), p.Experience(), lastLogin,
	)

	if err != nil {
		return fmt.Errorf("updating character %d: %w", p.CharacterID(), err)
	}

	return nil
}

// UpdateLocation — hot path для movement packets.
// Обновляет только координаты, избегая UPDATE всех полей.
func (r *CharacterRepository) UpdateLocation(ctx context.Context, characterID int64, loc model.Location) error {
	query := `
		UPDATE characters
		SET x = $2, y = $3, z = $4, heading = $5
		WHERE character_id = $1
	`

	_, err := r.db.Exec(ctx, query, characterID, loc.X, loc.Y, loc.Z, loc.Heading)
	if err != nil {
		return fmt.Errorf("updating location for character %d: %w", characterID, err)
	}

	return nil
}

// UpdateStats — hot path для combat packets.
// Обновляет только HP/MP/CP, избегая UPDATE всех полей.
func (r *CharacterRepository) UpdateStats(ctx context.Context, characterID int64, hp, mp, cp int32) error {
	query := `
		UPDATE characters
		SET current_hp = $2, current_mp = $3, current_cp = $4
		WHERE character_id = $1
	`

	_, err := r.db.Exec(ctx, query, characterID, hp, mp, cp)
	if err != nil {
		return fmt.Errorf("updating stats for character %d: %w", characterID, err)
	}

	return nil
}

// Delete удаляет персонажа из БД.
func (r *CharacterRepository) Delete(ctx context.Context, characterID int64) error {
	query := `DELETE FROM characters WHERE character_id = $1`

	result, err := r.db.Exec(ctx, query, characterID)
	if err != nil {
		return fmt.Errorf("deleting character %d: %w", characterID, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("character %d not found", characterID)
	}

	return nil
}
