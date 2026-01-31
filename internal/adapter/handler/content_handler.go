package handler

import (
	"bwanews/internal/adapter/handler/request"
	"bwanews/internal/adapter/handler/response"
	"bwanews/internal/core/domain/entity"
	"bwanews/internal/core/service"
	"bwanews/lib/conv"
	validatorLib "bwanews/lib/validator"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type ContentHandler interface {
	GetContents(c *fiber.Ctx) error
	GetContentByID(c *fiber.Ctx) error
	CreateContent(c *fiber.Ctx) error
	EditContentByID(c *fiber.Ctx) error
	DeleteContent(c *fiber.Ctx) error
	UploadImageR2(c *fiber.Ctx) error

	GetContentWithQuery(c *fiber.Ctx) error
	GetContentDetail(c *fiber.Ctx) error
}

type contentHandler struct {
	contentService service.ContentService
}

// GetContentDetail implements ContentHandler.
func (ch *contentHandler) GetContentDetail(c *fiber.Ctx) error {
	idParam := c.Params("contentID")
	id, err := conv.StringToInt64(idParam)
	if err != nil {
		code := "[HANDLER] GetContentDetail = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	result, err := ch.contentService.GetContentByID(c.Context(), id)
	if err != nil {
		code := "[HANDLER] GetContentDetail = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Success"

	respContent := response.ContentResponse{
		ID:           result.ID,
		Title:        result.Title,
		Excerpt:      result.Excerpt,
		Description:  result.Description,
		Image:        result.Image,
		Tags:         result.Tags,
		Status:       result.Status,
		CategoryID:   result.CategoryID,
		CreatedByID:  result.CreatedByID,
		CreatedAt:    result.CreatedAt.Format(time.RFC3339),
		CategoryName: result.Category.Title,
		Author:       result.User.Name,
	}

	defaultSuccessResponse.Data = respContent

	return c.Status(fiber.StatusOK).JSON(defaultSuccessResponse)
}

// GetContentWithQuery implements ContentHandler.
func (ch *contentHandler) GetContentWithQuery(c *fiber.Ctx) error {
	page := 1
	if c.Query("page") != "" {
		page, err = conv.StringToInt(c.Query("page"))
		if err != nil {
			code := "[HANDLER] GetContentWithQuery = 1"
			log.Errorw(code, err)
			errorResp.Status = false
			errorResp.Message = "Invalid page number"

			return c.Status(fiber.StatusBadRequest).JSON(errorResp)
		}
	}

	limit := 6
	if c.Query("limit") != "" {
		limit, err = conv.StringToInt(c.Query("limit"))
		if err != nil {
			code := "[HANDLER] GetContentWithQuery = 2"
			log.Errorw(code, err)
			errorResp.Status = false
			errorResp.Message = "Invalid limit number"

			return c.Status(fiber.StatusBadRequest).JSON(errorResp)
		}
	}

	orderBy := "created_at"
	if c.Query("orderBy") != "" {
		orderBy = c.Query("orderBy")
	}

	orderType := "desc"
	if c.Query("orderType") != "" {
		orderType = c.Query("orderType")
	}

	search := c.Query("search")
	if c.Query("search") != "" {
		search = c.Query("search")
	}

	categoryID := 0
	if c.Query("categoryID") != "" {
		categoryID, err = conv.StringToInt(c.Query("categoryID"))
		if err != nil {
			code := "[HANDLER] GetContentWithQuery = 3"
			log.Errorw(code, err)
			errorResp.Status = false
			errorResp.Message = "Invalid categoryID number"

			return c.Status(fiber.StatusBadRequest).JSON(errorResp)
		}
	}

	reqEntity := entity.QueryString{
		Limit:      limit,
		Page:       page,
		OrderBy:    orderBy,
		OrderType:  orderType,
		Search:     search,
		Status:     "PUBLISH",
		CategoryID: int64(categoryID),
	}

	results, totalData, totalPages, err := ch.contentService.GetContents(c.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] GetContentWithQuery = 4"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Success"

	respContents := []response.ContentResponse{}

	for _, content := range results {
		respContents = append(respContents, response.ContentResponse{
			ID:           content.ID,
			Title:        content.Title,
			Excerpt:      content.Excerpt,
			Description:  content.Description,
			Image:        content.Image,
			Tags:         content.Tags,
			Status:       content.Status,
			CategoryID:   content.CategoryID,
			CreatedByID:  content.CreatedByID,
			CreatedAt:    content.CreatedAt.Format(time.RFC3339),
			CategoryName: content.Category.Title,
			Author:       content.User.Name,
		})
	}

	defaultSuccessResponse.Data = respContents
	defaultSuccessResponse.Pagination = &response.PaginationResponse{
		TotalRecords: int(totalData),
		Page:         page,
		PerPage:      limit,
		TotalPages:   int(totalPages),
	}

	return c.JSON(defaultSuccessResponse)
}

// CreateContent implements ContentHandler.
func (ch *contentHandler) CreateContent(c *fiber.Ctx) error {
	claims := c.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 {
		code := "[HANDLER] CreateContent = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	userID := claims.UserID
	var req request.ContentRequest
	if err = c.BodyParser(&req); err != nil {
		code := "[HANDLER] CreateContent = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Invalid request body"

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	if err = validatorLib.ValidateStruct(req); err != nil {
		code := "[HANDLER] CreateContent = 3"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	tags := strings.Split(req.Tags, ",")
	reqEntity := entity.ContentEntity{
		Title:       req.Title,
		Excerpt:     req.Excerpt,
		Description: req.Description,
		Image:       req.Image,
		Tags:        tags,
		Status:      req.Status,
		CategoryID:  req.CategoryID,
		CreatedByID: int64(userID),
	}

	err = ch.contentService.CreateContent(c.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] CreateContent = 4"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Content created successfully"
	defaultSuccessResponse.Data = nil

	return c.Status(fiber.StatusCreated).JSON(defaultSuccessResponse)
}

// DeleteContent implements ContentHandler.
func (ch *contentHandler) DeleteContent(c *fiber.Ctx) error {
	claims := c.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 {
		code := "[HANDLER] DeleteContent = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	idParam := c.Params("contentID")
	id, err := conv.StringToInt64(idParam)
	if err != nil {
		code := "[HANDLER] DeleteContent = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	err = ch.contentService.DeleteContent(c.Context(), id)
	if err != nil {
		code := "[HANDLER] DeleteContent = 3"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Success"
	defaultSuccessResponse.Data = nil

	return c.Status(fiber.StatusOK).JSON(defaultSuccessResponse)
}

// EditContentByID implements ContentHandler.
func (ch *contentHandler) EditContentByID(c *fiber.Ctx) error {
	claims := c.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 {
		code := "[HANDLER] EditContentByID = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	userID := claims.UserID
	var req request.ContentRequest
	if err = c.BodyParser(&req); err != nil {
		code := "[HANDLER] EditContentByID = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Invalid request body"

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	if err = validatorLib.ValidateStruct(req); err != nil {
		code := "[HANDLER] EditContentByID = 3"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	idParam := c.Params("contentID")
	id, err := conv.StringToInt64(idParam)
	if err != nil {
		code := "[HANDLER] EditContentByID = 4"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	tags := strings.Split(req.Tags, ",")
	reqEntity := entity.ContentEntity{
		ID:          id,
		Title:       req.Title,
		Excerpt:     req.Excerpt,
		Description: req.Description,
		Image:       req.Image,
		Tags:        tags,
		Status:      req.Status,
		CategoryID:  req.CategoryID,
		CreatedByID: int64(userID),
	}

	err = ch.contentService.EditContentByID(c.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] EditContentByID = 5"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Content updated successfully"
	defaultSuccessResponse.Data = nil

	return c.JSON(defaultSuccessResponse)
}

// GetContentByID implements ContentHandler.
func (ch *contentHandler) GetContentByID(c *fiber.Ctx) error {
	claims := c.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 {
		code := "[HANDLER] GetContentByID = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	idParam := c.Params("contentID")
	id, err := conv.StringToInt64(idParam)
	if err != nil {
		code := "[HANDLER] GetContentByID = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	result, err := ch.contentService.GetContentByID(c.Context(), id)
	if err != nil {
		code := "[HANDLER] GetContentByID = 3"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Success"

	respContent := response.ContentResponse{
		ID:           result.ID,
		Title:        result.Title,
		Excerpt:      result.Excerpt,
		Description:  result.Description,
		Image:        result.Image,
		Tags:         result.Tags,
		Status:       result.Status,
		CategoryID:   result.CategoryID,
		CreatedByID:  result.CreatedByID,
		CreatedAt:    result.CreatedAt.Format(time.RFC3339),
		CategoryName: result.Category.Title,
		Author:       result.User.Name,
	}

	defaultSuccessResponse.Data = respContent

	return c.Status(fiber.StatusOK).JSON(defaultSuccessResponse)
}

// GetContents implements ContentHandler.
func (ch *contentHandler) GetContents(c *fiber.Ctx) error {
	claims := c.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 {
		code := "[HANDLER] GetContents = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}
	page := 1
	if c.Query("page") != "" {
		page, err = conv.StringToInt(c.Query("page"))
		if err != nil {
			code := "[HANDLER] GetContents = 2"
			log.Errorw(code, err)
			errorResp.Status = false
			errorResp.Message = "Invalid page number"

			return c.Status(fiber.StatusBadRequest).JSON(errorResp)
		}
	}

	limit := 10
	if c.Query("limit") != "" {
		limit, err = conv.StringToInt(c.Query("limit"))
		if err != nil {
			code := "[HANDLER] GetContents = 3"
			log.Errorw(code, err)
			errorResp.Status = false
			errorResp.Message = "Invalid limit number"

			return c.Status(fiber.StatusBadRequest).JSON(errorResp)
		}
	}

	orderBy := "created_at"
	if c.Query("orderBy") != "" {
		orderBy = c.Query("orderBy")
	}

	orderType := "desc"
	if c.Query("orderType") != "" {
		orderType = c.Query("orderType")
	}

	search := c.Query("search")
	if c.Query("search") != "" {
		search = c.Query("search")
	}

	categoryID := 0
	if c.Query("categoryID") != "" {
		categoryID, err = conv.StringToInt(c.Query("categoryID"))
		if err != nil {
			code := "[HANDLER] GetContents = 4"
			log.Errorw(code, err)
			errorResp.Status = false
			errorResp.Message = "Invalid categoryID number"

			return c.Status(fiber.StatusBadRequest).JSON(errorResp)
		}
	}

	reqEntity := entity.QueryString{
		Limit:      limit,
		Page:       page,
		OrderBy:    orderBy,
		OrderType:  orderType,
		Search:     search,
		CategoryID: int64(categoryID),
	}

	results, totalData, totalPages, err := ch.contentService.GetContents(c.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] GetContents = 5"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Success"

	respContents := []response.ContentResponse{}

	for _, content := range results {
		respContents = append(respContents, response.ContentResponse{
			ID:           content.ID,
			Title:        content.Title,
			Excerpt:      content.Excerpt,
			Description:  content.Description,
			Image:        content.Image,
			Tags:         content.Tags,
			Status:       content.Status,
			CategoryID:   content.CategoryID,
			CreatedByID:  content.CreatedByID,
			CreatedAt:    content.CreatedAt.Format(time.RFC3339),
			CategoryName: content.Category.Title,
			Author:       content.User.Name,
		})
	}

	defaultSuccessResponse.Data = respContents

	defaultSuccessResponse.Pagination = &response.PaginationResponse{
		TotalRecords: int(totalData),
		Page:         1,
		PerPage:      len(respContents),
		TotalPages:   int(totalPages),
	}

	return c.JSON(defaultSuccessResponse)
}

// UploadImageR2 implements ContentHandler.
func (ch *contentHandler) UploadImageR2(c *fiber.Ctx) error {
	claims := c.Locals("user").(*entity.JwtData)
	if claims.UserID == 0 {
		code := "[HANDLER] UploadImageR2 = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	var req request.FileUploadRequest
	file, err := c.FormFile("image")
	if err != nil {
		code := "[HANDLER] UploadImageR2 = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	if err = c.SaveFile(file, fmt.Sprintf("./temp/content/%s", file.Filename)); err != nil {
		code := "[HANDLER] UploadImageR2 = 3"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	req.Image = fmt.Sprintf("./temp/content/%s", file.Filename)
	reqEntity := entity.FileUploadEntity{
		Name: fmt.Sprintf("%d-%d", int64(claims.UserID), time.Now().UnixNano()),
		Path: req.Image,
	}

	imageUrl, err := ch.contentService.UploadImageR2(c.Context(), reqEntity)
	if err != nil {
		code := "[HANDLER] UploadImageR2 = 4"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	if req.Image != "" {
		err = os.Remove(req.Image)
		if err != nil {
			code := "[HANDLER] UploadImageR2 = 5"
			log.Errorw(code, err)
			errorResp.Status = false
			errorResp.Message = err.Error()

			return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
		}
	}

	urlImageResp := map[string]interface{}{
		"urlImage": imageUrl,
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Success"
	defaultSuccessResponse.Data = urlImageResp

	return c.Status(fiber.StatusCreated).JSON(defaultSuccessResponse)
}

func NewContentHandler(contentService service.ContentService) ContentHandler {
	return &contentHandler{contentService: contentService}
}
