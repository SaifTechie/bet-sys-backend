package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NewBet struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	LotteryID primitive.ObjectID `bson:"lotteryID,omitempty"`
	Numbers   []int              `bson:"numbers,omitempty"`
	Price     float64            `bson:"price,omitempty"`
	Date      string             `bson:"date,omitempty"`
	Time      string             `bson:"time,omitempty"`
	MinTicket int                `bson:"minticket,omitempty"`
	Visible   bool               `bson:"visible,omitempty"`
	Deadline  time.Duration      `bson:"deadline,omitempty"`
}

func (nb *NewBet) SetDeadline(duration time.Duration) {
	nb.Deadline = duration
}

func (nb *NewBet) GetFormattedDeadline() string {
	return nb.Deadline.String()
}
