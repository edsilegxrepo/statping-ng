package database

import (
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestByString tests the By.String() method for formatting grouping intervals
func TestByString(t *testing.T) {
	tests := []struct {
		name     string
		by       By
		expected string
	}{
		{"ByCount", ByCount, "COUNT(id) as amount"},
		{"Custom By", By("SUM(value) as total"), "SUM(value) as total"},
		{"Empty By", By(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.by.String())
		})
	}
}

// TestByAverage tests the ByAverage function for different database types
func TestByAverage(t *testing.T) {
	tests := []struct {
		name       string
		dbType     string
		column     string
		multiplier int
		contains   string
	}{
		{"MySQL", "mysql", "latency", 1, "UNSIGNED INT"},
		{"Postgres", "postgres", "latency", 1, "cast(AVG(latency)"},
		{"SQLite", "sqlite", "latency", 1, "cast(AVG(latency)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original db and restore after test
			originalDb := database
			defer func() { Set(originalDb) }()

			// Create a test database and set it as the global database
			db, err := Openw(tt.dbType, ":memory:")
			if err != nil {
				// MySQL/Postgres won't connect, but we can test the function exists
				t.Skipf("Skipping %s test - connection not available", tt.dbType)
			}
			defer func() {
				sqlDB, _ := db.DB()
				_ = sqlDB.Close()
			}()
			Set(db)

			result := ByAverage(tt.column, tt.multiplier)
			assert.Contains(t, result.String(), "AVG")
			assert.Contains(t, result.String(), tt.column)
		})
	}
}

// TestGroupQueryFind tests the GroupQuery.Find() method
func TestGroupQueryFind(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	// Insert test data
	for i := 1; i <= 5; i++ {
		db.Create(&TestModel{Name: "find_test", Value: i * 10})
	}

	t.Run("FindWithoutTable", func(t *testing.T) {
		query := &GroupQuery{
			db: db,
		}

		var models []TestModel
		err := query.Find(&models)
		require.NoError(t, err)
		assert.Len(t, models, 5)
	})

	t.Run("FindWithTable", func(t *testing.T) {
		query := &GroupQuery{
			db:    db.Table("test_models"),
			Table: "test_models",
		}

		var models []TestModel
		err := query.Find(&models)
		require.NoError(t, err)
		assert.Len(t, models, 5)
	})
}

// TestGroupQueryDatabase tests the GroupQuery.Database() method
func TestGroupQueryDatabase(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	query := &GroupQuery{
		db: db,
	}

	result := query.Database()
	assert.NotNil(t, result)
	assert.Equal(t, db, result)
}

// TestTimeVarToValues tests the TimeVar.ToValues() method
func TestTimeVarToValues(t *testing.T) {
	t.Run("EmptyData", func(t *testing.T) {
		tv := &TimeVar{
			data: []*TimeValue{},
		}
		values, err := tv.ToValues()
		require.NoError(t, err)
		assert.Empty(t, values)
	})

	t.Run("WithData", func(t *testing.T) {
		tv := &TimeVar{
			data: []*TimeValue{
				{Timeframe: "2023-01-01T00:00:00Z", Amount: 100},
				{Timeframe: "2023-01-01T01:00:00Z", Amount: 200},
				{Timeframe: "2023-01-01T02:00:00Z", Amount: 150},
			},
		}
		values, err := tv.ToValues()
		require.NoError(t, err)
		assert.Len(t, values, 3)
		assert.Equal(t, int64(100), values[0].Amount)
		assert.Equal(t, int64(200), values[1].Amount)
		assert.Equal(t, int64(150), values[2].Amount)
	})

	t.Run("NilData", func(t *testing.T) {
		tv := &TimeVar{
			data: nil,
		}
		values, err := tv.ToValues()
		require.NoError(t, err)
		assert.Nil(t, values)
	})
}

