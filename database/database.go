package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/statping-ng/statping-ng/types/metrics"
	"github.com/statping-ng/statping-ng/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var database Database

// Database is an interface which DB implements
type Database interface {
	Close() error
	DB() (*sql.DB, error)
	New() Database
	SetLogger(l logger.Interface)
	LogMode(enable bool) Database
	Where(query interface{}, args ...interface{}) Database
	Or(query interface{}, args ...interface{}) Database
	Not(query interface{}, args ...interface{}) Database
	Limit(value int) Database
	Offset(value int) Database
	Order(value interface{}) Database
	Select(query interface{}, args ...interface{}) Database
	Omit(columns ...string) Database
	Group(name string) Database
	Having(query interface{}, args ...interface{}) Database
	Joins(query string, args ...interface{}) Database
	Scopes(funcs ...func(*gorm.DB) *gorm.DB) Database
	Unscoped() Database
	First(out interface{}, where ...interface{}) Database
	Last(out interface{}, where ...interface{}) Database
	Find(out interface{}, where ...interface{}) Database
	Scan(dest interface{}) Database
	Row() *sql.Row
	Rows() (*sql.Rows, error)
	ScanRows(rows *sql.Rows, result interface{}) error
	Pluck(column string, value interface{}) Database
	Count(value *int64) Database
	Update(attrs ...interface{}) Database
	Updates(values interface{}) Database
	UpdateColumn(attrs ...interface{}) Database
	UpdateColumns(values interface{}) Database
	Save(value interface{}) Database
	Create(value interface{}) Database
	Delete(value interface{}, where ...interface{}) Database
	Raw(sql string, values ...interface{}) Database
	Exec(sql string, values ...interface{}) Database
	Model(value interface{}) Database
	Table(name string) Database
	Debug() Database
	Begin() Database
	Commit() Database
	Rollback() Database
	RecordNotFound() bool
	HasTable(value interface{}) bool
	AutoMigrate(values ...interface{}) Database
	Association(column string) *gorm.Association
	Preload(query string, args ...interface{}) Database
	Set(name string, value interface{}) Database
	Get(name string) (value interface{}, ok bool)
	AddError(err error) error

	// migration tools (v1 compatibility)
	CreateTable(values ...interface{}) Database
	DropTable(values ...interface{}) Database
	DropTableIfExists(values ...interface{}) Database
	AddIndex(indexName string, columns ...string) Database
	AddUniqueIndex(indexName string, columns ...string) Database

	// extra
	Error() error
	Status() int
	RowsAffected() int64

	Since(time.Time) Database
	Between(time.Time, time.Time) Database

	SelectByTime(time.Duration) string
	MultipleSelects(args ...string) Database

	FormatTime(t time.Time) string
	ParseTime(t string) (time.Time, error)
	DbType() string
	GormDB() *gorm.DB
	ChunkSize() int
}

func (it *Db) ChunkSize() int {
	switch it.Type {
	case "mysql":
		return 3000
	case "postgres":
		return 3000
	default:
		return 100
	}
}

func Routine() {
	for {
		sqlDB, err := database.DB()
		if err != nil || sqlDB == nil {
			time.Sleep(5 * time.Second)
			continue
		}
		metrics.CollectDatabase(sqlDB.Stats())
		time.Sleep(5 * time.Second)
	}
}

func (it *Db) GormDB() *gorm.DB {
	return it.Database
}

func (it *Db) DbType() string {
	return it.Type
}

func Close(db Database) error {
	if db == nil {
		return nil
	}
	return db.Close()
}

func LogMode(db Database, b bool) Database {
	return db.LogMode(b)
}

func Begin(db Database, model interface{}) Database {
	if all, ok := model.(string); ok {
		if all == "migration" {
			return db.Begin()
		}
	}
	return db.Model(model).Begin()
}

func Available(db Database) bool {
	if db == nil {
		return false
	}
	sqlDB, err := db.DB()
	if err != nil {
		return false
	}
	if err := sqlDB.Ping(); err != nil {
		return false
	}
	return true
}

