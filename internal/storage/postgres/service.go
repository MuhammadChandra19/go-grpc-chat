package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"

	"github.com/MuhammadChandra19/go-grpc-chat/config"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/errors"
	"github.com/MuhammadChandra19/go-grpc-chat/internal/storage/postgres/migration"
	"github.com/jmoiron/sqlx"
)

type database struct {
	sqlxDB *sqlx.DB
	sqlxTX *sqlx.Tx
}

type table struct {
	TableName *string `json:"table_name" db:"table_name"`
}

type metadata struct {
	Key   string `json:"key" db:"key"`
	Value string `json:"value" db:"value"`
}

var (
	onceDB sync.Once
	db     *sqlx.DB
)

var (
	ErrResShouldBePtr    = errors.N(errors.CodeSystemError, "invalid response, response should be in pointers format")
	ErrResShouldBeStruct = errors.N(errors.CodeSystemError, "invalid response, response should be in pointers struct")
	ErrDataNotFound      = errors.N(errors.CodeNotFoundError, "data not found")
	ErrCreateTx          = errors.N(errors.CodeSystemError, "database error, create db transaction")
	ErrCommitTx          = errors.N(errors.CodeSystemError, "database error, commit db transaction")
	ErrRollbackTx        = errors.N(errors.CodeSystemError, "database error, rollback db transaction")
)

const (
	ptrStruct        = iota
	prtSliceOfStruct = iota
	ptrSingleType    = iota
)

type DatabaseInterface interface {
	Migrate(forceMigrate bool) error
	Exec(ctx context.Context, stmt string, params interface{}) error
	Query(ctx context.Context, query string, params, response interface{}, forUpdate bool) error
	RunInTransaction(ctx context.Context, f func(tctx context.Context) error) error
	GenerateQueryParams(query string, params map[string]interface{}, searchBy map[string]interface{}) string
	WithLimitOffset(query string, limit, offset int) string
	WithOrder(query string, orderBy, orderDir string) string
	NamedExec(ctx context.Context, stmt string, params interface{}) error
}

