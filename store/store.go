package store

import (
	"database/sql"
	"fmt"
	"time"

	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/mgutz/dat.v1/sqlx-runner"

	// postgresql driver for database/sql
	_ "github.com/lib/pq"

	"github.com/jllopis/try6/log"
)

// Storer es la interfaz que deben implementar los diferentes backends que realizen
// la persistencia de la aplicación.
type Storer interface {
	Dial(options Options) error
	Status() (int, string)
	Close() error
	Tenanter
	Accounter
	Keyer
	Directer
	Scoper
}

/*

	// Tokens
	LoadToken(kid string) (string, error)
	GetTokenByEmail(email string) (string, error)
	GetTokenByAccountID(uid string) (string, error)
	SaveToken(uid string, tok *string) error
	DeleteToken(kid string) error
	// RBAC
*/

const (
	// DISCONNECTED indicates that there is no connection with the Storer
	DISCONNECTED = iota
	// CONNECTED indicate that the connection with the Storer is up and running
	CONNECTED
)

var (
	// StatusStr is a string representation of the status of the connections with the Storer
	StatusStr = []string{"Disconnected", "Connected"}
)

// DefaultStore is a default Storer implementation over PostgresSQL
type DefaultStore struct {
	C    *runner.DB
	Stat int
}

// Options is a map to hold the database connection options
type Options map[string]interface{}

var _ = (*DefaultStore)(nil)

// NewDefaultStore is a default Storer implementation based upon PostgresSQL
func NewDefaultStore() (*DefaultStore, error) {
	return &DefaultStore{}, nil
}

// Dial perform the connection to the underlying database server
func (d *DefaultStore) Dial(options Options) error {
	if v, ok := options["sslMode"]; !ok || v == "" {
		options["sslMode"] = "disable"
	}
	if v, ok := options["maxIdleConns"]; !ok || v.(int) == 0 {
		options["maxIdleConns"] = 20
	}
	if v, ok := options["maxOpenConns"]; !ok || v.(int) == 0 {
		options["maxOpenConns"] = 50
	}
	ds := fmt.Sprintf("user=%s dbname=%s sslmode=%s password=%s host=%s port=%d", options["user"], options["name"], options["sslMode"], options["password"], options["host"], options["port"])
	log.LogI("connecting to postgresql", "string", ds)
	db, err := sql.Open("postgres", ds)
	if err != nil {
		return err
	}
	// ensures the database can be pinged with an exponential backoff (15 min)
	runner.MustPing(db)

	db.SetMaxIdleConns(options["maxIdleConns"].(int))
	db.SetMaxOpenConns(options["maxOpenConns"].(int))

	// set this to enable interpolation
	dat.EnableInterpolation = true
	// set to check things like sessions closing.
	// Should be disabled in production/release builds.
	dat.Strict = false

	// Log any query over 10ms as warnings. (optional)
	runner.LogQueriesThreshold = 10 * time.Millisecond

	d.C = runner.NewDB(db, "postgres")
	d.Stat = CONNECTED

	return nil
}

// Status return the current status of the underlying database
func (d *DefaultStore) Status() (int, string) {
	return d.Stat, StatusStr[d.Stat]
}

// Close effectively close de database connection
func (d *DefaultStore) Close() error {
	log.LogW("DefaultStore CLOSING", "pkg", "store", "func", "(d *DefaultStore) Close() error", "msg", "closing default store. app will not query anymore")
	return d.C.DB.Close()
}
