package models

import "time"

// Session represents a single session conncected to the web site.
type Session struct {
	ID         string    `bson:"_id"`
	Username   string    `bson:"username"`
	RemoteAddr string    `bson:"remoteaddr"`
	Host       string    `bson:"host"`
	CreatedAt  time.Time `bson:"createdat"`
	ExpiresAt  time.Time `bson:"expiresat"`
}
