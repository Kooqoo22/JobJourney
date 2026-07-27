package dto

type ListUsersQuery struct {
	Q      string `form:"q"`
	Status string `form:"status"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}
