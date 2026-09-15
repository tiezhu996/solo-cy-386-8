-- 002_product_review.sql 商品平台审核闭环（GORM AutoMigrate 会生成等价结构，本脚本供手工迁移参考）。
-- 1) products 增加审核字段：存量行先以 DEFAULT 'approved' 回填，保证原有已上架商品继续可售。
ALTER TABLE products ADD COLUMN IF NOT EXISTS review_status VARCHAR(32) NOT NULL DEFAULT 'approved';
ALTER TABLE products ADD COLUMN IF NOT EXISTS review_round INTEGER NOT NULL DEFAULT 1;
ALTER TABLE products ADD COLUMN IF NOT EXISTS reject_reason TEXT;
-- 回填后把默认值切回待审核：此后新发布商品必须通过平台审核。
ALTER TABLE products ALTER COLUMN review_status SET DEFAULT 'pending_review';
CREATE INDEX IF NOT EXISTS idx_products_review_status ON products (review_status);

-- 2) product_reviews 审核记录表：每个商品每一轮送审仅一条记录。
CREATE TABLE IF NOT EXISTS product_reviews (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    round INTEGER NOT NULL DEFAULT 1,
    status VARCHAR(32) NOT NULL DEFAULT 'pending_review',
    reason TEXT,
    reviewer_id BIGINT,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uk_product_review_round UNIQUE (product_id, round)
);
CREATE INDEX IF NOT EXISTS idx_product_reviews_status ON product_reviews (status);
CREATE INDEX IF NOT EXISTS idx_product_reviews_product ON product_reviews (product_id);
CREATE INDEX IF NOT EXISTS idx_product_reviews_seller ON product_reviews (seller_id);
