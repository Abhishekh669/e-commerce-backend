package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductReviewReaction represents reactions (like 👍 ❤️ etc.) to a review

// ProductReviewReply represents a reply to a review
type ProductReviewReply struct {
	ReplyID   primitive.ObjectID `bson:"replyId,omitempty" json:"replyId"`
	Message   string             `bson:"message" json:"message"`
	UserId    string             `bson:"userId" json:"userId"`
	UserName  string             `bson:"userName" json:"userName"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

// ProductReview represents a user's review on a product
type ProductReview struct {
	ID        primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	ProductId string               `bson:"productId" json:"productId"`
	UserId    string               `bson:"userId" json:"userId"`
	UserName  string               `bson:"userName" json:"userName"`
	Rating    int                  `bson:"rating" json:"rating"` // e.g. 1-5 stars
	Comment   string               `bson:"comment" json:"comment"`
	CreatedAt time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time            `bson:"updatedAt" json:"updatedAt"`
	Replies   []ProductReviewReply `bson:"replies,omitempty" json:"replies,omitempty"`
}

type ProductReviewFromClient struct {
	ProductId string `bson:"productId" json:"productId"`
	UserId    string `bson:"userId" json:"userId"`
	UserName  string `bson:"userName" json:"userName"`
	Rating    int    `bson:"rating" json:"rating"` // e.g. 1-5 stars
	Comment   string `bson:"comment" json:"comment"`
}
