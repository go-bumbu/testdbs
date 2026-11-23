package testdbs

const (
	DBTypeMariadb = "mariadb"
	// defaultMariadbImage runs when no image is given: the current MariaDB
	// LTS, the one Debian stable ships.
	defaultMariadbImage = "mariadb:11.8"
)

// NewMariadb returns a MariaDB target, the one InitDBS runs. An empty image
// runs the default mariadb:11.8; give another, e.g. "mariadb:10.11", to run
// that. The image must start like the official mariadb image: it accepts the
// MYSQL_* env vars, takes server flags as the container command (set to turn
// durability off), and logs "mariadbd: ready for connections" with
// "port: 3306" once the real server is up.
//
// MariaDB shares MySQL's protocol and GORM driver, so DbType() is the only way
// to tell the two apart: it returns DBTypeMariadb.
func NewMariadb(image string) TargetDb {
	return &testDBMysql{flavor: flavorMariadb, image: image}
}
