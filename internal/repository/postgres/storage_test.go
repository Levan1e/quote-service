package postgres

import (
	"context"
	"quote-service/internal/models"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockConn реализует DBConn для мока
type MockConn struct {
	mock.Mock
}

type MockCommandTag struct {
	tag string
}

func (m MockCommandTag) RowsAffected() int64 {
	if strings.HasPrefix(m.tag, "DELETE") {
		num, _ := strconv.ParseInt(strings.TrimPrefix(m.tag, "DELETE "), 10, 64)
		return num
	}
	return 0
}

func (m MockCommandTag) String() string {
	return m.tag
}

func (m MockCommandTag) Insert() bool { return false }
func (m MockCommandTag) Update() bool { return false }
func (m MockCommandTag) Delete() bool { return strings.HasPrefix(m.tag, "DELETE") }
func (m MockCommandTag) Select() bool { return false }
func (m MockCommandTag) Copy() bool   { return false }

func (m *MockConn) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	ret := m.Called(ctx, sql, args)
	return ret.Get(0).(pgx.Row)
}

func (m *MockConn) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	ret := m.Called(ctx, sql, args)
	return ret.Get(0).(pgx.Rows), ret.Error(1)
}

func (m *MockConn) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	argsCalled := m.Called(ctx, sql, args)
	return argsCalled.Get(0).(pgconn.CommandTag), argsCalled.Error(1)
}

// MockRow для мока pgx.Row
type MockRow struct {
	mock.Mock
}

func (m *MockRow) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

// MockRows для мока pgx.Rows
type MockRows struct {
	mock.Mock
}

func (m *MockRows) Close() {
	m.Called()
}

func (m *MockRows) Next() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	args := m.Called(dest...)
	return args.Error(0)
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	args := m.Called()
	return args.Get(0).(pgconn.CommandTag)
}

func (m *MockRows) Conn() *pgx.Conn {
	args := m.Called()
	return args.Get(0).(*pgx.Conn)
}

func (m *MockRows) Err() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	args := m.Called()
	return args.Get(0).([]pgconn.FieldDescription)
}

func (m *MockRows) RawValues() [][]byte {
	args := m.Called()
	return args.Get(0).([][]byte)
}

func (m *MockRows) Values() ([]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]interface{}), args.Error(1)
}

func TestStorage_Create(t *testing.T) {
	mockConn := new(MockConn)
	mockRow := new(MockRow)
	storage := NewStorage(mockConn)

	quote := &models.Quote{
		Author: "Confucius",
		Quote:  "Life is simple",
	}

	t.Run("successful create", func(t *testing.T) {
		// Мок для поиска свободного ID
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			id := args.Get(0).(*int)
			*id = 1
		}).Return(nil).Once()
		mockConn.On("QueryRow", mock.Anything, "\n        WITH RECURSIVE ids AS (\n            SELECT generate_series(1, (SELECT COALESCE(MAX(id), 0) + 1 FROM quotes)) AS id\n        )\n        SELECT MIN(id)\n        FROM ids\n        WHERE NOT EXISTS (SELECT 1 FROM quotes WHERE quotes.id = ids.id)\n    ", []interface{}(nil)).Return(mockRow).Once()

		// Мок для вставки цитаты
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			createdAt := args.Get(0).(*time.Time)
			*createdAt = time.Now()
		}).Return(nil).Once()
		mockConn.On("QueryRow", mock.Anything, "INSERT INTO quotes (id, author, quote) VALUES ($1, $2, $3) RETURNING created_at", []interface{}{1, quote.Author, quote.Quote}).Return(mockRow).Once()

		err := storage.Create(context.Background(), quote)
		assert.NoError(t, err)
		assert.Equal(t, 1, quote.ID)
	})

	t.Run("db error", func(t *testing.T) {
		// Мок для поиска свободного ID
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			id := args.Get(0).(*int)
			*id = 1
		}).Return(nil).Once()
		mockConn.On("QueryRow", mock.Anything, "\n        WITH RECURSIVE ids AS (\n            SELECT generate_series(1, (SELECT COALESCE(MAX(id), 0) + 1 FROM quotes)) AS id\n        )\n        SELECT MIN(id)\n        FROM ids\n        WHERE NOT EXISTS (SELECT 1 FROM quotes WHERE quotes.id = ids.id)\n    ", []interface{}(nil)).Return(mockRow).Once()

		// Мок для ошибки при вставке
		mockRow.On("Scan", mock.Anything).Return(pgx.ErrNoRows).Once()
		mockConn.On("QueryRow", mock.Anything, "INSERT INTO quotes (id, author, quote) VALUES ($1, $2, $3) RETURNING created_at", []interface{}{1, quote.Author, quote.Quote}).Return(mockRow).Once()

		err := storage.Create(context.Background(), quote)
		assert.Error(t, err)
	})
}

