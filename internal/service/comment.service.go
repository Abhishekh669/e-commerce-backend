package service

import (
	"context"
	"fmt"
	"time"

	"e-commerce.com/internal/models"
	"e-commerce.com/internal/repository"
)

type CommentService interface {
	CreateNewComment(ctx context.Context, userData *models.ProductReviewFromClient) error
}

type commentService struct {
	commentRepo repository.CommentRepo
}

func (s *commentService) CreateNewComment(ctx context.Context, userData *models.ProductReviewFromClient) error {
	time := time.Now()
	newData := models.ProductReview{
		ProductId: userData.ProductId,
		UserId:    userData.UserId,
		UserName:  userData.UserName,
		Rating:    userData.Rating,
		Comment:   userData.Comment,
		CreatedAt: time,
		UpdatedAt: time,
		Replies:   nil,
	}

	fmt.Println("this is new data for comment : ", newData.ProductId)

	err := s.commentRepo.CreateComment(ctx, &newData)
	if err != nil {
		fmt.Println("failed to create in service section ", err)
		return fmt.Errorf("failed to create comment")
	}
	return nil
}

func NewCommnetService(commentRepo repository.CommentRepo) CommentService {
	return &commentService{
		commentRepo: commentRepo,
	}
}
