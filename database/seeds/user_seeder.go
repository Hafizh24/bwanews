package seeds

import (
	"bwanews/internal/core/domain/model"
	"bwanews/lib/conv"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func SeedRoles(db *gorm.DB) {
	bytes, err := conv.HashPassword("admin123")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to hash password")
	}

	admin := model.User{
		Name:     "Admin",
		Email:    "admin@mail.com",
		Password: string(bytes),
	}

	if err := db.FirstOrCreate(&admin, model.User{Email: admin.Email}).Error; err != nil {
		log.Fatal().Err(err).Msg("Failed to seed admin user")
	} else {
		log.Info().Msg("Admin user seeded successfully")
	}
}
