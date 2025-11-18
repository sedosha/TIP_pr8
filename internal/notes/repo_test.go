package notes

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"example.com/pz8-mongo/internal/db"
)

func TestCreateAndGet(t *testing.T) {
	ctx := context.Background()
	
	// Создаём уникальное имя БД для изоляции
	dbName := fmt.Sprintf("pz8_test_%d", time.Now().UnixNano())
	
	deps, err := db.ConnectMongo(ctx, os.Getenv("MONGO_URI"), dbName)
	if err != nil {
		t.Fatal("Connect failed:", err)
	}
	
	// Очистка: дропим БД и отключаемся после теста
	t.Cleanup(func() {
		deps.Client.Database(dbName).Drop(ctx)
		deps.Client.Disconnect(ctx)
	})
	
	r, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatal("NewRepo failed:", err)
	}
	
	// Тест Create
	created, err := r.Create(ctx, "T1", "C1")
	if err != nil {
		t.Fatal("Create failed:", err)
	}
	
	// Тест ByID
	got, err := r.ByID(ctx, created.ID.Hex())
	if err != nil {
		t.Fatal("ByID failed:", err)
	}
	if got.Title != "T1" {
		t.Fatalf("want T1 got %s", got.Title)
	}
}
