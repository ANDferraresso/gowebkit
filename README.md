# gowebkit
A simple and modular toolkit for Go, designed to help you build web applications with HTTP more easily.

```sh
go get github.com/ANDferraresso/gowebkit
```

Requires Go 1.24 or later.

## Packages

| Package     | What it does |
|-------------|--------------|
| `orm`       | Query builders (SELECT, INSERT, UPDATE, DELETE) for MySQL/MariaDB and PostgreSQL, plus raw queries |
| `session`   | Server-side sessions stored in MySQL, PostgreSQL, Redis or Valkey |
| `cookie`    | Plain, base64, signed (HMAC) and encrypted cookies; signed session cookie |
| `csrf`      | CSRF protection with the double-submit cookie pattern |
| `jwtlib`    | HS256 JWT generation and verification |
| `form`      | Form definitions and validation of HTTP requests |
| `validator` | String validators (dates, numbers, e-mail, URL, passwords, ...) |
| `logg`      | Asynchronous JSON logger with one file per level and daily rotation |
| `cast`      | Conversions from `interface{}` to string/bool |
| `input`     | Input normalization helpers |

## orm

Every builder takes a `*sql.DB` and a `debug` flag (with `debug` true, error
messages include the DBMS error and the query). `Build()` returns an `orm.Res`:

```go
type Res struct {
	Err  bool
	Msg  string
	Data []map[string]interface{}
}
```

SELECT values are returned as `string` (or `nil` for NULL).

```go
res := orm.SelectQuery(db, false).
	Select("id", "c.name", "email AS mail").
	From("clienti AS c").
	Where("(", "id = ?", 1, "OR", "c.name IS NULL", ")", "AND", "email LIKE ?", "%@example.com").
	Order([][2]string{{"id", "DESC"}}).
	Limit(10).Offset(20).
	Build()

res = orm.InsertQuery(db, false).
	Insert("clienti").
	Columns("name", "email").
	Values("Andrea", nil).
	Build() // res.Data[0]["lastInsertId"]

res = orm.UpdateQuery(db, false).
	Update("clienti").
	Set("name = ?", "Andrea", "email = ?", "andrea@example.com").
	Where("id = ?", 1).
	Build() // res.Data[0]["rowsAffected"]

res = orm.DeleteQuery(db, false).
	Delete("clienti").
	Where("id = ?", 1).
	Build()
```

### WHERE syntax

`Where()` takes a sequence of tokens. Table and column names are validated and
quoted, values are always passed as bound parameters:

- a condition: `"col = ?"` followed by its value, `"col IS NULL"`,
  `"col IS NOT NULL"`, or a comparison between two columns `"a >= b"`;
- operators: `<`, `=`, `>`, `<>`, `<=`, `>=`, `LIKE`, `ILIKE`;
- logical operators between conditions: `AND`, `OR`, `NOT`, `AND NOT`, `OR NOT`;
- groups: `"("`, `"(("`, `")"`, `"))"`, ... (they must be balanced).

`table.column` is accepted in SELECT queries. `ILIKE` is a case-insensitive
`LIKE`: it is sent as is to PostgreSQL and becomes `LIKE` on MySQL, where `LIKE`
is already case-insensitive with the default `_ci` collations.

### Safety checks

- `UPDATE` and `DELETE` without `WHERE` are refused. Call `AllRows()` to really
  affect every row of the table.
- `Order()` only accepts `ASC`, `DESC` or an empty direction (= `ASC`).
- ``SetRaw("`counter` = `counter` + 1")`` (UPDATE) inserts the text as is:
  never pass user data to it.

### PostgreSQL

The builders generate MySQL syntax by default. To use PostgreSQL, set the
dialect once at startup or on a single builder:

```go
orm.DefaultDialect = orm.Postgres // whole program

res := orm.SelectQuery(db, false).Dialect(orm.Postgres)./* ... */Build() // one query
```

With `Postgres`, identifiers are quoted with `"double quotes"` and placeholders
become `$1, $2, ...`. Keep in mind:

- quoted identifiers are case-sensitive in PostgreSQL, and names created without
  quotes are lowercase: write `Select("id")`, not `Select("ID")`;