func (db *database) Migrate(forceMigrate bool) error {
	ctx := context.Background()
	if forceMigrate {
		_, err := db.sqlxDB.Exec(`DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO public;`)
		if err != nil {
			log.Println("Postgres Migrate: ", err)
			return err
		}
	}
	checktable := table{}
	emptyParams := map[string]interface{}{}
	err := db.Query(ctx, "SELECT to_regclass('metadata') as table_name", emptyParams, &checktable, false)
	if err != nil {
		log.Println("Postgres Migrate: ", err)
		return err
	}

	tx, err := db.sqlxDB.Begin()
	if err != nil {
		log.Println("Postgres Migrate: ", err)
		return err
	}

	migrationVersion := 0
	migrationFile := migration.Sequance
	if checktable.TableName == nil {
		_, err = tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS "metadata" (
			key VARCHAR (50) PRIMARY KEY,
			value VARCHAR (50) NOT NULL
		);
		INSERT INTO public.metadata ("key",value) VALUES ('MIRAGRATION_VERSION','`+strconv.Itoa(len(migrationFile))+`');`)
		if err != nil {
			log.Println("Postgres Migrate: ", err)
			return err
		}
	} else {
		var meta metadata
		err = db.Query(ctx, "SELECT key,value from metadata", emptyParams, &meta, false)
		if err != nil {
			return err
		}
		migrationVersion, _ = strconv.Atoi(meta.Value)
	}

	err = execMigration(ctx, migrationFile, tx, migrationVersion)
	if err != nil {
		if errRb := tx.Rollback(); errRb != nil {
			log.Println("Postgres Migrate: ", errRb)
			return errRb
		}
		log.Println("Postgres Migrate: ", err)
		return err
	}
	if len(migrationFile) > migrationVersion {
		queryUpdate := fmt.Sprintf("UPDATE metadata set value = %d where key = 'MIRAGRATION_VERSION'", len(migrationFile))
		_, err = tx.Exec(queryUpdate)
	}
	if errCommit := tx.Commit(); errCommit != nil {
		log.Println("Postgres Migrate: ", errCommit)
		return errCommit
	}

	return nil
}

func (db *database) GenerateQueryParams(query string, params map[string]interface{}, searchBy map[string]interface{}) string {
	i := 0
	opr := "WHERE"
	for k := range params {
		if i > 0 {
			opr = "AND"
		}
		query += fmt.Sprintf(" %s %s = :%s", opr, k, k)
		i++
	}

	if searchBy != nil {
		if i > 0 {
			opr = "AND"
		}

		query += fmt.Sprintf(" %s (", opr)
		indexSearch := 0
		for k, v := range searchBy {
			if indexSearch > 0 {
				query += " or "
			}
			query += fmt.Sprintf("lower(%s) like lower('%%%s%%')", k, v)
			indexSearch++
		}
		query += ")"
		i++
	}
	return query
}

func (db *database) WithLimitOffset(query string, limit, offset int) string {
	if limit != 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
	}
	return query
}

func (db *database) WithOrder(query string, orderBy, orderDir string) string {
	if orderBy != "" {
		query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)
	}
	return query
}

func (db *database) Query(ctx context.Context, query string, params, response interface{}, forUpdate bool) error {
	kind, err := findKind(response)
	if err != nil {
		return err
	}

	if forUpdate {
		query += " FOR UPDATE"
	}
	newQuery := query
	var args []interface{}
	if params != nil {
		newQuery, args, err = db.sqlxDB.BindNamed(query, params)
		if err != nil {
			log.Println("Postgres Query: ", err)
			return err
		}
	}

	rows, err := db.sqlxDB.QueryxContext(ctx, newQuery, args...)
	if err != nil {
		log.Println("Postgres Query: ", err)
		return err
	}
	defer rows.Close()
	switch kind {
	case ptrStruct:
		if !rows.Next() {
			return ErrDataNotFound
		}
		return rows.StructScan(response)
	case prtSliceOfStruct:
		slcElem := reflect.ValueOf(response).Elem()
		slcElem.Set(reflect.Zero(slcElem.Type()))
		for rows.Next() {
			n := slcElem.Len()
			slcElem.Set(reflect.Append(slcElem, reflect.New(slcElem.Type().Elem().Elem())))
			if err := rows.StructScan(slcElem.Index(n).Interface()); err != nil {
				return err
			}
		}
		return nil
	case ptrSingleType:
		if !rows.Next() {
			return ErrDataNotFound
		}
		return rows.Scan(response)
	}
	return nil
}

func (db *database) RunInTransaction(ctx context.Context, f func(tctx context.Context) error) error {
	tx, err := db.sqlxDB.Beginx()
	if err != nil {
		log.Println("Postgres RunInTransaction: ", err)
		return ErrCreateTx
	}

	ctx = NewContextTx(ctx, tx)
	if err := f(ctx); err != nil {
		if errRb := tx.Rollback(); errRb != nil {
			log.Println("Postgres RunInTransaction: ", errRb)
			return ErrRollbackTx
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Println("Postgres RunInTransaction: ", err)
		return ErrCommitTx
	}

	return nil
}

func (db *database) Exec(ctx context.Context, stmt string, params interface{}) error {
	if tx, ok := TxFromContext(ctx); ok {
		db.sqlxTX = tx
	}
	newStmt, args, err := db.sqlxTX.BindNamed(stmt, params)
	if err != nil {
		log.Println("Postgres Exec: ", err)
		return err
	}

	if _, err := db.sqlxTX.ExecContext(ctx, newStmt, args...); err != nil {
		log.Println("Postgres Exec: ", err)
		return err
	}
	return nil
}

func (db *database) NamedExec(ctx context.Context, stmt string, params interface{}) error {
	if tx, ok := TxFromContext(ctx); ok {
		db.sqlxTX = tx
	}

	if _, err := db.sqlxTX.NamedExecContext(ctx, stmt, params); err != nil {
		log.Println("Postgres NamedExec: ", err)
		return err
	}
	return nil
}

func execMigration(ctx context.Context, listMigration []string, tx *sql.Tx, migrationVersion int) error {
	log.Printf("Migration Version %d", migrationVersion)
	log.Printf("Current Migration Version %d", len(listMigration))
	i := 1
	for _, data := range listMigration {
		if i > migrationVersion {
			log.Printf("Executing Migration %d:", i)
			_, err := tx.ExecContext(ctx, data)
			if err != nil {
				log.Println("Postgres execMigration: ", err)
				return err
			}
		}
		i++
	}

	return nil
}

func findKind(response interface{}) (int, error) {
	t := reflect.TypeOf(response)
	if t.Kind() != reflect.Ptr {
		log.Printf("invalid format type %T - type should be pointers\n", response)
		return -1, ErrResShouldBePtr
	}

	t = t.Elem()

	kind := ptrStruct
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		if t.Kind() != reflect.Ptr {
			log.Printf("invalid format type %T - type should be pointers\n", response)
			return -1, ErrResShouldBePtr
		}
		t = t.Elem()
		kind = prtSliceOfStruct
	} else if t.Kind() != reflect.Struct {
		kind = ptrSingleType
	}

	return kind, nil
}

func getAndStartConnection() *sqlx.DB {
	onceDB.Do(func() {
		conf := config.GetConfiguration()
		db = sqlx.MustConnect("postgres", conf.PostgresConn)
	})
	return db
}

func NewDatabase() DatabaseInterface {
	db := getAndStartConnection()
	return &database{
		sqlxDB: db,
	}
}