func (it *Db) MultipleSelects(args ...string) Database {
	joined := strings.Join(args, ", ")
	return it.Select(joined)
}

type Db struct {
	Database *gorm.DB
	Type     string
	ReadOnly bool
}

// Openw is a drop-in replacement for Open()
func Openw(dialect string, args ...interface{}) (db Database, err error) {
	if dialect == "sqlite" {
		dialect = "sqlite3"
	}
	dsn := args[0].(string)
	var dialector gorm.Dialector

	switch dialect {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "postgres":
		dialector = postgres.Open(dsn)
	case "sqlite3":
		dialector = sqlite.Open(dsn)
	case "mssql":
		dialector = sqlserver.Open(dsn)
	}

	gormdb, err := gorm.Open(dialector, &gorm.Config{
		NowFunc: func() time.Time {
			return utils.Now()
		},
	})
	if err != nil {
		return nil, err
	}
	database = Wrap(gormdb)
	go Routine()
	return database, err
}

func OpenTester() (Database, error) {
	testDB := utils.Params.GetString("DB_CONN")
	var dbString string

	switch testDB {
	case "mysql":
		dbString = fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8&parseTime=True&loc=UTC&time_zone=%%27UTC%%27",
			utils.Params.GetString("DB_HOST"),
			utils.Params.GetString("DB_PASS"),
			utils.Params.GetString("DB_HOST"),
			utils.Params.GetInt("DB_PORT"),
			utils.Params.GetString("DB_DATABASE"),
		)
	case "postgres":
		dbString = fmt.Sprintf("host=%s port=%v user=%s dbname=%s password=%s sslmode=disable timezone=UTC",
			utils.Params.GetString("DB_HOST"),
			utils.Params.GetInt("DB_PORT"),
			utils.Params.GetString("DB_USER"),
			utils.Params.GetString("DB_DATABASE"),
			utils.Params.GetString("DB_PASS"))
	default:
		dbString = fmt.Sprintf("file:%s?mode=memory&cache=shared", utils.RandomString(12))
	}
	if utils.Params.IsSet("DB_DSN") {
		dbString = utils.Params.GetString("DB_DSN")
	}
	newDb, err := Openw(testDB, dbString)
	if err != nil {
		return nil, err
	}
	sqlDB, _ := newDb.DB()
	sqlDB.SetMaxOpenConns(1)
	if testDB != "sqlite3" {
		sqlDB.SetMaxOpenConns(25)
	}
	return newDb, err
}

// Wrap wraps gorm.DB in an interface
func (it *Db) wrap(db *gorm.DB) Database {
	return &Db{
		Database: db,
		Type:     it.Type,
		ReadOnly: it.ReadOnly,
	}
}

func Wrap(db *gorm.DB) Database {
	return &Db{
		Database: db,
		Type:     db.Name(),
		ReadOnly: utils.Params.GetBool("READ_ONLY"),
	}
}

func (it *Db) Close() error {
	sqlDB, err := it.Database.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (it *Db) DB() (*sql.DB, error) {
	return it.Database.DB()
}

func (it *Db) New() Database {
	return it.wrap(it.Database.Session(&gorm.Session{}))
}

func (it *Db) SetLogger(log logger.Interface) {
	it.Database.Logger = log
}

func (it *Db) LogMode(enable bool) Database {
	if enable {
		return it.wrap(it.Database.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Info)}))
	}
	return it.wrap(it.Database.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}))
}

func (it *Db) Where(query interface{}, args ...interface{}) Database {
	return it.wrap(it.Database.Where(query, args...))
}

func (it *Db) Or(query interface{}, args ...interface{}) Database {
	return it.wrap(it.Database.Or(query, args...))
}

func (it *Db) Not(query interface{}, args ...interface{}) Database {
	return it.wrap(it.Database.Not(query, args...))
}

func (it *Db) Limit(value int) Database {
	return it.wrap(it.Database.Limit(value))
}

func (it *Db) Offset(value int) Database {
	return it.wrap(it.Database.Offset(value))
}

func (it *Db) Order(value interface{}) Database {
	return it.wrap(it.Database.Order(value))
}

