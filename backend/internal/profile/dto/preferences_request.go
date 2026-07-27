package dto

type UpdatePreferencesRequest struct {
	Timezone string `json:"timezone" binding:"required,max=64"`
}

type PreferencesResponse struct {
	Timezone string `json:"timezone"`
}
