package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/udisondev/la2go/internal/gameserver/clan"
)

// ClanRepository handles clan persistence to PostgreSQL.
type ClanRepository struct {
	pool *pgxpool.Pool
}

// NewClanRepository creates a new clan repository.
func NewClanRepository(pool *pgxpool.Pool) *ClanRepository {
	return &ClanRepository{pool: pool}
}

// ClanRow represents a clan_data row.
type ClanRow struct {
	ClanID          int32
	ClanName        string
	LeaderID        int64
	ClanLevel       int32
	Reputation      int32
	CrestID         int32
	LargeCrestID    int32
	AllyID          int32
	AllyName        string
	AllyCrestID     int32
	Notice          string
	NoticeEnabled   bool
	Introduction    string
	DissolutionTime int64
}

// ClanMemberRow represents a clan_members row.
type ClanMemberRow struct {
	CharacterID int64
	ClanID      int32
	PledgeType  int32
	PowerGrade  int32
	Title       string
	SponsorID   int64
	Apprentice  int64
}

// SubPledgeRow represents a clan_subpledges row.
type SubPledgeRow struct {
	ClanID      int32
	SubPledgeID int32
	Name        string
	LeaderID    int64
}

// ClanWarRow represents a clan_wars row.
type ClanWarRow struct {
	Clan1ID int32
	Clan2ID int32
}

// ClanSkillRow represents a clan_skills row.
type ClanSkillRow struct {
	ClanID     int32
	SkillID    int32
	SkillLevel int32
}

// ClanPrivRow represents a clan_privs row.
type ClanPrivRow struct {
	ClanID     int32
	PowerGrade int32
	Privs      int32
}

