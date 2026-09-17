package models

import "go.mongodb.org/mongo-driver/v2/bson"

type Poll struct {
	ID       bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Question string        `bson:"question" json:"question"`
	Options  []string      `bson:"options" json:"options"`
	Votes    []int         `bson:"votes" json:"votes"`
}