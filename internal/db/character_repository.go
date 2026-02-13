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
		       experience, sp, created_at, last_login,
		       COALESCE(title, ''), COALESCE(base_class, 0),
		       COALESCE(pvpkills, 0), COALESCE(nobless, false), COALESCE(hero, false),
		       COALESCE(accesslevel, 0), COALESCE(name_color, 16777215), COALESCE(title_color, 15530402),
		       COALESCE(rec_have, 0), COALESCE(rec_left, 0),
		       COALESCE(karma, 0), COALESCE(pk_kills, 0)
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
	var sp int64
	var createdAt time.Time
	var lastLogin *time.Time // nullable
	var title string
	var baseClass int32
	var pvpKills int32
	var nobless bool
	var hero bool
	var accessLevel int32
	var nameColor int32
	var titleColor int32
	var recHave int32
	var recLeft int32
	var karma int32
	var pkKills int32

	err := r.db.QueryRow(ctx, query, characterID).Scan(
		&characterIDDB, &accountIDDB, &name, &level, &raceID, &classID,
		&x, &y, &z, &heading,
		&currentHP, &maxHP, &currentMP, &maxMP, &currentCP, &maxCP,
		&experience, &sp, &createdAt, &lastLogin,
		&title, &baseClass,
		&pvpKills, &nobless, &hero,
		&accessLevel, &nameColor, &titleColor,
		&recHave, &recLeft,
		&karma, &pkKills,
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

	// Устанавливаем Experience и SP
	player.SetExperience(experience)
	player.SetSP(sp)

	// Устанавливаем timestamps
	player.SetCreatedAt(createdAt)
	if lastLogin != nil {
		player.SetLastLogin(*lastLogin)
	}

	// Phase 34: Extended character fields
	player.SetTitle(title)
	if baseClass != 0 {
		player.SetBaseClassID(baseClass)
	}
	player.SetPvPKills(pvpKills)
	player.SetNoble(nobless)
	player.SetHero(hero)
	player.SetAccessLevel(accessLevel)
	player.SetNameColor(nameColor)
	player.SetTitleColor(titleColor)
	player.SetRecomHave(recHave)
	player.SetRecomLeft(recLeft)
	player.SetKarma(karma)
	player.SetPKKills(pkKills)

	return player, nil
}