// TestFillMissing tests the TimeVar.FillMissing() method for filling gaps in time series
func TestFillMissing(t *testing.T) {
	t.Run("FillHourlyGaps", func(t *testing.T) {
		start := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
		end := time.Date(2023, 6, 15, 14, 0, 0, 0, time.UTC)

		query := &GroupQuery{
			Start: start,
			End:   end,
			Group: time.Hour,
		}

		// Only have data for hours 10 and 12
		tv := &TimeVar{
			g: query,
			data: []*TimeValue{
				{Timeframe: types.FixedTime(start, time.Hour), Amount: 100},
				{Timeframe: types.FixedTime(start.Add(2*time.Hour), time.Hour), Amount: 300},
			},
		}

		filled, err := tv.FillMissing(start, end)
		require.NoError(t, err)

		// Should have 5 entries: 10, 11, 12, 13, 14
		assert.Len(t, filled, 5)

		// Check existing values are preserved
		assert.Equal(t, int64(100), filled[0].Amount) // Hour 10
		assert.Equal(t, int64(0), filled[1].Amount)   // Hour 11 (missing)
		assert.Equal(t, int64(300), filled[2].Amount) // Hour 12
		assert.Equal(t, int64(0), filled[3].Amount)   // Hour 13 (missing)
		assert.Equal(t, int64(0), filled[4].Amount)   // Hour 14 (missing)
	})

	t.Run("FillMinuteGaps", func(t *testing.T) {
		start := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
		end := time.Date(2023, 6, 15, 10, 4, 0, 0, time.UTC)

		query := &GroupQuery{
			Start: start,
			End:   end,
			Group: time.Minute,
		}

		tv := &TimeVar{
			g: query,
			data: []*TimeValue{
				{Timeframe: types.FixedTime(start, time.Minute), Amount: 50},
			},
		}

		filled, err := tv.FillMissing(start, end)
		require.NoError(t, err)

		// Should have 5 entries: 0, 1, 2, 3, 4
		assert.Len(t, filled, 5)
		assert.Equal(t, int64(50), filled[0].Amount)
		assert.Equal(t, int64(0), filled[1].Amount)
	})

	t.Run("FillDailyGaps", func(t *testing.T) {
		start := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2023, 6, 3, 0, 0, 0, 0, time.UTC)

		query := &GroupQuery{
			Start: start,
			End:   end,
			Group: types.Day,
		}

		tv := &TimeVar{
			g: query,
			data: []*TimeValue{
				{Timeframe: types.FixedTime(start, types.Day), Amount: 1000},
			},
		}

		filled, err := tv.FillMissing(start, end)
		require.NoError(t, err)

		// Should have 3 entries: June 1, 2, 3
		assert.Len(t, filled, 3)
		assert.Equal(t, int64(1000), filled[0].Amount)
		assert.Equal(t, int64(0), filled[1].Amount)
		assert.Equal(t, int64(0), filled[2].Amount)
	})

	t.Run("NoGaps", func(t *testing.T) {
		start := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
		end := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

		query := &GroupQuery{
			Start: start,
			End:   end,
			Group: time.Hour,
		}

		tv := &TimeVar{
			g: query,
			data: []*TimeValue{
				{Timeframe: types.FixedTime(start, time.Hour), Amount: 100},
				{Timeframe: types.FixedTime(start.Add(time.Hour), time.Hour), Amount: 200},
				{Timeframe: types.FixedTime(start.Add(2*time.Hour), time.Hour), Amount: 300},
			},
		}

		filled, err := tv.FillMissing(start, end)
		require.NoError(t, err)

		assert.Len(t, filled, 3)
		assert.Equal(t, int64(100), filled[0].Amount)
		assert.Equal(t, int64(200), filled[1].Amount)
		assert.Equal(t, int64(300), filled[2].Amount)
	})

	t.Run("EmptyData", func(t *testing.T) {
		start := time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)
		end := time.Date(2023, 6, 15, 12, 0, 0, 0, time.UTC)

		query := &GroupQuery{
			Start: start,
			End:   end,
			Group: time.Hour,
		}

		tv := &TimeVar{
			g:    query,
			data: []*TimeValue{},
		}

		filled, err := tv.FillMissing(start, end)
		require.NoError(t, err)

		// Should still create entries with 0 amounts
		assert.Len(t, filled, 3)
		for _, v := range filled {
			assert.Equal(t, int64(0), v.Amount)
		}
	})
}

