package dto

type UpdateProfileRequest struct {
	FullName  *string `json:"full_name" binding:"omitempty,max=100"`
	AvatarURL *string `json:"avatar_url" binding:"omitempty,url"`
	Timezone  *string `json:"timezone" binding:"omitempty,max=64"`
}
