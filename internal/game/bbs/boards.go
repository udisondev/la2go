package bbs

import (
	"fmt"
	"strings"
	"time"
)

// HomeBoard handles the main community board page.
// Commands: _bbshome, _bbstop
//
// Java reference: HomeBoard.java
type HomeBoard struct{}

func (b *HomeBoard) Commands() []string {
	return []string{"_bbshome", "_bbstop"}
}

func (b *HomeBoard) OnCommand(cmd string, charID int64, charName string) string {
	if cmd == "_bbshome" || cmd == "_bbstop" {
		return homeHTML(charName)
	}

	// _bbstop;page.html — навигация по подстраницам
	if page, ok := strings.CutPrefix(cmd, "_bbstop;"); ok {
		if page != "" && strings.HasSuffix(page, ".html") {
			return fmt.Sprintf(topicPageHTML, page)
		}
	}

	return homeHTML(charName)
}

// RegionBoard handles region information (castles, clan halls).
// Commands: _bbsloc
//
// Java reference: RegionBoard.java
type RegionBoard struct{}

func (b *RegionBoard) Commands() []string {
	return []string{"_bbsloc"}
}

func (b *RegionBoard) OnCommand(cmd string, charID int64, charName string) string {
	if cmd == "_bbsloc" {
		return regionListHTML()
	}

	// _bbsloc;N — детали региона
	if regionID, ok := strings.CutPrefix(cmd, "_bbsloc;"); ok {
		return regionDetailHTML(regionID)
	}

	return regionListHTML()
}

// ClanBoard handles clan-related community board pages.
// Commands: _bbsclan
//
// Java reference: ClanBoard.java
type ClanBoard struct{}

func (b *ClanBoard) Commands() []string {
	return []string{"_bbsclan"}
}

func (b *ClanBoard) OnCommand(cmd string, charID int64, charName string) string {
	return clanHTML()
}

// MemoBoard handles personal notes (memo).
// Commands: _bbsmemo, _bbstopics
//
// Java reference: MemoBoard.java (uses TopicBBSManager)
type MemoBoard struct{}

func (b *MemoBoard) Commands() []string {
	return []string{"_bbsmemo", "_bbstopics"}
}

func (b *MemoBoard) OnCommand(cmd string, charID int64, charName string) string {
	return memoHTML(charName)
}

// MailBoard handles in-game mail.
// Commands: _bbsmail
//
// Java reference: MailBoard.java
type MailBoard struct{}

func (b *MailBoard) Commands() []string {
	return []string{"_bbsmail"}
}

func (b *MailBoard) OnCommand(cmd string, charID int64, charName string) string {
	return mailHTML(charName)
}

// FriendsBoard handles friends list.
// Commands: _bbsfriends
//
// Java reference: FriendsBoard.java
type FriendsBoard struct{}

func (b *FriendsBoard) Commands() []string {
	return []string{"_bbsfriends"}
}

func (b *FriendsBoard) OnCommand(cmd string, charID int64, charName string) string {
	return friendsHTML(charName)
}

// FavoriteBoard handles bookmarks/favorites.
// Commands: _bbsgetfav, bbs_add_fav
//
// Java reference: FavoriteBoard.java
type FavoriteBoard struct{}

func (b *FavoriteBoard) Commands() []string {
	return []string{"_bbsgetfav", "bbs_add_fav"}
}

func (b *FavoriteBoard) OnCommand(cmd string, charID int64, charName string) string {
	if cmd == "bbs_add_fav" {
		return favoriteAddedHTML()
	}
	return favoritesHTML(charName)
}

// --- HTML templates ---

func homeHTML(charName string) string {
	return fmt.Sprintf(`<html><body>
<br>
<center>
<table width=610>
<tr><td align=center><font color="LEVEL">Welcome to Community Board</font></td></tr>
</table>
<br>
<table width=610>
<tr><td>Hello, %s!</td></tr>
<tr><td><br>Server time: %s</td></tr>
<tr><td><br>Use the navigation buttons above to browse.</td></tr>
</table>
</center>
</body></html>`, charName, time.Now().Format("2006-01-02 15:04:05"))
}

