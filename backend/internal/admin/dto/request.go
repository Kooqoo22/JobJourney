package dto

type ListUsersQuery struct {
	Q      string `form:"q"`
	Status string `form:"status"`
	Page   int    `form:"page"`
	Limit  int    `form:"limit"`
}

type UserIDParam struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

type BanUserRequest struct {
	Reason *string `json:"reason" binding:"omitempty,max=500"`
}
