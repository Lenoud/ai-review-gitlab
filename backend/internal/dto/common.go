package dto

type PageRequest struct {
	Page int    `form:"page" json:"page"`
	Size int    `form:"size" json:"size"`
	Sort string `form:"sort" json:"sort"`
}

type PageResponse[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"size"`
}
