package models

import "time"

// Role represents the level of permissions a user has on the web site.
type Role struct {
	ID          string    `bson:"_id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	CreatedBy   string    `bson:"createdBy"`
	CreatedAt   time.Time `bson:"createdAt"`
	ModifiedBy  string    `bson:"modifiedBy"`
	ModifiedAt  time.Time `bson:"modifiedAt"`
}
