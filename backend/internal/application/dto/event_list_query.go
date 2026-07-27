package dto

type EventListQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}
