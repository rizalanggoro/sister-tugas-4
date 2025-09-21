package responses

import "sister/internal/models"

type GetAllGlobalMessages struct {
	Messages []models.Message `json:"messages"`
}