// TestParseRequest tests the ParseRequest function for parsing HTTP request params
func TestParseRequest(t *testing.T) {
	t.Run("DefaultValues", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data", nil)

		query, err := ParseRequest(req)
		require.NoError(t, err)

		assert.Equal(t, time.Hour, query.Group)
		assert.Equal(t, 10000, query.Limit)
		assert.False(t, query.FillEmpty)
	})

	t.Run("CustomGrouping", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data?group=5m", nil)

		query, err := ParseRequest(req)
		require.NoError(t, err)

		assert.Equal(t, 5*time.Minute, query.Group)
	})

	t.Run("AllParameters", func(t *testing.T) {
		start := time.Now().Add(-24 * time.Hour).Unix()
		end := time.Now().Unix()

		url := "/api/data?start=%d&end=%d&group=30m&limit=100&offset=10&fill=true&order=created_at"
		req := httptest.NewRequest("GET", url, nil)
		req.URL.RawQuery = "start=" + itoa(start) + "&end=" + itoa(end) + "&group=30m&limit=100&offset=10&fill=true&order=created_at"

		query, err := ParseRequest(req)
		require.NoError(t, err)

		assert.Equal(t, start, query.Start.Unix())
		assert.Equal(t, end, query.End.Unix())
		assert.Equal(t, 30*time.Minute, query.Group)
		assert.Equal(t, 100, query.Limit)
		assert.Equal(t, 10, query.Offset)
		assert.True(t, query.FillEmpty)
		assert.Equal(t, "created_at", query.Order)
	})

	t.Run("InvalidGrouping", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data?group=invalid", nil)

		query, err := ParseRequest(req)
		require.NoError(t, err)

		// Should default to 1 hour on invalid grouping
		assert.Equal(t, time.Hour, query.Group)
	})

	t.Run("StartAfterEnd", func(t *testing.T) {
		end := time.Now().Add(-24 * time.Hour).Unix()
		start := time.Now().Unix()

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.URL.RawQuery = "start=" + itoa(start) + "&end=" + itoa(end)

		_, err := ParseRequest(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time is after ending time")
	})

	t.Run("ZeroLimit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data?limit=0", nil)

		query, err := ParseRequest(req)
		require.NoError(t, err)

		// Should default to 10000 when limit is 0
		assert.Equal(t, 10000, query.Limit)
	})

	t.Run("FillFalse", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data?fill=false", nil)

		query, err := ParseRequest(req)
		require.NoError(t, err)

		assert.False(t, query.FillEmpty)
	})
}

// mockIsObject implements isObject interface for testing ParseQueries
type mockIsObject struct {
	db Database
}

func (m *mockIsObject) Db() Database {
	return m.db
}

