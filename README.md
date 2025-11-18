## Практическое занятие №8 Работа с MongoDB: подключение, создание коллекции, CRUD-операции
## Седова Мария Александровна, ЭФМО-01-25

### Цели:
-	Понять базовые принципы документной БД MongoDB (документ, коллекция, BSON, _id:ObjectID).
-	Научиться подключаться к MongoDB из Go с использованием официального драйвера.
-	Создать коллекцию, индексы и реализовать CRUD для одной сущности (например, notes).
-	Отработать фильтрацию, пагинацию, обновления (в т.ч. частичные), удаление и обработку ошибок.

### Требования:
- Go ≥ 1.21
- Docker (у меня version 28.2.2) & Docker Compose (у меня version v2.40.3)
- MongoDB (через Docker)
- Git (у меня version 2.43.0)
- Systemd  (255)
- Postman (для тестирования API)
- curl 

### Команды запуска (как запустить Mongo и API):
```
cd ~/pz8-mongo
docker-compose up -d
docker exec -it mongo-dev mongosh -u root -p secret --authenticationDatabase admin
exit
```
```
go run ./cmd/api
```

### Скриншоты curl-проверок 

```base_url = http://37.230.117.32:8083```

- Health Check
- <img width="1280" height="396" alt="image" src="https://github.com/user-attachments/assets/d9cb6579-113a-46d5-bb15-4c3db9b0f557" />

- Создать заметку
- <img width="976" height="605" alt="image" src="https://github.com/user-attachments/assets/a1daef3d-2c52-4ab0-abbc-6995cfb1db56" />

- Получить список
- <img width="1280" height="610" alt="image" src="https://github.com/user-attachments/assets/be90fcb2-d7c5-4155-aa93-dbe6f1664a1e" />

- Получить заметку по ID
- <img width="1280" height="488" alt="image" src="https://github.com/user-attachments/assets/25e215c3-fb4e-48f1-a8dc-be142711250b" />

- Обновить заметку 
- <img width="1280" height="539" alt="image" src="https://github.com/user-attachments/assets/6535cdd0-1bfa-4806-88eb-7543064af572" />

- Удалить заметку
- <img width="1280" height="448" alt="image" src="https://github.com/user-attachments/assets/e442ef17-9760-457a-bb24-823bcdd39eb6" />

### Тесты go test (internal/notes/repo_test.go)
```
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
```
<img width="734" height="171" alt="image" src="https://github.com/user-attachments/assets/cb98acc8-f03e-42b2-a2ff-876410b68d2a" />

### Выполненная «звёздочка»
- Поиск по слову
- <img width="1280" height="515" alt="image" src="https://github.com/user-attachments/assets/757313eb-49d6-40cb-8f8f-60003c3a76f4" />

- Первая страница (2 заметки)
- <img width="1280" height="648" alt="image" src="https://github.com/user-attachments/assets/b0110441-5bf2-4a69-b7c6-a472236e3876" />

- Вторая страница (курсорная пагинация)
- <img width="1280" height="647" alt="image" src="https://github.com/user-attachments/assets/76e20d46-4399-4020-8f71-766d727da51e" />

- Получение статистики
- <img width="1280" height="389" alt="image" src="https://github.com/user-attachments/assets/a0240052-0bd7-4c82-8d42-49dee028d4a7" />

- Тестирование TTL-индекса
- <img width="1399" height="645" alt="image" src="https://github.com/user-attachments/assets/ba4a8500-4a58-47cb-a92f-35aca24f4124" />
- <img width="1280" height="709" alt="image" src="https://github.com/user-attachments/assets/dcd0176f-c0e5-4ba8-80b2-c271795b6c3a" />
- <img width="1280" height="484" alt="image" src="https://github.com/user-attachments/assets/699a4e77-7d1e-4d82-aaf6-e1365813e129" />


### Ссылки для проверки
-
-
-
-
-

### Контрольные вопросы

### 1. Чем документная модель MongoDB принципиально отличается от реляционной? Когда она удобнее?

Документная модель MongoDB хранит данные в формате документов (JSON-подобные BSON), где вся информация о сущности хранится в одном документе, включая вложенные структуры и массивы. В реляционных базах данные хранятся в связанных таблицах с фиксированной схемой и нормализацией.

**Преимущества документной модели:**
- Гибкая схема — можно добавлять поля без изменения всей базы.
- Данные, связанные логически, хранятся вместе, уменьшая необходимость JOIN.
- Хорошо подходит для быстро меняющихся или иерархических данных (например, контент, профили пользователей, логика документа).
- Более простое масштабирование и высокая производительность для чтения.

