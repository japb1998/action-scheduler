package types

type Action struct {
	Id   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
	Arn  string `json:"arn" binding:"required"`
	Role string `json:"role" binding:"required"`
}