const topicPageHTML = `<html><body>
<br>
<center>
<table width=610>
<tr><td align=center><font color="LEVEL">Page: %s</font></td></tr>
</table>
</center>
</body></html>`

func regionListHTML() string {
	// Стандартный список из 9 регионов (замков Interlude)
	regions := []string{
		"Gludio", "Dion", "Giran", "Oren", "Aden",
		"Innadril", "Goddard", "Rune", "Schuttgart",
	}

	var sb strings.Builder
	sb.WriteString(`<html><body><br><center>`)
	sb.WriteString(`<table width=610><tr><td align=center><font color="LEVEL">Region Information</font></td></tr></table><br>`)
	sb.WriteString(`<table width=610 bgcolor=000000>`)
	sb.WriteString(`<tr><td width=5></td><td width=200>Region</td><td width=200>Lord</td><td width=200>Tax</td><td width=5></td></tr></table>`)

	for i, name := range regions {
		sb.WriteString(fmt.Sprintf(`<table width=610><tr><td width=5></td><td width=200><a action="bypass _bbsloc;%d">%s</a></td><td width=200>NPC</td><td width=200>0%%</td><td width=5></td></tr></table>`, i, name))
	}

	sb.WriteString(`</center></body></html>`)
	return sb.String()
}

func regionDetailHTML(regionID string) string {
	return fmt.Sprintf(`<html><body>
<br>
<center>
<table width=610>
<tr><td align=center><font color="LEVEL">Region #%s Details</font></td></tr>
</table>
<br>
<table width=610>
<tr><td>Lord: NPC</td></tr>
<tr><td>Tax Rate: 0%%</td></tr>
<tr><td>Alliance: None</td></tr>
</table>
<br>
<table width=610>
<tr><td align=center><a action="bypass _bbsloc">Back to Regions</a></td></tr>
</table>
</center>
</body></html>`, regionID)
}

func clanHTML() string {
	return `<html><body>
<br>
<center>
<table width=610>
<tr><td align=center><font color="LEVEL">Clan Notice Board</font></td></tr>
</table>
<br>
<table width=610>
<tr><td>No clan notices available.</td></tr>
</table>
</center>
</body></html>`
}

func memoHTML(charName string) string {
	return fmt.Sprintf(`<html><body>
<br>
<center>
<table width=610>
<tr><td align=center><font color="LEVEL">%s's Memo</font></td></tr>
</table>
<br>
<table width=610>
<tr><td>No memos yet.</td></tr>
</table>
</center>
</body></html>`, charName)
}

func mailHTML(charName string) string {
	return fmt.Sprintf(`<html><body>
<br>
<center>
<table width=610>
<tr><td align=center><font color="LEVEL">%s's Mailbox</font></td></tr>
</table>
<br>
<table width=610>
<tr><td>No mail messages.</td></tr>
</table>
</center>
</body></html>`, charName)
}

func friendsHTML(charName string) string {
	return fmt.Sprintf(`<html><body>
<br>
<center>
<table width=610>
<tr><td align=center><font color="LEVEL">%s's Friends</font></td></tr>
</table>
<br>
<table width=610>
<tr><td>No friends added.</td></tr>
</table>
</center>
</body></html>`, charName)
}

func favoritesHTML(charName string) string {
	return fmt.Sprintf(`<html><body>
<br>
<center>
<table width=610>
<tr><td align=center><font color="LEVEL">%s's Favorites</font></td></tr>
</table>
<br>
<table width=610>
<tr><td>No favorites saved.</td></tr>
</table>
</center>
</body></html>`, charName)
}

func favoriteAddedHTML() string {
	return `<html><body>
<br>
<center>
<table width=610>
<tr><td align=center><font color="LEVEL">Favorite Added</font></td></tr>
</table>
<br>
<table width=610>
<tr><td>Current page has been added to your favorites.</td></tr>
<tr><td><br><a action="bypass _bbsgetfav">View Favorites</a></td></tr>
</table>
</center>
</body></html>`
}