**Когда удобнее:**
- Проекты с нефиксированной схемой.
- Приложения с часто меняющейся структурой данных.
- Большие объемы данных с иерархической структурой.
- Быстрая разработка и прототипирование.

В реляционных БД лучше использовать строгую схему с четкой структурой и сложными транзакциями, например, финансовые системы.

***

### 2. Что такое ObjectID и зачем нужен _id? Как корректно парсить/валидировать его в Go?

`_id` — это уникальный идентификатор документа в коллекции MongoDB. По умолчанию это тип `ObjectID` — 12-байтовый BSON тип, который генерируется автоматически и содержит в себе временную метку, уникальный идентификатор машины, идентификатор процесса и счетчик.

- `_id` служит первичным ключом для уникальной идентификации документа.
- Обеспечивает автоматический индекс и делает быстрые запросы по документу.
- Можно заменить `_id` на любой другой уникальный ключ, но обычно используется `ObjectID`.

**Валидация и парсинг в Go:**

Импортируем `go.mongodb.org/mongo-driver/bson/primitive`

```go
idStr := "507f1f77bcf86cd799439011"
if primitive.IsValidObjectID(idStr) {
    oid, err := primitive.ObjectIDFromHex(idStr)
    if err == nil {
        // oid корректен
    }
} else {
    // id невалиден
}
```

Проверка правильности формата hex-строки обязательна, т.к. неправильный формат вызовет ошибку при конвертации.

***

### 3. Какие операции CRUD предоставляет драйвер MongoDB и какие операторы обновления вы знаете?

MongoDB поддерживает стандартные CRUD-операции:

- **Create**: `InsertOne()`, `InsertMany()`
- **Read**: `Find()`, `FindOne()`
- **Update**: `UpdateOne()`, `UpdateMany()`, `ReplaceOne()`
- **Delete**: `DeleteOne()`, `DeleteMany()`

Операторы обновления, применяемые внутри операций update:

- `$set`: установка значений полей.
- `$unset`: удаление поля.
- `$inc`: увеличение числового значения.
- `$push`: добавить элемент в массив.
- `$pull`: удалить элемент из массива.
- `$addToSet`: добавить в массив элемент, если его там нет.
- `$rename`: переименование поля.

Пример обновления:

```go
update := bson.M{
    "$set": bson.M{"title": "New Title"},
    "$inc": bson.M{"views": 1},
}
collection.UpdateOne(ctx, bson.M{"_id": id}, update)
```

Эти операторы позволяют гибко изменять документы без полного их перезаписывания.

***

### 4. Как устроены индексы в MongoDB? Как создать уникальный индекс и чем он грозит при вставке?

Индексы в MongoDB создаются для ускорения запросов по ключам и полям. Они могут быть:

- Однополевыми и составными (композитными).
- Регулярными, уникальными и с TTL.
- Текстовыми для полнотекстового поиска.

**Уникальный индекс** обеспечивает уникальность значений в индексируемом поле(ях):

```go
_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
  Keys: bson.D{{Key: "email", Value: 1}},
  Options: options.Index().SetUnique(true),
})
```

**Риски при вставке с уникальным индексом:**
- Вставка документа с дублирующимся значением вызовет ошибку.
- При создании уникального индекса на коллекции с дубликатами операция завершится с ошибкой.
- Уникальные индексы не могут быть "разреженными" (`sparse: true`) — если поле отсутствует в нескольких документах, это считается дублированием null.

Поэтому перед созданием индексов важно проверить уникальность данных или очистить дубликаты.

***

### 5. Почему важно использовать context.WithTimeout при вызовах к базе? Что произойдет при его срабатывании?

`context.WithTimeout` позволяет задать ограничение по времени для операции с БД, чтобы она автоматически прерывалась при превышении лимита.

**Почему важно:**

- Предотвращает "зависание" запросов к базе.
- Освобождает ресурсы при слишком долгих операциях.
- Повышает устойчивость приложения, позволяя корректно обрабатывать тайм-ауты.
- Позволяет отменять запросы, если клиент прервал сетевое соединение.

**Что происходит при срабатывании:**

- Контекст отменяется, и вызываемый метод MongoDB прерывается с ошибкой.
- Ошибка `context.DeadlineExceeded` возвращается вызывающему коду.
- Можно обработать ситуацию и, например, вернуть пользователю ошибку тайм-аута или повторить попытку.

Пример в Go:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := collection.FindOne(ctx, filter).Err()
if errors.Is(err, context.DeadlineExceeded) {
    // тайм-аут произошел
}
```
Это обеспечивает надежность и стабильность приложений при работе с БД.
