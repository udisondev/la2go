package model

// FriendList returns a copy of all friend ObjectIDs (relation=0).
// Thread-safe: acquires read lock.
func (p *Player) FriendList() []int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	if len(p.friendList) == 0 {
		return nil
	}
	result := make([]int32, 0, len(p.friendList))
	for id := range p.friendList {
		result = append(result, id)
	}
	return result
}

// AddFriend adds a friend to the player's friend list.
// Thread-safe: acquires write lock.
func (p *Player) AddFriend(friendID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.friendList == nil {
		p.friendList = make(map[int32]bool)
	}
	p.friendList[friendID] = true
}

// RemoveFriend removes a friend from the player's friend list.
// Thread-safe: acquires write lock.
func (p *Player) RemoveFriend(friendID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	delete(p.friendList, friendID)
}

// IsFriend returns true if targetID is in the player's friend list.
// Thread-safe: acquires read lock.
func (p *Player) IsFriend(targetID int32) bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	return p.friendList[targetID]
}

// SetFriendList replaces the friend list with given IDs (for DB load).
// Thread-safe: acquires write lock.
func (p *Player) SetFriendList(ids []int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.friendList = make(map[int32]bool, len(ids))
	for _, id := range ids {
		p.friendList[id] = true
	}
}

// BlockList returns a copy of all blocked ObjectIDs (relation=1).
// Thread-safe: acquires read lock.
func (p *Player) BlockList() []int32 {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	if len(p.blockList) == 0 {
		return nil
	}
	result := make([]int32, 0, len(p.blockList))
	for id := range p.blockList {
		result = append(result, id)
	}
	return result
}

// AddBlock adds a player to the block list.
// Thread-safe: acquires write lock.
func (p *Player) AddBlock(targetID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	if p.blockList == nil {
		p.blockList = make(map[int32]bool)
	}
	p.blockList[targetID] = true
}

// RemoveBlock removes a player from the block list.
// Thread-safe: acquires write lock.
func (p *Player) RemoveBlock(targetID int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	delete(p.blockList, targetID)
}

// IsBlocked returns true if targetID is in the player's block list.
// Thread-safe: acquires read lock.
func (p *Player) IsBlocked(targetID int32) bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	return p.blockList[targetID]
}

// SetBlockList replaces the block list with given IDs (for DB load).
// Thread-safe: acquires write lock.
func (p *Player) SetBlockList(ids []int32) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.blockList = make(map[int32]bool, len(ids))
	for _, id := range ids {
		p.blockList[id] = true
	}
}

// MessageRefusal returns true if player has "block all messages" enabled.
// Thread-safe: acquires read lock.
func (p *Player) MessageRefusal() bool {
	p.playerMu.RLock()
	defer p.playerMu.RUnlock()

	return p.messageRefusal
}

// SetMessageRefusal sets the "block all messages" flag.
// Thread-safe: acquires write lock.
func (p *Player) SetMessageRefusal(v bool) {
	p.playerMu.Lock()
	defer p.playerMu.Unlock()

	p.messageRefusal = v
}
