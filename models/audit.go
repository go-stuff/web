package models

import "time"

// Audit is a document that records any inserts, updates
// or deltes a user has made in the system
type Audit struct {
	ID          string    `bson:"_id"`
	UserID      string    `bson:"userid"`
	Username    string    `bson:"username"`
	Action      string    `bson:"action"`
	Description string    `bson:"description"`
	CreatedBy   string    `bson:"createdBy"`
	CreatedAt   time.Time `bson:"createdAt"`
}
