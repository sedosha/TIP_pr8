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
	
	uri := getMongoURI()
	dbName := fmt.Sprintf("pz8_test_%d", time.Now().UnixNano())
	
	deps, err := db.ConnectMongo(ctx, uri, dbName)
	if err != nil {
		t.Fatal("Connect failed:", err)
	}
	
	t.Cleanup(func() {
		deps.Client.Database(dbName).Drop(ctx)
		deps.Client.Disconnect(ctx)
	})
	
	r, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatal("NewRepo failed:", err)
	}
	
	created, err := r.Create(ctx, "T1", "C1")
	if err != nil {
		t.Fatal("Create failed:", err)
	}
	
	got, err := r.ByID(ctx, created.ID.Hex())
	if err != nil {
		t.Fatal("ByID failed:", err)
	}
	if got.Title != "T1" {
		t.Fatalf("want T1 got %s", got.Title)
	}
}

func TestCreateDuplicateTitle(t *testing.T) {
	ctx := context.Background()
	
	uri := getMongoURI()
	dbName := fmt.Sprintf("pz8_test_dup_%d", time.Now().UnixNano())
	
	deps, err := db.ConnectMongo(ctx, uri, dbName)
	if err != nil {
		t.Fatal("Connect failed:", err)
	}
	
	t.Cleanup(func() {
		deps.Client.Database(dbName).Drop(ctx)
		deps.Client.Disconnect(ctx)
	})
	
	r, err := NewRepo(deps.Database)
	if err != nil {
		t.Fatal("NewRepo failed:", err)
	}

	// Создаем первую заметку
	_, err = r.Create(ctx, "Duplicate Title", "Content 1")
	if err != nil {
		t.Fatal("Failed to create first note:", err)
	}

	// Пытаемся создать заметку с тем же заголовком
	_, err = r.Create(ctx, "Duplicate Title", "Content 2")
	if err == nil {
		t.Error("Expected error for duplicate title, but got none")
	}
}

// Вспомогательная функция для получения URI MongoDB
func getMongoURI() string {
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}
	return "mongodb://root:secret@localhost:27017/?authSource=admin"
}
EOF
