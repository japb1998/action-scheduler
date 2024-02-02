package types

type PaginationOps struct {
	Page  int `json:"page" form:"page" binding:"omitempty,min=0"`
	Limit int `json:"limit" form:"limit" binding:"omitempty,min=0"`
}

func PaginationDefaults() *PaginationOps {
	return &PaginationOps{
		Page:  0,
		Limit: 10,
	}
}

type PaginationItem interface {
	Schedule
}

type PaginatedResult[T PaginationItem] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}
