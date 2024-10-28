# MiniDB

MiniDB is a simplified, small-scale database system implemented in Go. It's designed to provide basic database functionality and help developers understand core database principles.

## Features

- Basic SQL query support (SELECT, INSERT, UPDATE, DELETE)
- In-memory storage engine
- Simple network communication layer
- Query parsing and execution
- Concurrent access support

## Project Structure

```bash

minidb/
├── cmd/
│ └── minidb/
│ └── main.go
├── internal/
│ ├── network/
│ │ └── server.go
│ ├── parser/
│ │ └── parser.go
│ ├── executor/
│ │ └── executor.go
│ └── storage/
│ └── engine.go
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

## Usage

Once connected, you can send SQL queries to the server. Here are some example queries:

```sql
INSERT INTO users VALUES (1, Yason, Lee);
SELECT * FROM users;
UPDATE users SET col2=Jane WHERE col1=1;
DELETE FROM users WHERE col1=1;
```


To exit the client, type `exit` and press Enter.

## Limitations

This is a simplified database system and has several limitations:

- Only supports basic SQL operations
- Uses in-memory storage (data is not persistent)
- Limited error handling and security features
- No support for complex queries (e.g., JOIN, GROUP BY)
- No indexing or query optimization

## Future Improvements

- Implement persistent storage
- Add support for more complex SQL operations
- Implement indexing and query optimization
- Add transaction support
- Improve error handling and add logging
- Implement security features (authentication, authorization)
- Add support for configuration options

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

