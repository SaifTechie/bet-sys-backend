package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Username  string             `bson:"username,omitempty"`
	Email     string             `bson:"email,omitempty"`
	CreatedAt time.Time          `bson:"createdAt,omitempty"`
}