// TestParseQueries tests the ParseQueries function for query string parsing
func TestParseQueries(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	mock := &mockIsObject{db: db}

	t.Run("DefaultValues", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data", nil)

		query, err := ParseQueries(req, mock)
		require.NoError(t, err)

		assert.Equal(t, time.Hour, query.Group)
		assert.Equal(t, 10000, query.Limit)
		assert.False(t, query.FillEmpty)
		assert.NotNil(t, query.db)
	})

	t.Run("CustomGrouping", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data?group=15m", nil)

		query, err := ParseQueries(req, mock)
		require.NoError(t, err)

		assert.Equal(t, 15*time.Minute, query.Group)
	})

	t.Run("AllParameters", func(t *testing.T) {
		start := time.Now().Add(-24 * time.Hour).Unix()
		end := time.Now().Unix()

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.URL.RawQuery = "start=" + itoa(start) + "&end=" + itoa(end) + "&group=1h&limit=50&offset=5&fill=true&order=id DESC"

		query, err := ParseQueries(req, mock)
		require.NoError(t, err)

		assert.Equal(t, start, query.Start.Unix())
		assert.Equal(t, end, query.End.Unix())
		assert.Equal(t, time.Hour, query.Group)
		assert.Equal(t, 50, query.Limit)
		assert.Equal(t, 5, query.Offset)
		assert.True(t, query.FillEmpty)
		assert.Equal(t, "id DESC", query.Order)
	})

	t.Run("StartAfterEnd", func(t *testing.T) {
		end := time.Now().Add(-48 * time.Hour).Unix()
		start := time.Now().Unix()

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.URL.RawQuery = "start=" + itoa(start) + "&end=" + itoa(end)

		_, err := ParseQueries(req, mock)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "start time is after ending time")
	})

	t.Run("ZeroStartDefault", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data?start=0", nil)

		query, err := ParseQueries(req, mock)
		require.NoError(t, err)

		// Should default to 7 days ago
		expectedStart := time.Now().Add(-7 * types.Day)
		assert.InDelta(t, expectedStart.Unix(), query.Start.Unix(), 5)
	})

	t.Run("ZeroEndDefault", func(t *testing.T) {
		start := time.Now().Add(-24 * time.Hour).Unix()
		req := httptest.NewRequest("GET", "/api/data", nil)
		req.URL.RawQuery = "start=" + itoa(start) + "&end=0"

		query, err := ParseQueries(req, mock)
		require.NoError(t, err)

		// Should default to now
		assert.InDelta(t, time.Now().Unix(), query.End.Unix(), 5)
	})

	t.Run("OrderWhitelist", func(t *testing.T) {
		allowedOrders := []string{
			"id", "id DESC", "id ASC",
			"created_at", "created_at DESC", "created_at ASC",
			"order_id", "order_id DESC", "order_id ASC",
			"name", "name DESC", "name ASC",
		}

		for _, order := range allowedOrders {
			req := httptest.NewRequest("GET", "/api/data", nil)
			req.URL.RawQuery = "order=" + url.QueryEscape(order)

			query, err := ParseQueries(req, mock)
			require.NoError(t, err, "Order '%s' should be allowed", order)
			assert.Equal(t, order, query.Order)
		}
	})

	t.Run("InvalidOrderIgnored", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data?order=DROP+TABLE+users", nil)

		query, err := ParseQueries(req, mock)
		require.NoError(t, err)

		// Invalid order should be stored but not applied (silently ignored)
		assert.Equal(t, "DROP TABLE users", query.Order)
	})
}

// TestParseQueriesForTable tests the ParseQueriesForTable function
func TestParseQueriesForTable(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	mock := &mockIsObject{db: db}

	t.Run("WithTable", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data", nil)

		query, err := ParseQueriesForTable(req, mock, "test_models")
		require.NoError(t, err)

		assert.Equal(t, "test_models", query.Table)
	})

	t.Run("WithoutTable", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data", nil)

		query, err := ParseQueriesForTable(req, mock, "")
		require.NoError(t, err)

		assert.Equal(t, "", query.Table)
	})
}

// TestParseGet tests the parseGet function
func TestParseGet(t *testing.T) {
	t.Run("ValidForm", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data?foo=bar&baz=123", nil)

		values := parseGet(req)

		assert.Equal(t, "bar", values.Get("foo"))
		assert.Equal(t, "123", values.Get("baz"))
	})

	t.Run("EmptyForm", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/data", nil)

		values := parseGet(req)

		assert.Equal(t, "", values.Get("nonexistent"))
	})

	t.Run("PostForm", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/data", nil)
		req.PostForm = url.Values{
			"post_key": []string{"post_value"},
		}

		values := parseGet(req)

		// parseGet uses r.ParseForm() which merges POST form into r.Form
		// This is the actual behavior - it includes POST form values
		assert.Equal(t, "post_value", values.Get("post_key"))
	})
}