- PostgreSQL drivers don't support `LastInsertId()`. Use `Returning()`:

  ```go
  res := orm.InsertQuery(db, false).
  	Dialect(orm.Postgres).
  	Insert("clienti").
  	Columns("name").
  	Values("Andrea").
  	Returning("id").
  	Build()
  id := res.Data[0]["lastInsertId"] // int64, or string for non-integer keys (UUID)
  ```

- `SetRaw()` text is not converted: write it in PostgreSQL syntax;
- booleans are returned as `"t"`/`"true"`: `cast.ValToBool` understands both.

The package imports no PostgreSQL driver: register one in your application, e.g.
`_ "github.com/jackc/pgx/v5/stdlib"` (`sql.Open("pgx", dsn)`).

### Raw queries

```go
res := orm.RawQuery(db, false, "SELECT id, name FROM clienti WHERE id > ?", 10)
res = orm.RawExec(db, false, "UPDATE clienti SET visits = visits + 1 WHERE id = ?", 1)
```

The SQL is sent to the driver unchanged, so use the placeholders of your
database (`$1, $2, ...` with PostgreSQL).

## session

All stores implement `session.Store`:

```go
var store session.Store = &session.MysqlStore{}
if err := store.InitStore(db, "sessions"); err != nil { /* ... */ }
fmt.Println(store.CreateStore()) // CREATE TABLE statement for the store

sid, err := store.Init(time.Now().Add(2 * time.Hour)) // zero time = no expiry
err = store.Set(sid, "userID", 42)
v, err := store.Get(sid, "userID")
newSid, err := store.Refresh(sid) // new sid, same data: call it after login
err = store.Delete(sid)
```

| Store           | `InitStore` client | Second argument |
|-----------------|--------------------|-----------------|
| `MysqlStore`    | `*sql.DB` (MySQL/MariaDB) | table name |
| `PostgresStore` | `*sql.DB` (PostgreSQL driver of your choice) | table name |
| `RedisStore`    | `*redis.Client` (go-redis v9) | key prefix |
| `ValkeyStore`   | `*redis.Client` connected to Valkey | key prefix |

Concurrent changes to the same session are safe: the SQL stores lock the row
(`SELECT ... FOR UPDATE`), the Redis/Valkey stores use optimistic transactions
and may return `ErrSessionConflict` under heavy contention.

Redis and Valkey expire sessions by themselves. The SQL stores delete an
expired session only when it is read, so call `DeleteExpired()` periodically to
remove abandoned ones:

```go
go func() {
	for range time.Tick(time.Hour) {
		if c, ok := store.(session.ExpiredCleaner); ok {
			if _, err := c.DeleteExpired(); err != nil {
				log.Println(err)
			}
		}
	}
}()
```

## cookie

```go
// Signed session cookie (HttpOnly, Secure, SameSite=Lax); key in hex, duration in hours.
err := cookie.WriteSessionCookie(w, hexKey, sid, 3)
sid, err := cookie.ReadSessionCookie(r, hexKey)
```

`WriteSigned`/`ReadSigned` and `WriteEncrypted`/`ReadEncrypted` work with any
`http.Cookie`.

## csrf

```go
token, err := csrf.GenerateCSRFToken()
err = csrf.SetCSRFTokenCookie(w, token)
// put token in the form field "_csrf" or in the "X-CSRF-Token" header

if !csrf.ValidateRequest(r) { // constant-time comparison with the cookie
	http.Error(w, "invalid CSRF token", http.StatusForbidden)
	return
}
```

## jwtlib

```go
token, err := jwtlib.GenerateJWT(secret, 24, "my-app", userID) // valid 24 hours
claims, err := jwtlib.VerifyJWT(secret, token)
```

Only HS256 is accepted, the secret can't be empty and tokens without `exp` are
rejected.

## logg

```go
l, err := logg.NewLogger(logg.INFO, "./logs")
defer l.Close()
l.Info("user logged in", logg.Field("userID", 42))
```

Each level is written as JSON lines to `<logDir>/<level>_<YYYY-MM-DD>.log` and
to stdout (stderr for errors). Files rotate when the date changes.

## input

`NormalizeSingleLine` trims surrounding whitespace. `Sanitize` and
`ContainsXSS` are deprecated: they are blacklists and don't protect against
XSS or SQL injection. Validate input with the `validator` package, use bound
parameters (the `orm` builders) and escape output with `html/template`.

## Development

```sh
./check.sh # go fmt, go vet, go build, go test, golangci-lint
```

## License

See [LICENSE](LICENSE).
