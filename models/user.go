package models

import "time"

// User represents a user of this web site
type User struct {
	ID         string    `bson:"_id"`
	Username   string    `bson:"username"`
	Groups     []string  `bson:"groups"`
	CreatedBy  string    `bson:"createdBy"`
	CreatedAt  time.Time `bson:"createdAt"`
	ModifiedBy string    `bson:"modifiedBy"`
	ModifiedAt time.Time `bson:"modifiedAt"`
}
