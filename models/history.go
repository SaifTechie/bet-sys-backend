package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type History struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	UserID          primitive.ObjectID `bson:"userID,omitempty"`
	LotteryID       primitive.ObjectID `bson:"lotteryID,omitempty"`
	SelectedNumbers []int              `bson:"selectedNumbers,omitempty"`
	IsWinner        bool               `bson:"isWinner,default:false"`
	Timestamp       time.Time          `bson:"timestamp,omitempty"`
}
