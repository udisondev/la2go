package login

import "math/rand/v2"

// SessionKey holds the 4 random int32 values used as session identifiers.
type SessionKey struct {
	LoginOkID1 int32
	LoginOkID2 int32
	PlayOkID1  int32
	PlayOkID2  int32
}

// NewSessionKey generates a new session key with 4 random int32 values.
func NewSessionKey() SessionKey {
	return SessionKey{
		LoginOkID1: rand.Int32(),
		LoginOkID2: rand.Int32(),
		PlayOkID1:  rand.Int32(),
		PlayOkID2:  rand.Int32(),
	}
}

// CheckLoginPair verifies the loginOk pair matches.
func (sk SessionKey) CheckLoginPair(ok1, ok2 int32) bool {
	return sk.LoginOkID1 == ok1 && sk.LoginOkID2 == ok2
}