func TestStorage_GetAll(t *testing.T) {
	mockConn := new(MockConn)
	mockRows := new(MockRows)
	storage := NewStorage(mockConn)

	quotes := []models.Quote{
		{ID: 1, Author: "Confucius", Quote: "Life is simple", CreatedAt: time.Now()},
	}

	t.Run("successful get all", func(t *testing.T) {
		t.Log("Настройка мока для GetAll")
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			t.Log("Scan вызван")
			id := args.Get(0).(*int)
			author := args.Get(1).(*string)
			quote := args.Get(2).(*string)
			createdAt := args.Get(3).(*time.Time)
			*id = quotes[0].ID
			*author = quotes[0].Author
			*quote = quotes[0].Quote
			*createdAt = quotes[0].CreatedAt
		}).Return(nil).Once()
		mockRows.On("Next").Return(false).Once()
		mockRows.On("Close").Return().Once()
		mockRows.On("Err").Return(nil).Once()
		mockConn.On("Query", mock.Anything, "SELECT id, author, quote, created_at FROM quotes ORDER BY id", []interface{}(nil)).Return(mockRows, nil).Once()

		t.Log("Вызов GetAll")
		result, err := storage.GetAll(context.Background())
		t.Logf("Результат GetAll: %+v", result)
		assert.NoError(t, err)
		assert.Len(t, result, 1, "ожидается 1 цитата, получено %d", len(result))
		if len(result) > 0 {
			assert.Equal(t, quotes[0].Author, result[0].Author)
			assert.Equal(t, quotes[0].Quote, result[0].Quote)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		mockRows.On("Next").Return(false).Once()
		mockRows.On("Close").Return().Once()
		mockRows.On("Err").Return(nil).Once()
		mockConn.On("Query", mock.Anything, "SELECT id, author, quote, created_at FROM quotes ORDER BY id", []interface{}(nil)).Return(mockRows, nil).Once()

		result, err := storage.GetAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, result, 0, "ожидается 0 цитат, получено %d", len(result))
	})
}

func TestStorage_GetRandom(t *testing.T) {
	mockConn := new(MockConn)
	mockRow := new(MockRow)
	storage := NewStorage(mockConn)

	quote := &models.Quote{ID: 1, Author: "Confucius", Quote: "Life is simple", CreatedAt: time.Now()}

	t.Run("successful get random", func(t *testing.T) {
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			id := args.Get(0).(*int)
			author := args.Get(1).(*string)
			quoteText := args.Get(2).(*string)
			createdAt := args.Get(3).(*time.Time)
			*id = quote.ID
			*author = quote.Author
			*quoteText = quote.Quote
			*createdAt = quote.CreatedAt
		}).Return(nil).Once()
		mockConn.On("QueryRow", mock.Anything, "SELECT id, author, quote, created_at FROM quotes ORDER BY RANDOM() LIMIT 1", []interface{}(nil)).Return(mockRow).Once()

		result, err := storage.GetRandom(context.Background())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, quote.Author, result.Author)
		assert.Equal(t, quote.Quote, result.Quote)
	})
}

func TestStorage_GetByAuthor(t *testing.T) {
	mockConn := new(MockConn)
	mockRows := new(MockRows)
	storage := NewStorage(mockConn)

	quotes := []models.Quote{
		{ID: 1, Author: "Confucius", Quote: "Life is simple", CreatedAt: time.Now()},
	}

	t.Run("successful get by author", func(t *testing.T) {
		t.Log("Настройка мока для GetByAuthor")
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			t.Log("Scan вызван")
			id := args.Get(0).(*int)
			author := args.Get(1).(*string)
			quote := args.Get(2).(*string)
			createdAt := args.Get(3).(*time.Time)
			*id = quotes[0].ID
			*author = quotes[0].Author
			*quote = quotes[0].Quote
			*createdAt = quotes[0].CreatedAt
		}).Return(nil).Once()
		mockRows.On("Next").Return(false).Once()
		mockRows.On("Close").Return().Once()
		mockRows.On("Err").Return(nil).Once()
		mockConn.On("Query", mock.Anything, "SELECT id, author, quote, created_at FROM quotes WHERE author = $1 ORDER BY id", []interface{}{"Confucius"}).Return(mockRows, nil).Once()

		t.Log("Вызов GetByAuthor")
		result, err := storage.GetByAuthor(context.Background(), "Confucius")
		t.Logf("Результат GetByAuthor: %+v", result)
		assert.NoError(t, err)
		assert.Len(t, result, 1, "ожидается 1 цитата, получено %d", len(result))
		if len(result) > 0 {
			assert.Equal(t, quotes[0].Author, result[0].Author)
			assert.Equal(t, quotes[0].Quote, result[0].Quote)
		}
	})
}

func TestStorage_Exists(t *testing.T) {
	mockConn := new(MockConn)
	mockRow := new(MockRow)
	storage := NewStorage(mockConn)

	t.Run("success", func(t *testing.T) {
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			exists := args.Get(0).(*bool)
			*exists = true
		}).Return(nil).Once()
		mockConn.On("QueryRow", mock.Anything, "SELECT EXISTS(SELECT 1 FROM quotes WHERE author = $1 AND quote = $2)", []interface{}{"Confucius", "Life is simple"}).Return(mockRow).Once()

		exists, err := storage.Exists(context.Background(), "Confucius", "Life is simple")
		assert.NoError(t, err)
		assert.True(t, exists)
	})
}
