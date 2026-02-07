package model

import "time"

// Account represents a player account stored in the database.
type Account struct {
	Login        string
	PasswordHash string
	AccessLevel  int
	LastServer   int
	LastIP       string
	LastActive   time.Time
}