// LoadAllClans loads all clans from the database.
func (r *ClanRepository) LoadAllClans(ctx context.Context) ([]ClanRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT clan_id, clan_name, leader_id, clan_level, reputation,
		        crest_id, large_crest_id, ally_id, ally_name, ally_crest_id,
		        notice, notice_enabled, introduction, dissolution_time
		 FROM clan_data ORDER BY clan_id`)
	if err != nil {
		return nil, fmt.Errorf("query clan_data: %w", err)
	}
	defer rows.Close()

	var result []ClanRow
	for rows.Next() {
		var c ClanRow
		if err := rows.Scan(
			&c.ClanID, &c.ClanName, &c.LeaderID, &c.ClanLevel, &c.Reputation,
			&c.CrestID, &c.LargeCrestID, &c.AllyID, &c.AllyName, &c.AllyCrestID,
			&c.Notice, &c.NoticeEnabled, &c.Introduction, &c.DissolutionTime,
		); err != nil {
			return nil, fmt.Errorf("scan clan_data: %w", err)
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// LoadClanMembers loads all members for a clan.
func (r *ClanRepository) LoadClanMembers(ctx context.Context, clanID int32) ([]ClanMemberRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT cm.character_id, cm.clan_id, cm.pledge_type, cm.power_grade,
		        cm.title, cm.sponsor_id, cm.apprentice
		 FROM clan_members cm WHERE cm.clan_id = $1`, clanID)
	if err != nil {
		return nil, fmt.Errorf("query clan_members: %w", err)
	}
	defer rows.Close()

	var result []ClanMemberRow
	for rows.Next() {
		var m ClanMemberRow
		if err := rows.Scan(
			&m.CharacterID, &m.ClanID, &m.PledgeType, &m.PowerGrade,
			&m.Title, &m.SponsorID, &m.Apprentice,
		); err != nil {
			return nil, fmt.Errorf("scan clan_members: %w", err)
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

// LoadSubPledges loads all sub-pledges for a clan.
func (r *ClanRepository) LoadSubPledges(ctx context.Context, clanID int32) ([]SubPledgeRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT clan_id, sub_pledge_id, name, leader_id
		 FROM clan_subpledges WHERE clan_id = $1`, clanID)
	if err != nil {
		return nil, fmt.Errorf("query clan_subpledges: %w", err)
	}
	defer rows.Close()

	var result []SubPledgeRow
	for rows.Next() {
		var sp SubPledgeRow
		if err := rows.Scan(&sp.ClanID, &sp.SubPledgeID, &sp.Name, &sp.LeaderID); err != nil {
			return nil, fmt.Errorf("scan clan_subpledges: %w", err)
		}
		result = append(result, sp)
	}
	return result, rows.Err()
}

// LoadClanWars loads all wars.
func (r *ClanRepository) LoadClanWars(ctx context.Context) ([]ClanWarRow, error) {
	rows, err := r.pool.Query(ctx, `SELECT clan1_id, clan2_id FROM clan_wars`)
	if err != nil {
		return nil, fmt.Errorf("query clan_wars: %w", err)
	}
	defer rows.Close()

	var result []ClanWarRow
	for rows.Next() {
		var w ClanWarRow
		if err := rows.Scan(&w.Clan1ID, &w.Clan2ID); err != nil {
			return nil, fmt.Errorf("scan clan_wars: %w", err)
		}
		result = append(result, w)
	}
	return result, rows.Err()
}

// LoadClanSkills loads all skills for a clan.
func (r *ClanRepository) LoadClanSkills(ctx context.Context, clanID int32) ([]ClanSkillRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT clan_id, skill_id, skill_level FROM clan_skills WHERE clan_id = $1`, clanID)
	if err != nil {
		return nil, fmt.Errorf("query clan_skills: %w", err)
	}
	defer rows.Close()

	var result []ClanSkillRow
	for rows.Next() {
		var s ClanSkillRow
		if err := rows.Scan(&s.ClanID, &s.SkillID, &s.SkillLevel); err != nil {
			return nil, fmt.Errorf("scan clan_skills: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

// LoadClanPrivileges loads all rank privileges for a clan.
func (r *ClanRepository) LoadClanPrivileges(ctx context.Context, clanID int32) ([]ClanPrivRow, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT clan_id, power_grade, privs FROM clan_privs WHERE clan_id = $1`, clanID)
	if err != nil {
		return nil, fmt.Errorf("query clan_privs: %w", err)
	}
	defer rows.Close()

	var result []ClanPrivRow
	for rows.Next() {
		var p ClanPrivRow
		if err := rows.Scan(&p.ClanID, &p.PowerGrade, &p.Privs); err != nil {
			return nil, fmt.Errorf("scan clan_privs: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

// SaveClan inserts or updates a clan_data row.
func (r *ClanRepository) SaveClan(ctx context.Context, c *clan.Clan) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO clan_data
		 (clan_id, clan_name, leader_id, clan_level, reputation,
		  crest_id, large_crest_id, ally_id, ally_name, ally_crest_id,
		  notice, notice_enabled, introduction, dissolution_time)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		 ON CONFLICT (clan_id) DO UPDATE SET
		  leader_id=$3, clan_level=$4, reputation=$5,
		  crest_id=$6, large_crest_id=$7, ally_id=$8,
		  ally_name=$9, ally_crest_id=$10,
		  notice=$11, notice_enabled=$12, introduction=$13,
		  dissolution_time=$14`,
		c.ID(), c.Name(), c.LeaderID(), c.Level(), c.Reputation(),
		c.CrestID(), c.LargeCrestID(), c.AllyID(), c.AllyName(), c.AllyCrestID(),
		c.Notice(), c.NoticeEnabled(), c.IntroductionMessage(), c.DissolutionTime(),
	)
	if err != nil {
		return fmt.Errorf("save clan %d: %w", c.ID(), err)
	}
	return nil
}

// SaveClanMember inserts or updates a clan_members row.
func (r *ClanRepository) SaveClanMember(ctx context.Context, clanID int32, m *clan.Member) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO clan_members
		 (character_id, clan_id, pledge_type, power_grade, title, sponsor_id, apprentice)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (character_id) DO UPDATE SET
		  clan_id=$2, pledge_type=$3, power_grade=$4, title=$5, sponsor_id=$6, apprentice=$7`,
		m.PlayerID(), clanID, m.PledgeType(), m.PowerGrade(),
		m.Title(), m.SponsorID(), m.Apprentice(),
	)
	if err != nil {
		return fmt.Errorf("save clan member %d: %w", m.PlayerID(), err)
	}
	return nil
}

// DeleteClanMember removes a member from clan_members.
func (r *ClanRepository) DeleteClanMember(ctx context.Context, characterID int64) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM clan_members WHERE character_id = $1`, characterID)
	if err != nil {
		return fmt.Errorf("delete clan member %d: %w", characterID, err)
	}
	return nil
}

// SaveSubPledge inserts or updates a sub-pledge.
func (r *ClanRepository) SaveSubPledge(ctx context.Context, clanID int32, sp *clan.SubPledge) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO clan_subpledges (clan_id, sub_pledge_id, name, leader_id)
		 VALUES ($1,$2,$3,$4)
		 ON CONFLICT (clan_id, sub_pledge_id) DO UPDATE SET name=$3, leader_id=$4`,
		clanID, sp.ID, sp.Name, sp.LeaderID,
	)
	if err != nil {
		return fmt.Errorf("save sub-pledge %d/%d: %w", clanID, sp.ID, err)
	}
	return nil
}

// SaveClanWar inserts a war record.
func (r *ClanRepository) SaveClanWar(ctx context.Context, clan1ID, clan2ID int32) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO clan_wars (clan1_id, clan2_id) VALUES ($1,$2)
		 ON CONFLICT (clan1_id, clan2_id) DO NOTHING`,
		clan1ID, clan2ID,
	)
	if err != nil {
		return fmt.Errorf("save clan war %d vs %d: %w", clan1ID, clan2ID, err)
	}
	return nil
}

