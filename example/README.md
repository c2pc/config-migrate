Run `go run main.go` and check config.yml in parent directory

```yaml

# Authentication and token settings
auth:

  # Access token time-to-live in minutes
  access_token_ttl: 15

  # Secret key for signing tokens
  key: Jz5ldHc@M5_+U!.pnXLM2YoHy(vKAGCVuY+xqc=HC09IJIRXERHL9MIIIHNgheBm

  # Refresh token time-to-live in minutes
  refresh_token_ttl: 540

# Database connection configuration
db:

  # PostgreSQL DSN connection string
  dsn: postgres:postgres@8.8.8.8:5432/postgres?sslmode=disable

  # Maximum number of idle database connections
  max_idle_conn: 10

  # Maximum number of open database connections
  max_open_conn: 100
force: false

# HTTP server configuration
http:

  # The IP address the HTTP server will bind to
  host: 8.8.8.8

  # The port the HTTP server will listen on
  port: 8052

# Logging configuration
log:

  # Whether to compress rotated log files
  compress: false

  # Logging level (e.g., debug, info, production)
  debug: production

  # Directory where log files will be stored
  dir: /var/log/Migrate

  # Number of days to keep old log files
  max_age: 28

  # Number of rotated log files to keep
  max_backups: 5

  # Maximum size (in MB) of a log file before it gets rotated
  max_size: 20

# Redis configuration
redis:

  # Redis server address list
  address:
    - localhost:6379

  # Enable Redis cluster mode
  cluster_mode: false

  # Redis database index
  db: 1

  # Maximum number of retry attempts
  # for failed Redis commands
  max_retry: 10

  # Redis password (optional)
  password:

  # Redis username (optional, for ACL-enabled Redis)
  username:

# Salt used for password hashing
salt:

  # Salt value used to hash user passwords
  password: 6.E_6kXr7l-L)Q)BKgmwAkkHBcD4ROuU
version: 3
```