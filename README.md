# MiniDB

MiniDB is a simplified, small-scale database system implemented in Go. It's designed to provide basic database functionality and help developers understand core database principles.

## Features

- SQL Query Support:
  - DDL (Data Definition Language):
    - CREATE TABLE - Create new tables with schema
    - DROP TABLE - Remove existing tables
    - SHOW TABLES - List all tables
  - DML (Data Manipulation Language):
    - SELECT - Query data from tables
    - INSERT - Add new records
    - UPDATE - Modify existing records
    - DELETE - Remove records
- In-memory storage engine with schema support
- Concurrent access support with proper locking
- Network communication layer
- SQL query parsing and execution

## Project Structure

```bash
minidb/
├── cmd/
│   └── minidb/
│       └── main.go
├── internal/
│   ├── network/
│   │   └── server.go
│   ├── parser/
│   │   └── parser.go
│   ├── executor/
│   │   └── executor.go
│   └── storage/
│       └── engine.go
├── go.mod
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.20 or later

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yyun543/minidb.git
   ```

2. Change to the project directory:
   ```bash
   cd minidb
   ```

3. Build the project:
   ```bash
   go build ./cmd/minidb
   ```

### Running the Server

To start the MiniDB server, run:

```bash
./minidb
```

The server will start and listen on port 8086 by default.

### Connecting to the Server

You can use any TCP client (like telnet) to connect to the server and send SQL queries:

```bash
telnet localhost 8086
```

## Usage Examples

Once connected, you can send SQL queries to the server. Here are some example queries:

### DDL Operations
```sql
-- Create a new table
CREATE TABLE users (id INT, name VARCHAR, email VARCHAR);
CREATE TABLE users_back (id INT, name VARCHAR, email VARCHAR);

-- Show all tables
SHOW TABLES;

-- Drop a table
DROP TABLE users_back;
```

### DML Operations
```sql
-- Insert data
INSERT INTO users VALUES (1, John, john@example.com);
INSERT INTO users VALUES (2, Harry, harry@example.com);

-- Query data
SELECT * FROM users;
SELECT name, email FROM users;

-- Update data
UPDATE users SET name=Jane WHERE id=1;

-- Delete data
DELETE FROM users WHERE id=1;
```

To exit the client, type `exit` or `quit` and press Enter.

## Limitations

This is a simplified database system and has several limitations:

- Only supports basic SQL operations
- Uses in-memory storage (data is not persistent)
- Limited data types support
- No support for complex queries (e.g., JOIN, GROUP BY)
- No indexing or query optimization
- Basic error handling and security features

## Future Improvements

- Implement persistent storage
- Add support for more complex SQL operations
- Add support for more data types
- Implement indexing and query optimization
- Add transaction support (ACID properties)
- Improve error handling and add logging
- Implement security features (authentication, authorization)
- Add support for configuration options
- Add support for foreign keys and constraints

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
