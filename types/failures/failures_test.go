package failures

import (
	"os"
	"sort"
	"testing"
	"time"

	"github.com/statping-ng/statping-ng/database"
	"github.com/statping-ng/statping-ng/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	tmpDir, err := os.MkdirTemp("", "statping-failures-test")
	if err != nil {
		os.Exit(1)
	}
	utils.Directory = tmpDir
	utils.InitEnvs()
	utils.Params.Set("STATPING_DIR", tmpDir)
	_ = utils.InitLogs()

	code := m.Run()

	_ = os.RemoveAll(tmpDir)
	os.Exit(code)
}

// mockColumnID implements ColumnIDInterfacer for testing
type mockColumnID struct {
	column string
	id     int64
}

func (m mockColumnID) FailuresColumnID() (string, int64) {
	return m.column, m.id
}

func setupTestDB(t *testing.T) database.Database {
	testDb, err := database.OpenTester()
	require.NoError(t, err)
	require.NoError(t, testDb.AutoMigrate(&Failure{}).Error())
	SetDB(testDb)
	return testDb
}

// ============================================================================
// Database CRUD Tests
// ============================================================================

func TestSetDB(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	assert.NotNil(t, DB())
	assert.Equal(t, testDb, db)
}

func TestDB(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	result := DB()
	assert.NotNil(t, result)
}

func TestFailure_Create(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	f := &Failure{
		Issue:     "Connection timeout",
		Method:    "GET",
		MethodId:  1,
		ErrorCode: 500,
		Service:   1,
		PingTime:  1500,
		Reason:    "timeout",
		CreatedAt: time.Now(),
	}

	err := f.Create()
	require.NoError(t, err)
	assert.True(t, f.Id > 0)
}

func TestFailure_Update(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	f := &Failure{
		Issue:     "Original issue",
		ErrorCode: 500,
		Service:   1,
		CreatedAt: time.Now(),
	}
	require.NoError(t, f.Create())

	f.Issue = "Updated issue"
	f.ErrorCode = 502
	err := f.Update()
	require.NoError(t, err)

	// Verify update by querying
	var updated Failure
	testDb.Model(&Failure{}).Where("id = ?", f.Id).First(&updated)
	assert.Equal(t, "Updated issue", updated.Issue)
	assert.Equal(t, 502, updated.ErrorCode)
}

func TestFailure_Delete(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	f := &Failure{
		Issue:     "To be deleted",
		ErrorCode: 404,
		Service:   1,
		CreatedAt: time.Now(),
	}
	require.NoError(t, f.Create())
	originalId := f.Id

	err := f.Delete()
	require.NoError(t, err)

	// Verify deletion
	var count int64
	testDb.Model(&Failure{}).Where("id = ?", originalId).Count(&count)
	assert.Equal(t, int64(0), count)
}

// ============================================================================
// GORM Hooks Tests
// ============================================================================

func TestFailure_AfterFind(t *testing.T) {
	f := &Failure{}
	err := f.AfterFind(&gorm.DB{})
	assert.NoError(t, err)
}

func TestFailure_AfterUpdate(t *testing.T) {
	f := &Failure{}
	err := f.AfterUpdate(&gorm.DB{})
	assert.NoError(t, err)
}

func TestFailure_AfterDelete(t *testing.T) {
	f := &Failure{}
	err := f.AfterDelete(&gorm.DB{})
	assert.NoError(t, err)
}

func TestFailure_AfterCreate(t *testing.T) {
	f := &Failure{}
	err := f.AfterCreate(&gorm.DB{})
	assert.NoError(t, err)
}

// ============================================================================
// Failurer Interface Tests
// ============================================================================

func TestFailurer_Db(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	failurer := Failurer{db: testDb.Model(&Failure{})}
	assert.NotNil(t, failurer.Db())
}

func TestFailurer_First(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Create multiple failures
	for i := 0; i < 3; i++ {
		f := &Failure{
			Issue:     "Test failure",
			ErrorCode: 500 + i,
			Service:   1,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Hour),
		}
		require.NoError(t, f.Create())
	}

	failurer := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 1)}
	first := failurer.First()
	assert.NotNil(t, first)
	assert.True(t, first.Id > 0)
	// First should have lowest ID
	assert.Equal(t, 500, first.ErrorCode)
}

func TestFailurer_Last(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Create multiple failures
	for i := 0; i < 3; i++ {
		f := &Failure{
			Issue:     "Test failure",
			ErrorCode: 500 + i,
			Service:   1,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Hour),
		}
		require.NoError(t, f.Create())
	}

	failurer := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 1)}
	last := failurer.Last()
	assert.NotNil(t, last)
	assert.True(t, last.Id > 0)
	// Last should have highest ID
	assert.Equal(t, 502, last.ErrorCode)
}

