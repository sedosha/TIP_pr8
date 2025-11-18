package notes

import (
	"context"
	"log"	
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Stats struct {
	Count     int64   `bson:"count" json:"count"`
	AvgLength float64 `bson:"avgLength" json:"avgLength"`
}

var ErrNotFound = errors.New("note not found")

type Repo struct {
	col *mongo.Collection
}

func NewRepo(db *mongo.Database) (*Repo, error) {
	col := db.Collection("notes")
	
	// Уникальный индекс на title
	_, err := col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "title", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return nil, err
	}
	
	// Текстовый индекс
	_, err = col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "content", Value: "text"},
		},
	})
	if err != nil {
		return nil, err
	}
	
	// TTL индекс - ВАЖНО: используем правильные настройки
	_, err = col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "expiresAt", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	if err != nil {
		log.Printf("Warning: TTL index creation failed: %v", err)
		// Не прерываем выполнение, продолжаем без TTL
	}
	
	return &Repo{col: col}, nil
}

func (r *Repo) Create(ctx context.Context, title, content string, ttlSeconds ...int64) (Note, error) {
	now := time.Now()
	n := Note{
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Если передан TTL, устанавливаем expiresAt
	if len(ttlSeconds) > 0 && ttlSeconds[0] > 0 {
		expiresAt := now.Add(time.Duration(ttlSeconds[0]) * time.Second)
		n.ExpiresAt = &expiresAt
	}
	
	res, err := r.col.InsertOne(ctx, n)
	if err != nil {
		return Note{}, err
	}
	n.ID = res.InsertedID.(primitive.ObjectID)
	return n, nil
}

func (r *Repo) ByID(ctx context.Context, idHex string) (Note, error) {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil { return Note{}, ErrNotFound }
	var n Note
	if err := r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&n); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) { return Note{}, ErrNotFound }
		return Note{}, err
	}
	return n, nil
}

func (r *Repo) List(ctx context.Context, q string, limit, skip int64) ([]Note, error) {
	filter := bson.M{}
	
	if q != "" {
		// Пробуем текстовый поиск (работает по отдельным словам)
		filter["$text"] = bson.M{"$search": q}
	}
	
	opts := options.Find().
		SetLimit(limit).
		SetSkip(skip).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})
	
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	
	var out []Note
	for cur.Next(ctx) {
		var n Note
		if err := cur.Decode(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, cur.Err()
}

func (r *Repo) Update(ctx context.Context, idHex string, title, content *string) (Note, error) {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil { return Note{}, ErrNotFound }

	set := bson.M{"updatedAt": time.Now()}
	if title != nil   { set["title"] = *title }
	if content != nil { set["content"] = *content }

	after := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated Note
	if err := r.col.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": set}, after).Decode(&updated); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) { return Note{}, ErrNotFound }
		return Note{}, err
	}
	return updated, nil
}

func (r *Repo) Delete(ctx context.Context, idHex string) error {
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil { return ErrNotFound }
	res, err := r.col.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil { return err }
	if res.DeletedCount == 0 { return ErrNotFound }
	return nil
}

func (r *Repo) CreateWithTTL(ctx context.Context, title, content string, ttlSeconds int64) (Note, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(ttlSeconds) * time.Second)
	
	n := Note{
		Title:     title,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: &expiresAt,
	}
	
	res, err := r.col.InsertOne(ctx, n)
	if err != nil {
		return Note{}, err
	}
	
	n.ID = res.InsertedID.(primitive.ObjectID)
	return n, nil
}

func (r *Repo) ListCursor(ctx context.Context, q string, after string, limit int64) ([]Note, error) {
	filter := bson.M{}
	
	// Курсорная пагинация
	if after != "" {
		oid, err := primitive.ObjectIDFromHex(after)
		if err == nil {
			filter["_id"] = bson.M{"$lt": oid}
		}
	}
	
	// Текстовый поиск
	if q != "" {
		filter["$text"] = bson.M{"$search": q}
	}
	
	if limit <= 0 || limit > 200 {
		limit = 20
	}
	
	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: -1}}).
		SetLimit(limit)
	
	cur, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	
	var out []Note
	for cur.Next(ctx) {
		var n Note
		if err := cur.Decode(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, cur.Err()
}

func (r *Repo) GetStats(ctx context.Context) (Stats, error) {
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id": nil,
				"count": bson.M{"$sum": 1},
				"avgLength": bson.M{
					"$avg": bson.M{"$strLenCP": "$content"},
				},
			},
		},
	}
	
	cur, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return Stats{}, err
	}
	defer cur.Close(ctx)
	
	var results []Stats
	if err := cur.All(ctx, &results); err != nil {
		return Stats{}, err
	}
	
	if len(results) == 0 {
		return Stats{}, nil
	}
	
	return results[0], nil
}
