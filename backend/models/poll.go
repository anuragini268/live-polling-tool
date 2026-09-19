package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Poll struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OwnerID   string        `bson:"owner_id" json:"owner_id"`
	ShareCode string        `bson:"share_code" json:"share_code"`
	Question  string        `bson:"question" json:"question"`
	Options   []string      `bson:"options" json:"options"`
	Votes     []int         `bson:"votes" json:"votes"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

type Vote struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID    bson.ObjectID `bson:"poll_id" json:"poll_id"`
	VoterID   string        `bson:"voter_id" json:"voter_id"`
	Option    int           `bson:"option" json:"option"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}
