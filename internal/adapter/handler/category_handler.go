package handler

import (
	"bwanews/internal/adapter/handler/request"
	"bwanews/internal/adapter/handler/response"
	"bwanews/internal/core/domain/entity"
	"bwanews/internal/core/service"
	"bwanews/lib/conv"
	validatorLib "bwanews/lib/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

var defaultSuccessResponse response.DefaultSuccessResponse

type CategoryHandler interface {
	GetCategories(c *fiber.Ctx) error
	GetCategoryByID(c *fiber.Ctx) error
	CreateCategory(c *fiber.Ctx) error
	EditCategoryByID(c *fiber.Ctx) error
	DeleteCategory(c *fiber.Ctx) error

	GetCategoryFE(c *fiber.Ctx) error
}

type categoryHandler struct {
	categoryService service.CategoryService
}

// GetCategoryFE implements CategoryHandler.
func (ch *categoryHandler) GetCategoryFE(c *fiber.Ctx) error {
	results, err := ch.categoryService.GetCategories(c.Context())
	if err != nil {
		code = "[HANDLER] GetCategoryFE = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	categoryResponses := []response.SuccessCategoryResponse{}
	for _, category := range results {
		categoryResponses = append(categoryResponses, response.SuccessCategoryResponse{
			ID:            category.ID,
			Title:         category.Title,
			Slug:          category.Slug,
			CreatedByName: category.User.Name,
		})
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Successfully retrieved categories"
	defaultSuccessResponse.Data = categoryResponses

	return c.JSON(defaultSuccessResponse)
}

// CreateCategory implements CategoryHandler.
func (ch *categoryHandler) CreateCategory(c *fiber.Ctx) error {
	var req request.CategoryRequest
	claims := c.Locals("user").(*entity.JwtData)
	userId := claims.UserID
	if userId == 0 {
		code = "[HANDLER] CreateCategory = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	if err = c.BodyParser(&req); err != nil {
		code = "[HANDLER] CreateCategory = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Invalid request body"

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	if err = validatorLib.ValidateStruct(req); err != nil {
		code = "[HANDLER] CreateCategory = 3"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	reqEntity := entity.CategoryEntity{
		Title: req.Title,
		User: entity.UserEntity{
			ID: int64(userId),
		},
	}

	err = ch.categoryService.CreateCategory(c.Context(), reqEntity)
	if err != nil {
		code = "[HANDLER] CreateCategory = 4"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	defaultSuccessResponse.Data = nil
	defaultSuccessResponse.Pagination = nil
	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Category created successfully"

	return c.Status(fiber.StatusCreated).JSON(defaultSuccessResponse)
}

// DeleteCategory implements CategoryHandler.
func (ch *categoryHandler) DeleteCategory(c *fiber.Ctx) error {
	claims := c.Locals("user").(*entity.JwtData)
	userId := claims.UserID
	if userId == 0 {
		code = "[HANDLER] DeleteCategory = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	paramsId := c.Params("categoryID")
	id, err := conv.StringToInt64(paramsId)
	if err != nil {
		code = "[HANDLER] DeleteCategory = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	err = ch.categoryService.DeleteCategory(c.Context(), id)
	if err != nil {
		code = "[HANDLER] DeleteCategory = 3"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	defaultSuccessResponse.Data = nil
	defaultSuccessResponse.Pagination = nil
	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Category deleted successfully"

	return c.JSON(defaultSuccessResponse)
}

// EditCategoryByID implements CategoryHandler.
func (ch *categoryHandler) EditCategoryByID(c *fiber.Ctx) error {
	var req request.CategoryRequest
	claims := c.Locals("user").(*entity.JwtData)
	userId := claims.UserID
	if userId == 0 {
		code = "[HANDLER] EditCategoryByID = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	if err = c.BodyParser(&req); err != nil {
		code = "[HANDLER] EditCategoryByID = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Invalid request body"

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	if err = validatorLib.ValidateStruct(req); err != nil {
		code = "[HANDLER] EditCategoryByID = 4"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	idParam := c.Params("categoryID")
	id, err := conv.StringToInt64(idParam)
	if err != nil {
		code = "[HANDLER] EditCategoryByID = 3"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	reqEntity := entity.CategoryEntity{
		ID:    id,
		Title: req.Title,
		User: entity.UserEntity{
			ID: int64(userId),
		},
	}

	err = ch.categoryService.EditCategoryByID(c.Context(), reqEntity)
	if err != nil {
		code = "[HANDLER] EditCategoryByID = 5"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	defaultSuccessResponse.Data = nil
	defaultSuccessResponse.Pagination = nil
	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Category edited successfully"

	return c.JSON(defaultSuccessResponse)
}

// GetCategories implements CategoryHandler.
func (ch *categoryHandler) GetCategories(c *fiber.Ctx) error {
	claims := c.Locals("user").(*entity.JwtData)
	userId := claims.UserID
	if userId == 0 {
		code = "[HANDLER] GetCategories = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	results, err := ch.categoryService.GetCategories(c.Context())
	if err != nil {
		code = "[HANDLER] GetCategories = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	categoryResponses := []response.SuccessCategoryResponse{}
	for _, category := range results {
		categoryResponses = append(categoryResponses, response.SuccessCategoryResponse{
			ID:            category.ID,
			Title:         category.Title,
			Slug:          category.Slug,
			CreatedByName: category.User.Name,
		})
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Meta.Message = "Successfully retrieved categories"
	defaultSuccessResponse.Data = categoryResponses

	return c.JSON(defaultSuccessResponse)
}

// GetCategoryByID implements CategoryHandler.
func (ch *categoryHandler) GetCategoryByID(c *fiber.Ctx) error {
	cliams := c.Locals("user").(*entity.JwtData)
	userId := cliams.UserID
	if userId == 0 {
		code = "[HANDLER] GetCategoryByID = 1"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = "Unauthorized"

		return c.Status(fiber.StatusUnauthorized).JSON(errorResp)
	}

	idParam := c.Params("categoryID")
	id, err := conv.StringToInt64(idParam)
	if err != nil {
		code = "[HANDLER] GetCategoryByID = 2"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusBadRequest).JSON(errorResp)
	}

	result, err := ch.categoryService.GetCategoryByID(c.Context(), id)
	if err != nil {
		code = "[HANDLER] GetCategoryByID = 3"
		log.Errorw(code, err)
		errorResp.Status = false
		errorResp.Message = err.Error()

		return c.Status(fiber.StatusInternalServerError).JSON(errorResp)
	}

	categoryResponse := response.SuccessCategoryResponse{
		ID:            result.ID,
		Title:         result.Title,
		Slug:          result.Slug,
		CreatedByName: result.User.Name,
	}

	defaultSuccessResponse.Meta.Status = true
	defaultSuccessResponse.Pagination = nil
	defaultSuccessResponse.Meta.Message = "Successfully retrieved category"
	defaultSuccessResponse.Data = categoryResponse

	return c.JSON(defaultSuccessResponse)
}

func NewCategoryHandler(categoryService service.CategoryService) CategoryHandler {
	return &categoryHandler{
		categoryService: categoryService,
	}
}
