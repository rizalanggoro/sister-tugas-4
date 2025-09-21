package models

import (
	"gorm.io/gorm"
)

type Message struct {
	gorm.Model

	Name    string `json:"name"`
	Message string `json:"message"`
}