// LoadByAccountName загружает всех персонажей аккаунта по имени аккаунта.
// NOTE: Requires migration 00005_fix_character_account_reference.sql to be applied.
func (r *CharacterRepository) LoadByAccountName(ctx context.Context, accountName string) ([]*model.Player, error) {
	query := `
		SELECT character_id, account_name, name, level, race_id, class_id,
		       x, y, z, heading,
		       current_hp, max_hp, current_mp, max_mp, current_cp, max_cp,
		       experience, sp, created_at, last_login,
		       COALESCE(title, ''), COALESCE(base_class, 0),
		       COALESCE(pvpkills, 0), COALESCE(nobless, false), COALESCE(hero, false),
		       COALESCE(accesslevel, 0), COALESCE(name_color, 16777215), COALESCE(title_color, 15530402),
		       COALESCE(rec_have, 0), COALESCE(rec_left, 0),
		       COALESCE(karma, 0), COALESCE(pk_kills, 0)
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
		var sp int64
		var createdAt time.Time
		var lastLogin *time.Time // nullable
		var title string
		var baseClass int32
		var pvpKills int32
		var nobless bool
		var hero bool
		var accessLevel int32
		var nameColor int32
		var titleColor int32
		var recHave int32
		var recLeft int32
		var karma int32
		var pkKills int32

		err := rows.Scan(
			&characterIDDB, &accountNameDB, &name, &level, &raceID, &classID,
			&x, &y, &z, &heading,
			&currentHP, &maxHP, &currentMP, &maxMP, &currentCP, &maxCP,
			&experience, &sp, &createdAt, &lastLogin,
			&title, &baseClass,
			&pvpKills, &nobless, &hero,
			&accessLevel, &nameColor, &titleColor,
			&recHave, &recLeft,
			&karma, &pkKills,
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

		// Устанавливаем Experience и SP
		player.SetExperience(experience)
		player.SetSP(sp)

		// Устанавливаем timestamps
		player.SetCreatedAt(createdAt)
		if lastLogin != nil {
			player.SetLastLogin(*lastLogin)
		}

		// Phase 34: Extended character fields
		player.SetTitle(title)
		if baseClass != 0 {
			player.SetBaseClassID(baseClass)
		}
		player.SetPvPKills(pvpKills)
		player.SetNoble(nobless)
		player.SetHero(hero)
		player.SetAccessLevel(accessLevel)
		player.SetNameColor(nameColor)
		player.SetTitleColor(titleColor)
		player.SetRecomHave(recHave)
		player.SetRecomLeft(recLeft)
		player.SetKarma(karma)
		player.SetPKKills(pkKills)

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
		       experience, sp, created_at, last_login
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
		var sp int64
		var createdAt time.Time
		var lastLogin *time.Time // nullable

		err := rows.Scan(
			&characterIDDB, &accountIDDB, &name, &level, &raceID, &classID,
			&x, &y, &z, &heading,
			&currentHP, &maxHP, &currentMP, &maxMP, &currentCP, &maxCP,
			&experience, &sp, &createdAt, &lastLogin,
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

		// Устанавливаем Experience и SP
		player.SetExperience(experience)
		player.SetSP(sp)

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
func (r *CharacterRepository) Create(ctx context.Context, accountName string, p *model.Player) error {
	query := `
		INSERT INTO characters (
			account_name, name, level, race_id, class_id,
			x, y, z, heading,
			current_hp, max_hp, current_mp, max_mp, current_cp, max_cp,
			experience, sp,
			sex, face, hair_style, hair_color
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING character_id, created_at
	`

	loc := p.Location()

	var sex int16
	if p.IsFemale() {
		sex = 1
	}

	var characterID int64
	var createdAt time.Time

	err := r.db.QueryRow(ctx, query,
		accountName, p.Name(), p.Level(), p.RaceID(), p.ClassID(),
		loc.X, loc.Y, loc.Z, loc.Heading,
		p.CurrentHP(), p.MaxHP(), p.CurrentMP(), p.MaxMP(), p.CurrentCP(), p.MaxCP(),
		p.Experience(), p.SP(),
		sex, int16(p.Face()), int16(p.HairStyle()), int16(p.HairColor()),
	).Scan(&characterID, &createdAt)

	if err != nil {
		return fmt.Errorf("creating character: %w", err)
	}

	// Устанавливаем ID и createdAt который вернула БД
	p.SetCharacterID(characterID)
	p.SetCreatedAt(createdAt)

	return nil
}

// NameExists checks if a character name already exists (case-insensitive).
func (r *CharacterRepository) NameExists(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM characters WHERE LOWER(name) = LOWER($1))`

	var exists bool
	if err := r.db.QueryRow(ctx, query, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("checking name existence %q: %w", name, err)
	}

	return exists, nil
}

// CountByAccountName returns the number of characters for an account.
func (r *CharacterRepository) CountByAccountName(ctx context.Context, accountName string) (int, error) {
	query := `SELECT COUNT(*) FROM characters WHERE account_name = $1`

	var count int
	if err := r.db.QueryRow(ctx, query, accountName).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting characters for account %s: %w", accountName, err)
	}

	return count, nil
}

