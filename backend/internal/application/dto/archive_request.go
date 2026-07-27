package dto

type ArchiveRequest struct {
	IsArchived *bool `json:"is_archived" binding:"required"`
}
