package commands

import (
	"fmt"
	"strings"

	"github.com/udisondev/la2go/internal/model"
)

// Info handles //info — shows information about the selected target.
//
// Java reference: AdminAdmin.java
type Info struct {
	clientMgr ClientManager
}

// NewInfo creates the info command handler.
func NewInfo(clientMgr ClientManager) *Info {
	return &Info{clientMgr: clientMgr}
}

func (c *Info) Names() []string           { return []string{"info", "status"} }
func (c *Info) RequiredAccessLevel() int32 { return 1 }

func (c *Info) Handle(player *model.Player, args []string) error {
	// If name specified, show info about that player
	if len(args) >= 2 {
		target := c.clientMgr.FindPlayerByName(args[1])
		if target == nil {
			return fmt.Errorf("player %q not found", args[1])
		}
		player.SetLastAdminMessage(formatPlayerInfo(target))
		return nil
	}

	// Show info about current target or server status
	target := player.Target()
	if target == nil {
		// Server status
		var b strings.Builder
		b.WriteString("=== Server Status ===\n")
		b.WriteString(fmt.Sprintf("Online: %d players\n", c.clientMgr.PlayerCount()))
		player.SetLastAdminMessage(b.String())
		return nil
	}

	switch d := target.Data.(type) {
	case *model.Player:
		player.SetLastAdminMessage(formatPlayerInfo(d))
	case *model.Monster:
		player.SetLastAdminMessage(formatNpcInfo(d.Npc))
	case *model.Npc:
		player.SetLastAdminMessage(formatNpcInfo(d))
	default:
		loc := target.Location()
		player.SetLastAdminMessage(fmt.Sprintf("Object: %s (ID: %d) at (%d, %d, %d)",
			target.Name(), target.ObjectID(), loc.X, loc.Y, loc.Z))
	}

	return nil
}

func formatPlayerInfo(p *model.Player) string {
	loc := p.Location()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== Player: %s ===\n", p.Name()))
	b.WriteString(fmt.Sprintf("ObjectID: %d, CharID: %d, AccountID: %d\n", p.ObjectID(), p.CharacterID(), p.AccountID()))
	b.WriteString(fmt.Sprintf("Level: %d, Class: %d, Race: %d\n", p.Level(), p.ClassID(), p.RaceID()))
	b.WriteString(fmt.Sprintf("HP: %d/%d, MP: %d/%d, CP: %d/%d\n",
		p.CurrentHP(), p.MaxHP(), p.CurrentMP(), p.MaxMP(), p.CurrentCP(), p.MaxCP()))
	b.WriteString(fmt.Sprintf("XP: %d, SP: %d\n", p.Experience(), p.SP()))
	b.WriteString(fmt.Sprintf("Location: (%d, %d, %d) heading: %d\n", loc.X, loc.Y, loc.Z, loc.Heading))
	b.WriteString(fmt.Sprintf("AccessLevel: %d, IsGM: %v\n", p.AccessLevel(), p.IsGM()))
	b.WriteString(fmt.Sprintf("Dead: %v, InCombat: %v\n", p.IsDead(), p.HasAttackStance()))

	inv := p.Inventory()
	if inv != nil {
		b.WriteString(fmt.Sprintf("Items: %d, Adena: %d\n", inv.Count(), inv.GetAdena()))
	}

	if p.IsInParty() {
		party := p.GetParty()
		if party != nil {
			b.WriteString(fmt.Sprintf("Party: %d members, leader: %s\n", party.MemberCount(), party.Leader().Name()))
		}
	}

	return b.String()
}

func formatNpcInfo(n *model.Npc) string {
	loc := n.Location()
	tmpl := n.Template()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== NPC: %s ===\n", n.Name()))
	b.WriteString(fmt.Sprintf("ObjectID: %d, TemplateID: %d\n", n.ObjectID(), n.TemplateID()))
	b.WriteString(fmt.Sprintf("Level: %d\n", tmpl.Level()))
	b.WriteString(fmt.Sprintf("HP: %d/%d, MP: %d/%d\n",
		n.CurrentHP(), n.MaxHP(), n.CurrentMP(), n.MaxMP()))
	b.WriteString(fmt.Sprintf("PAtk: %d, PDef: %d, MAtk: %d, MDef: %d\n",
		tmpl.PAtk(), tmpl.PDef(), tmpl.MAtk(), tmpl.MDef()))
	b.WriteString(fmt.Sprintf("Location: (%d, %d, %d)\n", loc.X, loc.Y, loc.Z))
	b.WriteString(fmt.Sprintf("Dead: %v\n", n.IsDead()))
	return b.String()
}
