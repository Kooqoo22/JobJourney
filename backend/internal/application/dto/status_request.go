package dto

type ChangeStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=wishlist applied screening interview offer accepted rejected withdrawn ghosted"`
}
