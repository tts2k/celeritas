//go:build integration

// run tests with this command: go test . --tags integration --count=1
package data

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host = "localhost"
	user = "postgres"
	password = "secret"
	dbName = "celeritas_test"
	port = "5435"
	dsn = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var dummyUser = User{
	FirstName: "Some",
	LastName: "Guy",
	Email: "me@here.com",
	Active: 1,
	Password: "password",
}

var models Models
var testDB *sql.DB
var resource *dockertest.Resource
var pool *dockertest.Pool

func TestMain(m *testing.M) {
	os.Setenv("DATABASE_TYPE", "postgres")

	p, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connectd to docker: %s", err)
	}

	pool = p

	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag: "13.4",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		if resource != nil {
			_ = pool.Purge(resource)
		}
		log.Fatalf("could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("cound not connect to docker: %s", err)
	}

	err = createTable(testDB)
	if err != nil {
		log.Fatalf("error creating tables: %s", err)
	}

	models = New(testDB)

	code := m.Run()


	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)

}

func createTable(db *sql.DB) error {
	// read sql file
	data, err := os.ReadFile("./auth_tables.sql")
	if (err != nil) {
		panic (err)
	}
	stmt := string(data)

	_, err = db.Exec(stmt)
	if (err != nil) {
		return err;
	}
	return nil
}

func TestUser_Table(t *testing.T) {
	s := models.Users.Table()
	if s != "users" {
		t.Error("wrong table name returned: ", s)
	}
}

func TestUser_Insert(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("failed to isnert user: ", err)
	}

	if id == 0 {
		t.Error("0 returned as id after insert")
	}
}

func TestUser_Get(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	if u.ID == 0 {
		t.Error("id of returned user is 0", err)
	}
}

func TestUser_GetAll(t *testing.T) {
	_, err := models.Users.GetAll()
	if err != nil {
		t.Error("failed to get user: ", err)
	}
}

func TestUser_GetByEmail(t *testing.T) {
	u, err := models.Users.GetByEmail("me@here.com")
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	if u.ID == 0 {
		t.Error("id of returned user is 0", err)
	}
}

func TestUser_Update(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	u.LastName = "Smith"
	err = u.Update(*u)
	if err != nil {
		t.Error("failed to get user", err)
	}

	u, err = models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	if u.LastName != "Smith" {
		t.Error("last name not updated in database")
	}
}

func TestUser_PasswordMatches(t *testing.T) {
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user: ", err)
	}

	matches, err := u.PasswordMatches("password")
	if err != nil {
		t.Error("error checking match: " , err)
	}

	if !matches {
		t.Error("password does not match when it should")
	}

	matches, err = u.PasswordMatches("123")
	if err != nil {
		t.Error("error checking match: " , err)
	}

	if matches {
		t.Error("password matches when it should not")
	}
}

func TestUser_ResetPassword(t *testing.T) {
	err := models.Users.ResetPassword(1, "new_password")
	if err != nil {
		t.Error("error resetting password: ", err)
	}

	err = models.Users.ResetPassword(2, "new_password")
	if err == nil {
		t.Error("did not get an error when trying to reset password")
	}
}

func TestUser_Delete(t *testing.T) {
	err := models.Users.Delete(1)
	if err != nil {
		t.Error("failed to delete user: ", err)
	}

	_, err = models.Users.Get(1)
	if err == nil {
		t.Error("retrieved user who was suppose to be deleted")
	}
}

func TestToken_Table(t *testing.T) {
	s := models.Tokens.Table()
	if s != "tokens" {
		t.Error("wrong table name returned for tokens")
	}
}

func TestToken_GenerateToken(t *testing.T) {
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("error inserting user: ", err)
	}

	_, err = models.Tokens.GenerateToken(id, time.Hour * 24 * 365)
	if err != nil {
		t.Error("error generating token: ", err)
	}
}

func TestToken_Insert(t *testing.T) {
	
}