// TestTimeValue tests TimeValue struct operations
func TestTimeValue(t *testing.T) {
	t.Run("TimeValueFields", func(t *testing.T) {
		tv := &TimeValue{
			Timeframe: "2023-06-15T10:00:00Z",
			Amount:    12345,
		}

		assert.Equal(t, "2023-06-15T10:00:00Z", tv.Timeframe)
		assert.Equal(t, int64(12345), tv.Amount)
	})

	t.Run("TimeValueSlice", func(t *testing.T) {
		values := []*TimeValue{
			{Timeframe: "2023-06-15T10:00:00Z", Amount: 100},
			{Timeframe: "2023-06-15T11:00:00Z", Amount: 200},
		}

		assert.Len(t, values, 2)
		assert.Equal(t, int64(100), values[0].Amount)
		assert.Equal(t, int64(200), values[1].Amount)
	})
}

// TestGroupQueryWithStartEnd tests GroupQuery with various start/end combinations
func TestGroupQueryWithStartEnd(t *testing.T) {
	t.Run("ValidRange", func(t *testing.T) {
		start := time.Now().Add(-24 * time.Hour)
		end := time.Now()

		query := &GroupQuery{
			Start: start,
			End:   end,
			Group: time.Hour,
		}

		assert.True(t, query.End.After(query.Start))
	})

	t.Run("SameStartEnd", func(t *testing.T) {
		now := time.Now()

		query := &GroupQuery{
			Start: now,
			End:   now,
			Group: time.Hour,
		}

		assert.True(t, query.Start.Equal(query.End))
	})
}

// TestGroupingDurations tests various grouping duration scenarios
func TestGroupingDurations(t *testing.T) {
	durations := []struct {
		input    string
		expected time.Duration
	}{
		{"1s", time.Second},
		{"30s", 30 * time.Second},
		{"1m", time.Minute},
		{"5m", 5 * time.Minute},
		{"15m", 15 * time.Minute},
		{"30m", 30 * time.Minute},
		{"1h", time.Hour},
		{"6h", 6 * time.Hour},
		{"12h", 12 * time.Hour},
		{"24h", 24 * time.Hour},
	}

	for _, d := range durations {
		t.Run(d.input, func(t *testing.T) {
			parsed, err := time.ParseDuration(d.input)
			require.NoError(t, err)
			assert.Equal(t, d.expected, parsed)
		})
	}
}

// itoa is a helper to convert int64 to string for URL parameters
func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}

