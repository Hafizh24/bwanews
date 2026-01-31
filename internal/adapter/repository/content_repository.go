package repository

import (
	"bwanews/internal/core/domain/entity"
	"bwanews/internal/core/domain/model"
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContentRepository interface {
	GetContents(ctx context.Context, query entity.QueryString) ([]entity.ContentEntity, int64, int64, error)
	GetContentByID(ctx context.Context, id int64) (*entity.ContentEntity, error)
	CreateContent(ctx context.Context, req entity.ContentEntity) error
	EditContentByID(ctx context.Context, req entity.ContentEntity) error
	DeleteContent(ctx context.Context, id int64) error
}

type contentRepository struct {
	db *gorm.DB
}

// CreateContent implements ContentRepository.
func (c *contentRepository) CreateContent(ctx context.Context, req entity.ContentEntity) error {
	tags := strings.Join(req.Tags, ",")
	modelContent := model.Content{
		Title:       req.Title,
		Excerpt:     req.Excerpt,
		Description: req.Description,
		Image:       req.Image,
		Tags:        tags,
		Status:      req.Status,
		CategoryID:  req.CategoryID,
		CreatedByID: req.CreatedByID,
	}

	err = c.db.Create(&modelContent).Error
	if err != nil {
		code = "[REPOSITORY] CreateContent = 1"
		log.Errorw(code, err)
		return err
	}

	return nil
}

// DeleteContent implements ContentRepository.
func (c *contentRepository) DeleteContent(ctx context.Context, id int64) error {
	err = c.db.Where("id = ?", id).Delete(&model.Content{}).Error
	if err != nil {
		code = "[REPOSITORY] DeleteContent = 1"
		log.Errorw(code, err)
		return err
	}

	return nil
}

// EditContentByID implements ContentRepository.
func (c *contentRepository) EditContentByID(ctx context.Context, req entity.ContentEntity) error {
	tags := strings.Join(req.Tags, ",")
	modelContent := model.Content{
		Title:       req.Title,
		Excerpt:     req.Excerpt,
		Description: req.Description,
		Image:       req.Image,
		Tags:        tags,
		Status:      req.Status,
		CategoryID:  req.CategoryID,
		CreatedByID: req.CreatedByID,
	}

	err = c.db.Where("id = ?", req.ID).Updates(&modelContent).Error
	if err != nil {
		code = "[REPOSITORY] EditContentByID = 1"
		log.Errorw(code, err)
		return err
	}

	return nil
}

// GetContentByID implements ContentRepository.
func (c *contentRepository) GetContentByID(ctx context.Context, id int64) (*entity.ContentEntity, error) {
	var modelContent model.Content

	err = c.db.Where("id = ?", id).Preload(clause.Associations).First(&modelContent).Error
	if err != nil {
		code = "[REPOSITORY] GetContentByID = 1"
		log.Errorw(code, err)
		return nil, err
	}

	tags := strings.Split(modelContent.Tags, ",")
	resp := entity.ContentEntity{
		ID:          modelContent.ID,
		Title:       modelContent.Title,
		Excerpt:     modelContent.Excerpt,
		Description: modelContent.Description,
		Image:       modelContent.Image,
		Tags:        tags,
		Status:      modelContent.Status,
		CategoryID:  modelContent.CategoryID,
		CreatedByID: modelContent.CreatedByID,
		CreatedAt:   modelContent.CreatedAt,
		Category: entity.CategoryEntity{
			ID:    modelContent.Category.ID,
			Title: modelContent.Category.Title,
			Slug:  modelContent.Category.Slug,
		},
		User: entity.UserEntity{
			ID:   modelContent.User.ID,
			Name: modelContent.User.Name,
		},
	}

	return &resp, nil
}

// GetContents implements ContentRepository.
func (c *contentRepository) GetContents(ctx context.Context, query entity.QueryString) ([]entity.ContentEntity, int64, int64, error) {
	var modelContents []model.Content
	var countData int64

	// ✅ ADD: Set default values if not provided
	if query.Limit <= 0 {
		query.Limit = 10 // Default limit
	}

	if query.Page <= 0 {
		query.Page = 1 // Default page
	}

	if query.OrderBy == "" {
		query.OrderBy = "created_at" // Default order
	}

	if query.OrderType == "" {
		query.OrderType = "DESC" // Default order type
	}

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit
	status := ""
	if query.Status != "" {
		status = query.Status
	}

	sqlMain := c.db.Preload(clause.Associations).
		Where("title ilike ? OR excerpt ilike ? OR description ilike ?", "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%").
		Where("status LIKE ?", "%"+status+"%")

	if query.CategoryID > 0 {
		sqlMain = sqlMain.Where("category_id = ?", query.CategoryID)
	}

	err = sqlMain.Model(&modelContents).Count(&countData).Error
	if err != nil {
		code = "[REPOSITORY] GetContents = 1"
		log.Errorw(code, err)
		return nil, 0, 0, err
	}

	totalPages := int(math.Ceil(float64(countData) / float64(query.Limit)))

	err = sqlMain.
		Order(order).
		Limit(query.Limit).
		Offset(offset).
		Find(&modelContents).Error
	if err != nil {
		code = "[REPOSITORY] GetContents = 2"
		log.Errorw(code, err)
		return nil, 0, 0, err
	}

	resps := []entity.ContentEntity{}
	for _, content := range modelContents {
		tags := strings.Split(content.Tags, ",")
		resps = append(resps, entity.ContentEntity{
			ID:          content.ID,
			Title:       content.Title,
			Excerpt:     content.Excerpt,
			Description: content.Description,
			Image:       content.Image,
			Tags:        tags,
			Status:      content.Status,
			CategoryID:  content.CategoryID,
			CreatedByID: content.CreatedByID,
			CreatedAt:   content.CreatedAt,
			Category: entity.CategoryEntity{
				ID:    content.Category.ID,
				Title: content.Category.Title,
				Slug:  content.Category.Slug,
			},
			User: entity.UserEntity{
				ID:   content.User.ID,
				Name: content.User.Name,
			},
		})

	}

	return resps, countData, int64(totalPages), nil

}

func NewContentRepository(db *gorm.DB) ContentRepository {
	return &contentRepository{db: db}
}
