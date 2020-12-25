package storage

import "context"

type Interface interface {
	Migrate(forceMigrate bool) error
	Exec(ctx context.Context, stmt string, params interface{}) error
	Query(ctx context.Context, query string, params, response interface{}, forUpdate bool) error
	RunInTransaction(ctx context.Context, f func(tctx context.Context) error) error
	GenerateQueryParams(query string, params map[string]interface{}, searchBy map[string]interface{}) string
	WithLimitOffset(query string, limit, offset int) string
	WithOrder(query string, orderBy, orderDir string) string
	NamedExec(ctx context.Context, stmt string, params interface{}) error
}
