package Helpers

import (
	"math"

	"github.com/gofiber/fiber/v2"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PaginationParams struct {
	Page     int
	PageSize int
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PageSize
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func NewPaginationMeta(params PaginationParams, totalItems int64) PaginationMeta {
	totalPages := int(math.Ceil(float64(totalItems) / float64(params.PageSize)))
	return PaginationMeta{
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

func CalculateTotalPages(pageSize int, totalItems int64) int {
	if pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(totalItems) / float64(pageSize)))
}

func ParsePagination(c *fiber.Ctx) PaginationParams {
	page := c.QueryInt("page", DefaultPage)
	pageSize := c.QueryInt("page_size", DefaultPageSize)

	if page < 1 {
		page = DefaultPage
	}

	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}