func (it *Db) Select(query interface{}, args ...interface{}) Database {
	return it.wrap(it.Database.Select(query, args...))
}

func (it *Db) Omit(columns ...string) Database {
	return it.wrap(it.Database.Omit(columns...))
}

func (it *Db) Group(name string) Database {
	return it.wrap(it.Database.Group(name))
}

func (it *Db) Having(query interface{}, args ...interface{}) Database {
	return it.wrap(it.Database.Having(query, args...))
}

func (it *Db) Joins(query string, args ...interface{}) Database {
	return it.wrap(it.Database.Joins(query, args...))
}

func (it *Db) Scopes(funcs ...func(*gorm.DB) *gorm.DB) Database {
	return it.wrap(it.Database.Scopes(funcs...))
}

func (it *Db) Unscoped() Database {
	return it.wrap(it.Database.Unscoped())
}

func (it *Db) First(out interface{}, where ...interface{}) Database {
	return it.wrap(it.Database.First(out, where...))
}

func (it *Db) Last(out interface{}, where ...interface{}) Database {
	return it.wrap(it.Database.Last(out, where...))
}

func (it *Db) Find(out interface{}, where ...interface{}) Database {
	return it.wrap(it.Database.Find(out, where...))
}

func (it *Db) Scan(dest interface{}) Database {
	return it.wrap(it.Database.Scan(dest))
}

func (it *Db) Row() *sql.Row {
	return it.Database.Row()
}

func (it *Db) Rows() (*sql.Rows, error) {
	return it.Database.Rows()
}

func (it *Db) ScanRows(rows *sql.Rows, result interface{}) error {
	return it.Database.ScanRows(rows, result)
}

func (it *Db) Pluck(column string, value interface{}) Database {
	return Wrap(it.Database.Pluck(column, value))
}

func (it *Db) Count(value *int64) Database {
	return Wrap(it.Database.Count(value))
}

func (it *Db) Update(attrs ...interface{}) Database {
	if it.ReadOnly {
		return it
	}
	if len(attrs) == 1 {
		return Wrap(it.Database.Updates(attrs[0]))
	}
	if len(attrs) == 2 {
		if col, ok := attrs[0].(string); ok {
			return Wrap(it.Database.Update(col, attrs[1]))
		}
	}
	return it
}

func (it *Db) Updates(values interface{}) Database {
	if it.ReadOnly {
		return it
	}
	return it.wrap(it.Database.Updates(values))
}

func (it *Db) UpdateColumn(attrs ...interface{}) Database {
	if it.ReadOnly {
		return it
	}
	if len(attrs) == 1 {
		return it.wrap(it.Database.UpdateColumns(attrs[0]))
	}
	if len(attrs) == 2 {
		if col, ok := attrs[0].(string); ok {
			return it.wrap(it.Database.UpdateColumn(col, attrs[1]))
		}
	}
	return it
}

func (it *Db) UpdateColumns(values interface{}) Database {
	if it.ReadOnly {
		return it
	}
	return it.wrap(it.Database.UpdateColumns(values))
}

func (it *Db) Save(value interface{}) Database {
	if it.ReadOnly {
		return it
	}
	return it.wrap(it.Database.Save(value))
}

func (it *Db) Create(value interface{}) Database {
	if it.ReadOnly {
		return it
	}
	return it.wrap(it.Database.Create(value))
}

func (it *Db) Delete(value interface{}, where ...interface{}) Database {
	if it.ReadOnly {
		return it
	}
	return it.wrap(it.Database.Delete(value, where...))
}

func (it *Db) Raw(sql string, values ...interface{}) Database {
	return it.wrap(it.Database.Raw(sql, values...))
}

func (it *Db) Exec(sql string, values ...interface{}) Database {
	return it.wrap(it.Database.Exec(sql, values...))
}

func (it *Db) Model(value interface{}) Database {
	return it.wrap(it.Database.Model(value))
}

func (it *Db) Table(name string) Database {
	return it.wrap(it.Database.Table(name))
}

func (it *Db) Debug() Database {
	return Wrap(it.Database.Debug())
}

