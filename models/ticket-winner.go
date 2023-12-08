package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TicketWinner struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty"`
	LotteryID primitive.ObjectID   `bson:"lotteryID,omitempty"`
	WinnerIDs []primitive.ObjectID `bson:"WinnerIDs"`
	Price     float64              `bson:"price,omitempty"`
	Date      string               `bson:"date,omitempty"`
	Time      string               `bson:"time,omitempty"`
	Deadline  time.Duration        `bson:"deadline,omitempty"`
}

func (nb *NewBet) SetTicketDeadline(duration time.Duration) {
	nb.Deadline = duration
}

func (nb *NewBet) GetTicketFormattedDeadline() string {
	return nb.Deadline.String()
}