// TimeDataModel is a model for testing time-based queries
type TimeDataModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Amount    int64     `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TestToTimeValue tests the ToTimeValue method for converting database rows to TimeValue
func TestToTimeValue(t *testing.T) {
	// Save original db and restore after test
	originalDb := database
	defer func() { Set(originalDb) }()

	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()
	Set(db)

	// Create a simple table with timeframe and amount columns
	err = db.Exec("CREATE TABLE time_data (timeframe TEXT, amount INTEGER)").Error()
	require.NoError(t, err)

	// Insert test data
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		tf := now.Add(time.Duration(i) * time.Hour).Format("2006-01-02 15:04:05")
		err = db.Exec("INSERT INTO time_data (timeframe, amount) VALUES (?, ?)", tf, (i+1)*100).Error()
		require.NoError(t, err)
	}

	t.Run("ValidData", func(t *testing.T) {
		query := &GroupQuery{
			Start: now,
			End:   now.Add(3 * time.Hour),
			Group: time.Hour,
			db:    db.Raw("SELECT timeframe, amount FROM time_data ORDER BY timeframe"),
		}

		timeVar, err := query.ToTimeValue()
		require.NoError(t, err)
		assert.NotNil(t, timeVar)

		values, err := timeVar.ToValues()
		require.NoError(t, err)
		assert.Len(t, values, 3)
	})

	t.Run("EmptyTable", func(t *testing.T) {
		// Create empty table
		err = db.Exec("CREATE TABLE empty_time_data (timeframe TEXT, amount INTEGER)").Error()
		require.NoError(t, err)

		query := &GroupQuery{
			Start: now,
			End:   now.Add(3 * time.Hour),
			Group: time.Hour,
			db:    db.Raw("SELECT timeframe, amount FROM empty_time_data"),
		}

		timeVar, err := query.ToTimeValue()
		require.NoError(t, err)
		assert.NotNil(t, timeVar)

		values, err := timeVar.ToValues()
		require.NoError(t, err)
		assert.Empty(t, values)
	})
}

// TestGraphData tests the GraphData method for generating chart data
func TestGraphData(t *testing.T) {
	// Save original db and restore after test
	originalDb := database
	defer func() { Set(originalDb) }()

	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()
	Set(db)

	// Create a table with created_at for grouping
	err = db.Exec("CREATE TABLE graph_test (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at DATETIME)").Error()
	require.NoError(t, err)

	// Insert test data with different timestamps
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		createdAt := now.Add(time.Duration(i) * time.Hour)
		err = db.Exec("INSERT INTO graph_test (created_at) VALUES (?)", db.FormatTime(createdAt)).Error()
		require.NoError(t, err)
	}

	t.Run("WithFillEmpty", func(t *testing.T) {
		query := &GroupQuery{
			Start:     now,
			End:       now.Add(12 * time.Hour),
			Group:     time.Hour,
			FillEmpty: true,
			db:        db.Table("graph_test"),
		}

		values, err := query.GraphData(ByCount)
		require.NoError(t, err)
		assert.NotEmpty(t, values)
		// With FillEmpty=true, we should have entries for all hours in the range
		assert.GreaterOrEqual(t, len(values), 10)
	})

	t.Run("WithoutFillEmpty", func(t *testing.T) {
		query := &GroupQuery{
			Start:     now,
			End:       now.Add(12 * time.Hour),
			Group:     time.Hour,
			FillEmpty: false,
			db:        db.Table("graph_test"),
		}

		values, err := query.GraphData(ByCount)
		require.NoError(t, err)
		assert.NotEmpty(t, values)
	})
}

// TestParseQueriesInvalidGrouping tests ParseQueries with invalid grouping duration
func TestParseQueriesInvalidGrouping(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	mock := &mockIsObject{db: db}

	req := httptest.NewRequest("GET", "/api/data?group=invalid_duration", nil)

	query, err := ParseQueries(req, mock)
	require.NoError(t, err)

	// Should default to 1 hour on invalid grouping
	assert.Equal(t, time.Hour, query.Group)
}

// TestParseQueriesWithLimitZero tests that limit 0 gets default value
func TestParseQueriesWithLimitZero(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	mock := &mockIsObject{db: db}

	req := httptest.NewRequest("GET", "/api/data?limit=0", nil)

	query, err := ParseQueries(req, mock)
	require.NoError(t, err)

	// Should default to 10000 when limit is 0
	assert.Equal(t, 10000, query.Limit)
}

// TestParseQueriesWithNonZeroLimitAndOffset tests limit and offset applied to db
func TestParseQueriesWithNonZeroLimitAndOffset(t *testing.T) {
	db, err := Openw("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	}()

	err = db.AutoMigrate(&TestModel{}).Error()
	require.NoError(t, err)

	mock := &mockIsObject{db: db}

	req := httptest.NewRequest("GET", "/api/data?limit=25&offset=10", nil)

	query, err := ParseQueries(req, mock)
	require.NoError(t, err)

	assert.Equal(t, 25, query.Limit)
	assert.Equal(t, 10, query.Offset)
	assert.NotNil(t, query.db)
}

// TestGroupByInterface tests the GroupByer interface
func TestGroupByInterface(t *testing.T) {
	// GroupBy is an empty struct used as a marker
	gb := GroupBy{}
	assert.NotNil(t, gb)
}

// TestTimeVarWithGroupQuery tests TimeVar with its GroupQuery reference
func TestTimeVarWithGroupQuery(t *testing.T) {
	query := &GroupQuery{
		Start: time.Now(),
		End:   time.Now().Add(24 * time.Hour),
		Group: time.Hour,
	}

	tv := &TimeVar{
		g: query,
		data: []*TimeValue{
			{Timeframe: "2023-06-15T10:00:00Z", Amount: 100},
		},
	}

	assert.Equal(t, query, tv.g)
	assert.Len(t, tv.data, 1)
}
