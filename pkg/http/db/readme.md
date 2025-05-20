# Database

This package helps us to work with database.

## Configuration

`Config` struct provides a template for database configuration, it also has annotation for loading variables
from environment.

Some environment you will interest to work with this:

- `DB_DRIVER`: core supported drivers: `sqlite3` & `postgres`, you also add more easily by
import the gorm adaptable driver.
- `DB_DSN`: data source name to connect
- `DB_MAX_OPEN_CONNS`: max open connections, default 25
- `DB_MAX_IDLE_CONNS`: max idle connections, default 25
- `DB_CONN_MAX_LIFETIME`: max idle connections lifetime (you know,
MySql will close any connection that has unused more than 8 hours)


## Get started

```go
package main

import (
  "github.com/praslar/cloud0"
)

func main() {
  config := &db.Config{Driver: "sqlite3", DSN: ":memory:"}
  db.MustOpenDefault(config)
  defer db.CloseDB()
  // work with db via db.GetDB()

  user := User{Name: "Kingsley", Age: 30}
  db.GetDB().Create(&user)

}
```