func TestFailurer_List(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Create failures for different services
	for i := 0; i < 5; i++ {
		f := &Failure{
			Issue:     "Test failure",
			ErrorCode: 500,
			Service:   1,
			CreatedAt: time.Now(),
		}
		require.NoError(t, f.Create())
	}
	for i := 0; i < 3; i++ {
		f := &Failure{
			Issue:     "Test failure",
			ErrorCode: 500,
			Service:   2,
			CreatedAt: time.Now(),
		}
		require.NoError(t, f.Create())
	}

	failurer := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 1)}
	list := failurer.List()
	assert.Len(t, list, 5)
}

func TestFailurer_LastAmount(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Create 5 failures
	for i := 0; i < 5; i++ {
		f := &Failure{
			Issue:     "Test failure",
			ErrorCode: 500 + i,
			Service:   1,
			CreatedAt: time.Now(),
		}
		require.NoError(t, f.Create())
	}

	failurer := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 1)}
	limited := failurer.LastAmount(3)
	assert.Len(t, limited, 3)
	// Should be in DESC order (latest first)
	assert.True(t, limited[0].Id > limited[1].Id)
}

func TestFailurer_Since(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	now := time.Now()
	// Create old failure
	oldFailure := &Failure{
		Issue:     "Old failure",
		ErrorCode: 500,
		Service:   1,
		CreatedAt: now.Add(-48 * time.Hour),
	}
	require.NoError(t, oldFailure.Create())

	// Create recent failures
	for i := 0; i < 3; i++ {
		f := &Failure{
			Issue:     "Recent failure",
			ErrorCode: 500,
			Service:   1,
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
		require.NoError(t, f.Create())
	}

	failurer := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 1)}
	since := failurer.Since(now.Add(-24 * time.Hour))
	assert.Len(t, since, 3)
}

func TestFailurer_Count(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Create failures
	for i := 0; i < 7; i++ {
		f := &Failure{
			Issue:     "Test failure",
			ErrorCode: 500,
			Service:   1,
			CreatedAt: time.Now(),
		}
		require.NoError(t, f.Create())
	}

	failurer := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 1)}
	count := failurer.Count()
	assert.Equal(t, 7, count)
}

func TestFailurer_DeleteAll(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Create failures for service 1
	for i := 0; i < 5; i++ {
		f := &Failure{
			Issue:     "Test failure",
			ErrorCode: 500,
			Service:   1,
			CreatedAt: time.Now(),
		}
		require.NoError(t, f.Create())
	}
	// Create failures for service 2
	for i := 0; i < 3; i++ {
		f := &Failure{
			Issue:     "Test failure",
			ErrorCode: 500,
			Service:   2,
			CreatedAt: time.Now(),
		}
		require.NoError(t, f.Create())
	}

	// Delete all for service 1
	failurer := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 1)}
	err := failurer.DeleteAll()
	require.NoError(t, err)

	// Verify service 1 failures are deleted
	count := failurer.Count()
	assert.Equal(t, 0, count)

	// Verify service 2 failures still exist
	failurer2 := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 2)}
	count2 := failurer2.Count()
	assert.Equal(t, 3, count2)
}

// ============================================================================
// AllFailures and Since Function Tests
// ============================================================================

func TestAllFailures(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Create failures for a service
	for i := 0; i < 4; i++ {
		f := &Failure{
			Issue:     "Test failure",
			ErrorCode: 500,
			Service:   5,
			CreatedAt: time.Now(),
		}
		require.NoError(t, f.Create())
	}

	mock := mockColumnID{column: "service", id: 5}
	failurer := AllFailures(mock)
	assert.NotNil(t, failurer)

	list := failurer.List()
	assert.Len(t, list, 4)
}

func TestSince(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	now := time.Now()
	// Create old failure
	oldFailure := &Failure{
		Issue:     "Old failure",
		ErrorCode: 500,
		Service:   6,
		CreatedAt: now.Add(-72 * time.Hour),
	}
	require.NoError(t, oldFailure.Create())

	// Create recent failures
	for i := 0; i < 2; i++ {
		f := &Failure{
			Issue:     "Recent failure",
			ErrorCode: 500,
			Service:   6,
			CreatedAt: now.Add(-time.Duration(i) * time.Hour),
		}
		require.NoError(t, f.Create())
	}

	mock := mockColumnID{column: "service", id: 6}
	failurer := Since(now.Add(-24*time.Hour), mock)
	list := failurer.List()
	assert.Len(t, list, 2)
}

// ============================================================================
// FailSort (Sorting Interface) Tests
// ============================================================================

func TestFailSort_Len(t *testing.T) {
	fails := FailSort{
		{Id: 1}, {Id: 2}, {Id: 3},
	}
	assert.Equal(t, 3, fails.Len())

	emptyFails := FailSort{}
	assert.Equal(t, 0, emptyFails.Len())
}

