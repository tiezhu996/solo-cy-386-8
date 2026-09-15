package repository

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/marketpal/marketpal/internal/model"
	"gorm.io/gorm"
)

// newTestDB 创建内存 SQLite 数据库并迁移模型。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	models := []interface{}{
		&model.User{}, &model.Product{}, &model.Favorite{}, &model.Address{},
		&model.CartItem{}, &model.Order{}, &model.Message{}, &model.Review{}, &model.AuditLog{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func TestUserRepository(t *testing.T) {
	db := newTestDB(t)
	repo := NewUserRepository(db)

	cases := []struct {
		name     string
		username string
		nickname string
	}{
		{"first", "alice", "Alice"},
		{"second", "bob", "Bob"},
	}
	for _, tc := range cases {
		t.Run("create_"+tc.name, func(t *testing.T) {
			u := &model.User{Username: tc.username, PasswordHash: "hash", Nickname: tc.nickname, Role: "user"}
			if err := repo.Create(u); err != nil {
				t.Fatalf("Create() error: %v", err)
			}
			if u.ID == 0 {
				t.Fatal("id should be assigned")
			}
		})
	}

	t.Run("get_by_username", func(t *testing.T) {
		u, err := repo.GetByUsername("alice")
		if err != nil {
			t.Fatalf("GetByUsername() error: %v", err)
		}
		if u.Nickname != "Alice" {
			t.Fatalf("expected Alice, got %s", u.Nickname)
		}
	})

	t.Run("get_by_username_missing", func(t *testing.T) {
		if _, err := repo.GetByUsername("nobody"); err != ErrNotFound {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("get_by_id", func(t *testing.T) {
		u, err := repo.GetByID(1)
		if err != nil {
			t.Fatalf("GetByID() error: %v", err)
		}
		if u.Username != "alice" {
			t.Fatalf("expected alice, got %s", u.Username)
		}
	})

	t.Run("update_credit", func(t *testing.T) {
		if err := repo.UpdateCredit(nil, 1, 5); err != nil {
			t.Fatalf("UpdateCredit() error: %v", err)
		}
		u, _ := repo.GetByID(1)
		if u.CreditScore != 105 {
			t.Fatalf("expected credit 105, got %d", u.CreditScore)
		}
	})

	t.Run("list", func(t *testing.T) {
		users, total, err := repo.List(1, 10)
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if total != 2 || len(users) != 2 {
			t.Fatalf("expected 2 users, got total=%d len=%d", total, len(users))
		}
	})
}

func TestProductRepository(t *testing.T) {
	db := newTestDB(t)
	repo := NewProductRepository(db)

	t.Run("create_and_get", func(t *testing.T) {
		p := &model.Product{SellerID: 1, Title: "iPhone 13", Description: "九成新", OriginalPrice: 5999, Price: 3999, Condition: "almost_new", Category: "digital", Status: "on_sale"}
		if err := repo.Create(p); err != nil {
			t.Fatalf("Create() error: %v", err)
		}
		got, err := repo.GetByID(p.ID)
		if err != nil {
			t.Fatalf("GetByID() error: %v", err)
		}
		if got.Title != "iPhone 13" {
			t.Fatalf("expected title, got %s", got.Title)
		}
	})

	t.Run("list_filter_category", func(t *testing.T) {
		_ = repo.Create(&model.Product{SellerID: 1, Title: "书", Description: "二手书", OriginalPrice: 50, Price: 20, Condition: "lightly_used", Category: "books", Status: "on_sale"})
		list, total, err := repo.List(map[string]interface{}{"category": "books", "status": "on_sale"}, "", 1, 10)
		if err != nil {
			t.Fatalf("List() error: %v", err)
		}
		if total != 1 || len(list) != 1 {
			t.Fatalf("expected 1 book, got total=%d len=%d", total, len(list))
		}
	})

	t.Run("incr_view_count", func(t *testing.T) {
		p := &model.Product{SellerID: 1, Title: "x", Description: "y", OriginalPrice: 1, Price: 1, Condition: "brand_new", Category: "other", Status: "on_sale"}
		_ = repo.Create(p)
		if err := repo.IncrViewCount(p.ID); err != nil {
			t.Fatalf("IncrViewCount() error: %v", err)
		}
		got, _ := repo.GetByID(p.ID)
		if got.ViewCount != 1 {
			t.Fatalf("expected view count 1, got %d", got.ViewCount)
		}
	})
}

func TestOrderRepository(t *testing.T) {
	db := newTestDB(t)
	repo := NewOrderRepository(db)

	t.Run("create_and_list_by_buyer", func(t *testing.T) {
		o := &model.Order{OrderNo: "20260101120000000001", BuyerID: 1, SellerID: 2, ProductID: 1, AddressID: 1, Quantity: 1, TotalPrice: 100, Status: "pending_payment"}
		if err := repo.Create(o); err != nil {
			t.Fatalf("Create() error: %v", err)
		}
		list, total, err := repo.ListByBuyer(1, "", 1, 10)
		if err != nil {
			t.Fatalf("ListByBuyer() error: %v", err)
		}
		if total != 1 || len(list) != 1 {
			t.Fatalf("expected 1 order, got total=%d len=%d", total, len(list))
		}
		if list[0].OrderNo != o.OrderNo {
			t.Fatalf("expected order no %s, got %s", o.OrderNo, list[0].OrderNo)
		}
	})

	t.Run("list_by_buyer_status_filter", func(t *testing.T) {
		o := &model.Order{OrderNo: "20260101120000000002", BuyerID: 1, SellerID: 2, ProductID: 2, AddressID: 1, Quantity: 1, TotalPrice: 50, Status: "shipped"}
		_ = repo.Create(o)
		_, total, _ := repo.ListByBuyer(1, "shipped", 1, 10)
		if total != 1 {
			t.Fatalf("expected 1 shipped order, got %d", total)
		}
	})
}
