package database

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/marketpal/marketpal/internal/config"
	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect 建立 PostgreSQL 连接并完成 AutoMigrate（禁止在业务代码中散落 sql.Open）。
func Connect(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	log.Info("database connected", "host", cfg.DBHost, "name", cfg.DBName)

	models := []interface{}{
		&model.User{},
		&model.Product{},
		&model.Favorite{},
		&model.Address{},
		&model.CartItem{},
		&model.Order{},
		&model.Message{},
		&model.Review{},
		&model.AuditLog{},
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if err := db.AutoMigrate(models...); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}
	log.Info("database migrated", "models", len(models))

	// 创建演示管理员账号：admin/admin123（仅当不存在时）。
	var adminCount int64
	db.Model(&model.User{}).Where("username = ?", "admin").Count(&adminCount)
	if adminCount == 0 {
		hash, _ := hashPassword("admin123")
		admin := model.User{Username: "admin", PasswordHash: hash, Nickname: "管理员", Role: "admin", CreditScore: 100, Status: "active"}
		if err := db.Create(&admin).Error; err != nil {
			return nil, fmt.Errorf("seed admin: %w", err)
		}
		log.Info("admin user seeded", "username", "admin")
	}
	return db, nil
}