// Update сохраняет изменения персонажа в БД.
func (r *CharacterRepository) Update(ctx context.Context, p *model.Player) error {
	query := `
		UPDATE characters
		SET level = $2, x = $3, y = $4, z = $5, heading = $6,
		    current_hp = $7, max_hp = $8, current_mp = $9, max_mp = $10,
		    current_cp = $11, max_cp = $12, experience = $13, sp = $14, last_login = $15,
		    title = $16, base_class = $17, pvpkills = $18, nobless = $19, hero = $20,
		    accesslevel = $21, name_color = $22, title_color = $23,
		    rec_have = $24, rec_left = $25, karma = $26, pk_kills = $27
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
		p.CurrentCP(), p.MaxCP(), p.Experience(), p.SP(), lastLogin,
		p.Title(), p.BaseClassID(), p.PvPKills(), p.IsNoble(), p.IsHero(),
		p.AccessLevel(), p.NameColor(), p.TitleColor(),
		p.RecomHave(), p.RecomLeft(), p.Karma(), p.PKKills(),
	)

	if err != nil {
		return fmt.Errorf("updating character %d: %w", p.CharacterID(), err)
	}

	return nil
}

// UpdateTx saves character changes within an existing transaction.
func (r *CharacterRepository) UpdateTx(ctx context.Context, tx pgx.Tx, p *model.Player) error {
	query := `
		UPDATE characters
		SET level = $2, x = $3, y = $4, z = $5, heading = $6,
		    current_hp = $7, max_hp = $8, current_mp = $9, max_mp = $10,
		    current_cp = $11, max_cp = $12, experience = $13, sp = $14, last_login = $15,
		    title = $16, base_class = $17, pvpkills = $18, nobless = $19, hero = $20,
		    accesslevel = $21, name_color = $22, title_color = $23,
		    rec_have = $24, rec_left = $25, karma = $26, pk_kills = $27
		WHERE character_id = $1
	`

	loc := p.Location()

	var lastLogin any = p.LastLogin()
	if p.LastLogin().IsZero() {
		lastLogin = nil
	}

	_, err := tx.Exec(ctx, query,
		p.CharacterID(), p.Level(),
		loc.X, loc.Y, loc.Z, loc.Heading,
		p.CurrentHP(), p.MaxHP(), p.CurrentMP(), p.MaxMP(),
		p.CurrentCP(), p.MaxCP(), p.Experience(), p.SP(), lastLogin,
		p.Title(), p.BaseClassID(), p.PvPKills(), p.IsNoble(), p.IsHero(),
		p.AccessLevel(), p.NameColor(), p.TitleColor(),
		p.RecomHave(), p.RecomLeft(), p.Karma(), p.PKKills(),
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

// MarkForDeletion sets the delete timer on a character.
// deleteTimerMs is the future time (Unix ms) when the character will be permanently deleted.
func (r *CharacterRepository) MarkForDeletion(ctx context.Context, characterID int64, deleteTimerMs int64) error {
	query := `UPDATE characters SET delete_timer = $2 WHERE character_id = $1`

	_, err := r.db.Exec(ctx, query, characterID, deleteTimerMs)
	if err != nil {
		return fmt.Errorf("marking character %d for deletion: %w", characterID, err)
	}

	return nil
}

// RestoreCharacter clears the delete timer on a character (cancels pending deletion).
func (r *CharacterRepository) RestoreCharacter(ctx context.Context, characterID int64) error {
	query := `UPDATE characters SET delete_timer = 0 WHERE character_id = $1`

	_, err := r.db.Exec(ctx, query, characterID)
	if err != nil {
		return fmt.Errorf("restoring character %d: %w", characterID, err)
	}

	return nil
}

// GetClanID returns the clan_id for a character (0 if not in a clan).
func (r *CharacterRepository) GetClanID(ctx context.Context, characterID int64) (int64, error) {
	query := `SELECT COALESCE(clan_id, 0) FROM characters WHERE character_id = $1`

	var clanID int64
	if err := r.db.QueryRow(ctx, query, characterID).Scan(&clanID); err != nil {
		return 0, fmt.Errorf("getting clan_id for character %d: %w", characterID, err)
	}

	return clanID, nil
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