// DeleteClanWar removes a war record.
func (r *ClanRepository) DeleteClanWar(ctx context.Context, clan1ID, clan2ID int32) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM clan_wars WHERE clan1_id = $1 AND clan2_id = $2`,
		clan1ID, clan2ID,
	)
	if err != nil {
		return fmt.Errorf("delete clan war %d vs %d: %w", clan1ID, clan2ID, err)
	}
	return nil
}

// SaveClanSkill inserts or updates a clan skill.
func (r *ClanRepository) SaveClanSkill(ctx context.Context, clanID, skillID, skillLevel int32) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO clan_skills (clan_id, skill_id, skill_level) VALUES ($1,$2,$3)
		 ON CONFLICT (clan_id, skill_id) DO UPDATE SET skill_level=$3`,
		clanID, skillID, skillLevel,
	)
	if err != nil {
		return fmt.Errorf("save clan skill %d/%d: %w", clanID, skillID, err)
	}
	return nil
}

// SaveClanPrivileges inserts or updates rank privileges.
func (r *ClanRepository) SaveClanPrivileges(ctx context.Context, clanID, powerGrade, privs int32) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO clan_privs (clan_id, power_grade, privs) VALUES ($1,$2,$3)
		 ON CONFLICT (clan_id, power_grade) DO UPDATE SET privs=$3`,
		clanID, powerGrade, privs,
	)
	if err != nil {
		return fmt.Errorf("save clan privs %d/grade=%d: %w", clanID, powerGrade, err)
	}
	return nil
}

// DeleteClan deletes a clan and all related data (cascading).
func (r *ClanRepository) DeleteClan(ctx context.Context, clanID int32) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM clan_data WHERE clan_id = $1`, clanID)
	if err != nil {
		return fmt.Errorf("delete clan %d: %w", clanID, err)
	}
	return nil
}

