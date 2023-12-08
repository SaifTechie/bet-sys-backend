package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BetSkeleton struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	SkeletonID   int                `bson:"skeletonID,omitempty"`
	Price        float64            `bson:"price,omitempty"`
	MinTicket    int                `bson:"minticket,omitempty"`
	DeadlineTime float64            `bson:"deadlinetime,omitempty"`
}
