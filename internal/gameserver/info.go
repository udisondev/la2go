package gameserver

import (
	"sync"
	"sync/atomic"
)

// GameServerInfo хранит информацию о зарегистрированном GameServer.
type GameServerInfo struct {
	mu sync.RWMutex

	id         int
	hexID      []byte
	port       int
	maxPlayers int
	status     int
	serverType int
	ageLimit   int
	hosts      []string

	// isAuthed использует atomic для visibility между goroutines
	isAuthed atomic.Bool

	// showingBrackets - показывать ли [brackets] в списке серверов
	showingBrackets bool
}

// NewGameServerInfo создаёт новый GameServerInfo.
func NewGameServerInfo(id int, hexID []byte) *GameServerInfo {
	return &GameServerInfo{
		id:    id,
		hexID: hexID,
	}
}

// ID возвращает ID сервера.
func (gsi *GameServerInfo) ID() int {
	gsi.mu.RLock()
	defer gsi.mu.RUnlock()
	return gsi.id
}

// SetID устанавливает ID сервера.
func (gsi *GameServerInfo) SetID(id int) {
	gsi.mu.Lock()
	defer gsi.mu.Unlock()
	gsi.id = id
}

// HexID возвращает hex ID сервера.
func (gsi *GameServerInfo) HexID() []byte {
	gsi.mu.RLock()
	defer gsi.mu.RUnlock()
	return gsi.hexID
}

// IsAuthed возвращает true, если GS аутентифицирован (thread-safe).
func (gsi *GameServerInfo) IsAuthed() bool {
	return gsi.isAuthed.Load()
}

// SetAuthed устанавливает статус аутентификации (thread-safe).
func (gsi *GameServerInfo) SetAuthed(authed bool) {
	gsi.isAuthed.Store(authed)
}

// Port возвращает порт GS для клиентов.
func (gsi *GameServerInfo) Port() int {
	gsi.mu.RLock()
	defer gsi.mu.RUnlock()
	return gsi.port
}

// SetPort устанавливает порт GS.
func (gsi *GameServerInfo) SetPort(port int) {
	gsi.mu.Lock()
	defer gsi.mu.Unlock()
	gsi.port = port
}

// MaxPlayers возвращает максимальное число игроков.
func (gsi *GameServerInfo) MaxPlayers() int {
	gsi.mu.RLock()
	defer gsi.mu.RUnlock()
	return gsi.maxPlayers
}

// SetMaxPlayers устанавливает максимальное число игроков.
func (gsi *GameServerInfo) SetMaxPlayers(max int) {
	gsi.mu.Lock()
	defer gsi.mu.Unlock()
	gsi.maxPlayers = max
}

// Status возвращает текущий статус сервера.
func (gsi *GameServerInfo) Status() int {
	gsi.mu.RLock()
	defer gsi.mu.RUnlock()
	return gsi.status
}

// SetStatus устанавливает статус сервера.
func (gsi *GameServerInfo) SetStatus(status int) {
	gsi.mu.Lock()
	defer gsi.mu.Unlock()
	gsi.status = status
}

// ServerType возвращает тип сервера.
func (gsi *GameServerInfo) ServerType() int {
	gsi.mu.RLock()
	defer gsi.mu.RUnlock()
	return gsi.serverType
}

// SetServerType устанавливает тип сервера.
func (gsi *GameServerInfo) SetServerType(serverType int) {
	gsi.mu.Lock()
	defer gsi.mu.Unlock()
	gsi.serverType = serverType
}

// AgeLimit возвращает возрастное ограничение.
func (gsi *GameServerInfo) AgeLimit() int {
	gsi.mu.RLock()
	defer gsi.mu.RUnlock()
	return gsi.ageLimit
}

// SetAgeLimit устанавливает возрастное ограничение.
func (gsi *GameServerInfo) SetAgeLimit(ageLimit int) {
	gsi.mu.Lock()
	defer gsi.mu.Unlock()
	gsi.ageLimit = ageLimit
}

// Hosts возвращает список хостов (копию).
func (gsi *GameServerInfo) Hosts() []string {
	gsi.mu.RLock()
	defer gsi.mu.RUnlock()
	hosts := make([]string, len(gsi.hosts))
	copy(hosts, gsi.hosts)
	return hosts
}

// SetHosts устанавливает список хостов.
func (gsi *GameServerInfo) SetHosts(hosts []string) {
	gsi.mu.Lock()
	defer gsi.mu.Unlock()
	gsi.hosts = make([]string, len(hosts))
	copy(gsi.hosts, hosts)
}

// ShowingBrackets возвращает флаг показа [brackets].
func (gsi *GameServerInfo) ShowingBrackets() bool {
	gsi.mu.RLock()
	defer gsi.mu.RUnlock()
	return gsi.showingBrackets
}

// SetShowingBrackets устанавливает флаг показа [brackets].
func (gsi *GameServerInfo) SetShowingBrackets(show bool) {
	gsi.mu.Lock()
	defer gsi.mu.Unlock()
	gsi.showingBrackets = show
}

// SetDown помечает сервер как offline.
func (gsi *GameServerInfo) SetDown() {
	gsi.mu.Lock()
	defer gsi.mu.Unlock()
	gsi.isAuthed.Store(false)
	gsi.port = 0
	gsi.status = StatusDown
}
