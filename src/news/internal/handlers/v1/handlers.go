package v1

import (
	"context"
	"errors"

	"github.com/go-playground/validator"
	"github.com/gofiber/fiber/v2"
	"github.com/zhavkk/news-service/src/news/internal/dto"
	"github.com/zhavkk/news-service/src/news/internal/repository/postgres"
)

var validate = validator.New()

type NewsService interface {
	CreateNews(ctx context.Context, req dto.CreateNewsRequest) (*dto.NewsResponse, error)
	UpdateNews(ctx context.Context, req dto.UpdateNewsRequest) (*dto.UpdateNewsResponse, error)
	GetNewsByID(ctx context.Context, req dto.GetNewsByIDRequest) (*dto.NewsResponse, error)
	DeleteNews(ctx context.Context, req dto.DeleteNewsRequest) (*dto.DeleteNewsResponse, error)
	ListNews(ctx context.Context, req dto.NewsListRequest) (*dto.NewsListResponse, error)
}

type NewsHandler struct {
	newsService NewsService
}

func NewHandler(newsService NewsService) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
	}
}

func (h *NewsHandler) RegisterRoutes(router fiber.Router) {
	news := router.Group("/news")

	news.Post("/", h.CreateNews)
	news.Get("/:id", h.GetNewsByID)
	news.Put("/:id", h.UpdateNews)
	news.Delete("/:id", h.DeleteNews)
	news.Get("/", h.ListNews)
}

func (h *NewsHandler) CreateNews(c *fiber.Ctx) error {
	ctx := c.Context()
	var req dto.CreateNewsRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Status:  fiber.StatusBadRequest,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	resp, err := h.newsService.CreateNews(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "Failed to create news",
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *NewsHandler) UpdateNews(c *fiber.Ctx) error {
	ctx := c.Context()
	var req dto.UpdateNewsRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
	}

	req.ID = c.Params("id")

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Status:  fiber.StatusBadRequest,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}
	id := c.Params("id")
	req.ID = id

	resp, err := h.newsService.UpdateNews(ctx, req)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Status:  fiber.StatusNotFound,
				Message: "News not found for update",
				Error:   err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "Failed to update news",
			Error:   err.Error(),
		})
	}

	return c.JSON(resp)
}

func (h *NewsHandler) GetNewsByID(c *fiber.Ctx) error {
	ctx := c.Context()

	req := dto.GetNewsByIDRequest{
		ID:              c.Params("id"),
		CheckVisibility: c.QueryBool("check_visibility", true),
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Status:  fiber.StatusBadRequest,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	resp, err := h.newsService.GetNewsByID(ctx, req)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Status:  fiber.StatusNotFound,
				Message: "News not found",
				Error:   err.Error(),
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "Failed to get news",
			Error:   err.Error(),
		})
	}

	return c.JSON(resp)
}

func (h *NewsHandler) DeleteNews(c *fiber.Ctx) error {
	ctx := c.Context()
	id := c.Params("id")

	req := dto.DeleteNewsRequest{ID: id}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Status:  fiber.StatusBadRequest,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	resp, err := h.newsService.DeleteNews(ctx, req)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Status:  fiber.StatusNotFound,
				Message: "News not found for deletion",
				Error:   err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "Failed to delete news",
			Error:   err.Error(),
		})
	}
	return c.JSON(resp)
}
func (h *NewsHandler) ListNews(c *fiber.Ctx) error {
	ctx := c.Context()
	var req dto.NewsListRequest

	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Status:  fiber.StatusBadRequest,
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
	}
	req.CheckVisibility = c.QueryBool("check_visibility", true)
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	if err := validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Status:  fiber.StatusBadRequest,
			Message: "Validation failed",
			Error:   err.Error(),
		})
	}

	resp, err := h.newsService.ListNews(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "Failed to list news",
			Error:   err.Error(),
		})
	}

	return c.JSON(resp)
}
