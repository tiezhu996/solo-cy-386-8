package config

import (
	"os"
	"time"
)

// Config 应用配置：全部通过环境变量注入。
type Config struct {
	Env            string
	Port           string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	RedisAddr      string
	RedisPassword  string
	JWTSecret      string
	JWTExpireHours int
	UploadDir      string
	PublicURL      string
}

// Load 从环境变量解析配置（提供合理的本地默认值）。
func Load() *Config {
	return &Config{
		Env:            getEnv("APP_ENV", "development"),
		Port:           getEnv("SERVER_PORT", "8080"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBUser:         getEnv("DB_USER", "marketpal_user"),
		DBPassword:     getEnv("DB_PASSWORD", "marketpal_pwd"),
		DBName:         getEnv("DB_NAME", "marketpal_db"),
		DBSSLMode:      getEnv("DB_SSLMODE", "disable"),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		JWTSecret:      getEnv("JWT_SECRET", "change_me_to_a_long_random_string"),
		JWTExpireHours: getEnvInt("JWT_EXPIRE_HOURS", 72),
		UploadDir:      getEnv("UPLOAD_DIR", "/app/uploads"),
		PublicURL:      getEnv("PUBLIC_URL", "http://localhost:19406/uploads"),
	}
}

// JWTExpireDuration 返回 JWT 过期时长。
func (c *Config) JWTExpireDuration() time.Duration {
	return time.Duration(c.JWTExpireHours) * time.Hour
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return def
		}
		n = n*10 + int(ch-'0')
	}
	return n
}
