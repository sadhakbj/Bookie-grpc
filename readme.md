# 📚 Learning gRPC with Go - Simple Book Service

A **learning project** to understand gRPC fundamentals in Go. This simple book management service demonstrates core gRPC concepts, Protocol Buffers, and how to bridge gRPC with HTTP REST APIs.

> 🎓 **Educational Purpose**: This is a learning repository to explore gRPC concepts, not a production application.

## 🎯 What You'll Learn

- **gRPC Basics**: How to create and consume gRPC services
- **Protocol Buffers**: Defining APIs with `.proto` files
- **Go gRPC**: Server and client implementation in Go
- **HTTP Gateway**: Converting gRPC to REST API
- **Graceful Shutdown**: Proper service lifecycle management
- **Code Quality**: Linting and best practices

## 🛠️ Technologies Used

- **Go 1.24** - Programming language
- **gRPC** - Remote Procedure Call framework
- **Protocol Buffers** - API definition language
- **Standard HTTP** - REST API gateway
- **Structured Logging** - JSON logging with `slog`
- **golangci-lint** - Code quality tools

## 📋 Prerequisites

Make sure you have these installed:

- **Go 1.24+**: [Download here](https://golang.org/dl/)
- **protoc** (Protocol Buffers compiler): [Install guide](https://grpc.io/docs/protoc-installation/)

### Install Required Tools

```bash
# Set up Go paths (add to your shell profile)
export GOPATH=~/go
export PATH=$PATH:/$GOPATH/bin

# Install Protocol Buffer plugins for Go
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## � Quick Start

1. **Clone and setup**:
   ```bash
   git clone https://github.com/sadhakbj/bookie-grpc.git
   cd bookie-grpc
   go mod download
   ```

2. **Generate protobuf code**:
   ```bash
   make generate
   ```

3. **Build everything**:
   ```bash
   make build
   ```

4. **Run the gRPC server** (Terminal 1):
   ```bash
   make serve
   # Starts gRPC server on :8020
   ```

5. **Run the HTTP gateway** (Terminal 2):
   ```bash
   make client
   # Starts HTTP server on :8080
   ```

## 🔍 Understanding the Architecture

```
┌─────────────────┐    gRPC     ┌──────────────────┐
│   HTTP Client   │◄────────────│   gRPC Server    │
│   (Port 8080)   │   :8020     │   (Port 8020)    │
│                 │             │                  │
│ • REST API      │             │ • Book Service   │
│ • JSON Response │             │ • In-memory DB   │
│ • HTTP Gateway  │             │ • Protocol Buf   │
└─────────────────┘             └──────────────────┘
```

**Learning Flow**:
1. **gRPC Server**: Implements the actual book service
2. **HTTP Client**: Acts as a gateway, converting HTTP → gRPC
3. **Protocol Buffers**: Define the contract between services

## 📖 API Examples

### Test the HTTP Gateway

**List all books**:
```bash
curl http://localhost:8080/books
```

**Get specific book**:
```bash
curl http://localhost:8080/books/1234
```

### Sample Response
```json
{
  "success": true,
  "message": "Successfully fetched books",
  "data": [
    {
      "id": "1234",
      "title": "Harry Potter",
      "author": "JK Rowling", 
      "price": 120,
      "description": "a lovely book"
    }
  ]
}
```

## 📁 Project Structure (Learning Guide)

```
bookie-grpc/
├── protos/
│   ├── book.proto              # 📝 API definition (start here!)
│   └── bookie/                 # Generated Go code (don't edit)
├── src/cmd/
│   ├── server/main.go          # 🖥️ gRPC server implementation
│   └── client/main.go          # 🌐 HTTP gateway server
├── src/internal/
│   ├── services/books/         # 📞 gRPC client logic
│   ├── client/controllers/     # 🎮 HTTP handlers
│   └── utils/                  # 🔧 Helper functions
└── makefile                    # 🔨 Build commands
```

**Learning Path**:
1. **Start with `protos/book.proto`** - understand the API
2. **Check `src/cmd/server/main.go`** - see gRPC server implementation
3. **Look at `src/cmd/client/main.go`** - HTTP gateway pattern
4. **Explore `src/internal/services/books/`** - gRPC client usage

## 🧰 Available Commands

| Command | Purpose |
|---------|---------|
| `make generate` | Generate Go code from `.proto` files |
| `make build` | Compile binaries |
| `make serve` | Run gRPC server |
| `make client` | Run HTTP gateway |
| `make lint` | Check code quality |
| `make clean` | Remove generated files |

## 🎓 Learning Exercises

Try these to deepen your understanding:

1. **Add a new RPC method** to create books
2. **Modify the proto file** and regenerate code
3. **Add validation** to the gRPC service  
4. **Implement error handling** for different scenarios
5. **Add logging** to trace requests

## 🔧 Development Tips

**Code Quality**:
```bash
make lint          # Check your code
```

**Testing**:
```bash
go test ./...      # Run tests (add some!)
```

**Graceful Shutdown**:
Both servers support `Ctrl+C` for clean shutdown - great for learning proper service lifecycle!

## 🤝 Contributing to Learning

Found something unclear? Want to add examples?
1. Fork the repo
2. Make your changes
3. Submit a Pull Request

## 📚 Further Learning Resources

- [gRPC Go Tutorial](https://grpc.io/docs/languages/go/quickstart/)
- [Protocol Buffers Guide](https://developers.google.com/protocol-buffers/docs/gotutorial)
- [Go gRPC Examples](https://github.com/grpc/grpc-go/tree/master/examples)

## ❓ Common Issues

**Port already in use?**
```bash
# Kill processes on ports 8020/8080
lsof -ti:8020 | xargs kill
lsof -ti:8080 | xargs kill
```

**Protoc not found?**
- Install Protocol Buffers compiler first
- Make sure `protoc-gen-go` tools are in your PATH

---

🎉 **Happy Learning!** This project is designed to be your hands-on introduction to gRPC in Go.
