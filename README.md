# budget

The same binary runs as the API and, with `CHECK_AND_QUIT=true`, as the budget check job; both need the
MongoDB settings below.

## MongoDB configuration

Set in `config.json`, overridden by environment variables:

| Variable                  | Default                     | Meaning                                                   |
|---------------------------|-----------------------------|-----------------------------------------------------------|
| `MONGO_URL`               | `mongodb://localhost:27017` | connection string including the scheme, used unchanged    |
| `MONGO_USER`              | empty                       | user name; empty means no authentication                  |
| `MONGO_PASSWORD`          | empty                       | password; required when `MONGO_USER` is set, never logged |
| `MONGO_AUTH_SOURCE`       | `admin`                     | database the user is defined in                           |
| `MONGO_DATABASE`          | `budget`                    | database of this service; must not be empty               |
| `MONGO_BUDGET_COLLECTION` | `budgets`                   | collection for budgets                                    |
| `MONGO_TIMEOUT`           | `10s`                       | timeout per query                                         |

Keep credentials out of `MONGO_URL` and use `MONGO_USER`/`MONGO_PASSWORD`.
When `MONGO_USER` is set, `MONGO_USER`, `MONGO_PASSWORD` and `MONGO_AUTH_SOURCE` (default `admin`)
replace the user, password, `authSource` and `authMechanism` given in `MONGO_URL`; the mechanism is then
negotiated with the server. At startup the service lists the collections of `MONGO_DATABASE` and exits
if MongoDB is unreachable within 10 seconds or the credentials are wrong, missing or lack rights on that database.

## Tests

Run `go test -short ./...`. Without `-short`, `TestNewAuthenticates` also runs against a throwaway MongoDB
with access control when these are set; it creates and drops its own users and databases:

| Variable                   | Meaning                               |
|----------------------------|---------------------------------------|
| `MONGO_AUTH_TEST_URL`      | connection string without credentials |
| `MONGO_AUTH_TEST_USER`     | root user of the throwaway server     |
| `MONGO_AUTH_TEST_PASSWORD` | its password                          |

    docker run -d --rm --name budget-auth-test -p 127.0.0.1:27018:27017 -e MONGO_INITDB_ROOT_USERNAME=root -e MONGO_INITDB_ROOT_PASSWORD=rootpw mongo:8.2
    MONGO_AUTH_TEST_URL=mongodb://127.0.0.1:27018 MONGO_AUTH_TEST_USER=root MONGO_AUTH_TEST_PASSWORD=rootpw go test ./pkg/database/
    docker stop budget-auth-test