func TestFailSort_Swap(t *testing.T) {
	fails := FailSort{
		{Id: 1, Issue: "First"},
		{Id: 2, Issue: "Second"},
	}
	fails.Swap(0, 1)
	assert.Equal(t, int64(2), fails[0].Id)
	assert.Equal(t, "Second", fails[0].Issue)
	assert.Equal(t, int64(1), fails[1].Id)
	assert.Equal(t, "First", fails[1].Issue)
}

func TestFailSort_Less(t *testing.T) {
	fails := FailSort{
		{Id: 1},
		{Id: 2},
		{Id: 3},
	}
	assert.True(t, fails.Less(0, 1))
	assert.True(t, fails.Less(0, 2))
	assert.True(t, fails.Less(1, 2))
	assert.False(t, fails.Less(1, 0))
	assert.False(t, fails.Less(2, 1))
}

func TestFailSort_Integration(t *testing.T) {
	// Test that FailSort works with sort.Sort
	fails := FailSort{
		{Id: 3, Issue: "Third"},
		{Id: 1, Issue: "First"},
		{Id: 2, Issue: "Second"},
	}

	sort.Sort(fails)

	assert.Equal(t, int64(1), fails[0].Id)
	assert.Equal(t, int64(2), fails[1].Id)
	assert.Equal(t, int64(3), fails[2].Id)
}

// ============================================================================
// Example and Samples Tests
// ============================================================================

func TestExample(t *testing.T) {
	example := Example()

	assert.Equal(t, int64(48533), example.Id)
	assert.Equal(t, "Response did not response a 200 status code", example.Issue)
	assert.Equal(t, 404, example.ErrorCode)
	assert.Equal(t, int64(1), example.Service)
	assert.Equal(t, int64(0), example.Checkin)
	assert.Equal(t, int64(48309), example.PingTime)
	assert.Equal(t, "status_code", example.Reason)
	assert.False(t, example.CreatedAt.IsZero())
}

func TestSamples(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	err := Samples()
	require.NoError(t, err)

	// Verify some failures were created (exact count depends on random chance)
	var count int64
	testDb.Model(&Failure{}).Count(&count)
	// Sample failures are created with a chance, so count may vary
	// Just verify the function ran without error
}

// ============================================================================
// Edge Cases and Error Handling Tests
// ============================================================================

func TestFailurer_EmptyResults(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Test with non-existent service
	failurer := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 999)}

	first := failurer.First()
	assert.Equal(t, int64(0), first.Id) // Should return zero-value failure

	last := failurer.Last()
	assert.Equal(t, int64(0), last.Id)

	list := failurer.List()
	assert.Len(t, list, 0)

	count := failurer.Count()
	assert.Equal(t, 0, count)
}

func TestFailure_AllFields(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	now := time.Now()
	f := &Failure{
		Issue:     "Complete test failure",
		Method:    "POST",
		MethodId:  42,
		ErrorCode: 503,
		Service:   10,
		Checkin:   5,
		PingTime:  2500,
		Reason:    "service_unavailable",
		CreatedAt: now,
	}

	require.NoError(t, f.Create())
	assert.True(t, f.Id > 0)

	// Query it back
	var retrieved Failure
	testDb.Model(&Failure{}).Where("id = ?", f.Id).First(&retrieved)

	assert.Equal(t, "Complete test failure", retrieved.Issue)
	assert.Equal(t, "POST", retrieved.Method)
	assert.Equal(t, int64(42), retrieved.MethodId)
	assert.Equal(t, 503, retrieved.ErrorCode)
	assert.Equal(t, int64(10), retrieved.Service)
	assert.Equal(t, int64(5), retrieved.Checkin)
	assert.Equal(t, int64(2500), retrieved.PingTime)
	assert.Equal(t, "service_unavailable", retrieved.Reason)
}

func TestFailurer_DeleteAll_Empty(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Try to delete from empty table
	failurer := Failurer{db: testDb.Model(&Failure{}).Where("service = ?", 999)}
	err := failurer.DeleteAll()
	// Should not error even when nothing to delete
	assert.NoError(t, err)
}

func TestAllFailures_Checkin(t *testing.T) {
	testDb := setupTestDB(t)
	defer testDb.Close()

	// Create failures for a checkin
	for i := 0; i < 3; i++ {
		f := &Failure{
			Issue:     "Checkin failure",
			ErrorCode: 500,
			Checkin:   7,
			CreatedAt: time.Now(),
		}
		require.NoError(t, f.Create())
	}

	mock := mockColumnID{column: "checkin", id: 7}
	failurer := AllFailures(mock)
	list := failurer.List()
	assert.Len(t, list, 3)
}
