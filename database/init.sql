-- marketpal 本地生活市场平台数据库初始化脚本（PostgreSQL）
-- 应用启动时由 GORM AutoMigrate 自动建表，本脚本用于手工建库与参考。

CREATE DATABASE marketpal_db;
CREATE USER marketpal_user WITH PASSWORD 'marketpal_pwd';
GRANT ALL PRIVILEGES ON DATABASE marketpal_db TO marketpal_user;
