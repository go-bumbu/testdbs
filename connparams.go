package testdbs

// ConnParams holds what a caller needs to open its own connection to a test
// database, e.g. to test code that builds its own DSN or gorm.Config.
type ConnParams struct {
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
	// SslMode is libpq's sslmode; "disable" for Postgres. Empty for MySQL,
	// which has no such setting.
	SslMode string
}

// ConnParamsProvider is implemented by the targets that run a database server
// (Postgres and MySQL), not by SQLite. Type-assert a TargetDb from DBs() to
// get it:
//
//	if p, ok := dbt.(testdbs.ConnParamsProvider); ok {
//		params := p.ConnParamsDbName(t.Name())
//	}
//
// It is a separate interface so that custom TargetDb implementations passed
// to InitCustomDbs do not have to implement it.
type ConnParamsProvider interface {
	// ConnParams returns the parameters of the default database, the one
	// Conn() uses.
	ConnParams() ConnParams
	// ConnParamsDbName returns the parameters of the database ConnDbName(name)
	// uses, creating it if needed.
	ConnParamsDbName(name string) ConnParams
}

// serverDbName returns the name the server knows the database ConnDbName(name)
// uses by. ConnDbName creates databases under the normalized name, but the
// default one is created by the container env as defaultDbName, and the
// server keeps its case.
func serverDbName(name string) string {
	name = normalizeDbName(name)
	if name == normalizeDbName(defaultDbName) {
		return defaultDbName
	}
	return name
}