// UpdateCharacterClanID updates the clan_id field on the characters table.
func (r *ClanRepository) UpdateCharacterClanID(ctx context.Context, characterID int64, clanID int32) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE characters SET clan_id = $1 WHERE character_id = $2`,
		clanID, characterID,
	)
	if err != nil {
		return fmt.Errorf("update character %d clan_id: %w", characterID, err)
	}
	return nil
}

// MaxClanID returns the highest clan_id in the database (for nextID on startup).
func (r *ClanRepository) MaxClanID(ctx context.Context) (int32, error) {
	var maxID *int32
	err := r.pool.QueryRow(ctx, `SELECT MAX(clan_id) FROM clan_data`).Scan(&maxID)
	if err != nil {
		return 0, fmt.Errorf("max clan_id: %w", err)
	}
	if maxID == nil {
		return 0, nil
	}
	return *maxID, nil
}

// SaveClanBatch saves the full clan state in a single transaction.
func (r *ClanRepository) SaveClanBatch(ctx context.Context, c *clan.Clan) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	// Save clan data.
	if _, err := tx.Exec(ctx,
		`INSERT INTO clan_data
		 (clan_id, clan_name, leader_id, clan_level, reputation,
		  crest_id, large_crest_id, ally_id, ally_name, ally_crest_id,
		  notice, notice_enabled, introduction, dissolution_time)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		 ON CONFLICT (clan_id) DO UPDATE SET
		  leader_id=$3, clan_level=$4, reputation=$5,
		  crest_id=$6, large_crest_id=$7, ally_id=$8,
		  ally_name=$9, ally_crest_id=$10,
		  notice=$11, notice_enabled=$12, introduction=$13,
		  dissolution_time=$14`,
		c.ID(), c.Name(), c.LeaderID(), c.Level(), c.Reputation(),
		c.CrestID(), c.LargeCrestID(), c.AllyID(), c.AllyName(), c.AllyCrestID(),
		c.Notice(), c.NoticeEnabled(), c.IntroductionMessage(), c.DissolutionTime(),
	); err != nil {
		return fmt.Errorf("save clan data: %w", err)
	}

	// Save members.
	members := c.Members()
	if len(members) > 0 {
		batch := &pgx.Batch{}
		for _, m := range members {
			batch.Queue(
				`INSERT INTO clan_members
				 (character_id, clan_id, pledge_type, power_grade, title, sponsor_id, apprentice)
				 VALUES ($1,$2,$3,$4,$5,$6,$7)
				 ON CONFLICT (character_id) DO UPDATE SET
				  clan_id=$2, pledge_type=$3, power_grade=$4, title=$5, sponsor_id=$6, apprentice=$7`,
				m.PlayerID(), c.ID(), m.PledgeType(), m.PowerGrade(),
				m.Title(), m.SponsorID(), m.Apprentice(),
			)
		}
		br := tx.SendBatch(ctx, batch)
		for range members {
			if _, err := br.Exec(); err != nil {
				br.Close() //nolint:errcheck
				return fmt.Errorf("save clan member batch: %w", err)
			}
		}
		if err := br.Close(); err != nil {
			return fmt.Errorf("close member batch: %w", err)
		}
	}

	// Save skills.
	skills := c.Skills()
	for skillID, level := range skills {
		if _, err := tx.Exec(ctx,
			`INSERT INTO clan_skills (clan_id, skill_id, skill_level) VALUES ($1,$2,$3)
			 ON CONFLICT (clan_id, skill_id) DO UPDATE SET skill_level=$3`,
			c.ID(), skillID, level,
		); err != nil {
			return fmt.Errorf("save clan skill %d: %w", skillID, err)
		}
	}

	// Save rank privileges.
	privs := c.AllRankPrivileges()
	for grade, priv := range privs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO clan_privs (clan_id, power_grade, privs) VALUES ($1,$2,$3)
			 ON CONFLICT (clan_id, power_grade) DO UPDATE SET privs=$3`,
			c.ID(), grade, int32(priv),
		); err != nil {
			return fmt.Errorf("save clan privs grade %d: %w", grade, err)
		}
	}

	return tx.Commit(ctx)
}