func (it *Db) Begin() Database {
	if it.ReadOnly {
		return it
	}
	return Wrap(it.Database.Begin())
}

func (it *Db) Commit() Database {
	if it.ReadOnly {
		return it
	}
	return Wrap(it.Database.Commit())
}

func (it *Db) Rollback() Database {
	if it.ReadOnly {
		return it
	}
	return Wrap(it.Database.Rollback())
}

func (it *Db) RecordNotFound() bool {
	return errors.Is(it.Database.Error, gorm.ErrRecordNotFound)
}

func (it *Db) HasTable(value interface{}) bool {
	return it.Database.Migrator().HasTable(value)
}

func (it *Db) AutoMigrate(values ...interface{}) Database {
	if it.ReadOnly {
		return it
	}
	if err := it.Database.AutoMigrate(values...); err != nil {
		it.Database.Error = err
	}
	return it
}

func (it *Db) CreateTable(values ...interface{}) Database {
	if it.ReadOnly {
		return it
	}
	if err := it.Database.Migrator().CreateTable(values...); err != nil {
		it.Database.Error = err
	}
	return it
}

func (it *Db) DropTable(values ...interface{}) Database {
	if it.ReadOnly {
		return it
	}
	if err := it.Database.Migrator().DropTable(values...); err != nil {
		it.Database.Error = err
	}
	return it
}

func (it *Db) DropTableIfExists(values ...interface{}) Database {
	if it.ReadOnly {
		return it
	}
	for _, v := range values {
		if err := it.Database.Migrator().DropTable(v); err != nil {
			it.Database.Error = err
		}
	}
	return it
}

func (it *Db) AddIndex(indexName string, columns ...string) Database {
	if it.ReadOnly {
		return it
	}
	if len(columns) > 0 {
		table := it.Database.Statement.Table
		if table == "" && it.Database.Statement.Model != nil {
			stmt := &gorm.Statement{DB: it.Database, ConnPool: it.Database.ConnPool, Context: it.Database.Statement.Context}
			if err := stmt.Parse(it.Database.Statement.Model); err == nil {
				table = stmt.Table
			}
		}
		if table != "" {
			sql := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", indexName, table, strings.Join(columns, ", "))
			it.Database.Exec(sql)
			return it
		}
	}
	if err := it.Database.Migrator().CreateIndex(it.Database.Statement.Model, indexName); err != nil {
		it.Database.Error = err
	}
	return it
}

func (it *Db) AddUniqueIndex(indexName string, columns ...string) Database {
	if it.ReadOnly {
		return it
	}
	if err := it.Database.Migrator().CreateIndex(it.Database.Statement.Model, indexName); err != nil {
		it.Database.Error = err
	}
	return it
}

func (it *Db) Association(column string) *gorm.Association {
	return it.Database.Association(column)
}

func (it *Db) Preload(query string, args ...interface{}) Database {
	return it.wrap(it.Database.Preload(query, args...))
}

func (it *Db) Set(name string, value interface{}) Database {
	return it.wrap(it.Database.Set(name, value))
}

func (it *Db) Get(name string) (interface{}, bool) {
	return it.Database.Get(name)
}

func (it *Db) AddError(err error) error {
	return it.Database.AddError(err)
}

func (it *Db) RowsAffected() int64 {
	return it.Database.RowsAffected
}

func (it *Db) Error() error {
	return it.Database.Error
}

func (it *Db) Status() int {
	if errors.Is(it.Database.Error, gorm.ErrRecordNotFound) {
		return 404
	}
	if it.Database.Error != nil {
		return 500
	}
	return 200
}

func (it *Db) Since(ago time.Time) Database {
	return it.Where("created_at > ?", it.FormatTime(ago))
}

func (it *Db) Between(t1 time.Time, t2 time.Time) Database {
	return it.Where("created_at BETWEEN ? AND ?", it.FormatTime(t1), it.FormatTime(t2))
}

type TimeValue struct {
	Timeframe string `json:"timeframe"`
	Amount    int64  `json:"amount"`
}
