// Package books is the BFF client / http server exposed to external clients.
package books

import (
	"context"
	"log/slog"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
)

// Book definiation
type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Price       int    `json:"price"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

// GRPCClient is the structure.
type GRPCClient struct {
	conn   *grpc.ClientConn
	client bookiePb.BookieClient
}

// NewGRPCClient creates new instance of grpc client
func NewGRPCClient() (*GRPCClient, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Get gRPC server address from environment variable, default to localhost:8020
	grpcServerAddr := "localhost:8020"
	if addr := os.Getenv("GRPC_SERVER_ADDR"); addr != "" {
		grpcServerAddr = addr
	}

	slog.Info("Connecting to gRPC server", slog.String("address", grpcServerAddr))
	conn, err := grpc.NewClient(grpcServerAddr, opts...)
	if err != nil {
		return nil, err
	}

	client := bookiePb.NewBookieClient(conn)

	return &GRPCClient{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the grpc connection
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// IsHealthy reports whether the gRPC server is serving, using the standard
// gRPC health checking protocol. Used by the BFF readiness probe so that
// Kubernetes only routes traffic when the upstream is actually reachable.
func (c *GRPCClient) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := healthpb.NewHealthClient(c.conn).Check(ctx, &healthpb.HealthCheckRequest{})
	return err == nil && resp.GetStatus() == healthpb.HealthCheckResponse_SERVING
}

// GetBooks returns all the books from grpc book service
func (c *GRPCClient) GetBooks() ([]*Book, error) {
	res, err := c.client.ListBooks(context.Background(), &bookiePb.ListBookRequest{PerPage: 10})
	if err != nil {
		return nil, err
	}

	// Convert the response to []*Book
	var bks []*Book
	for _, book := range res.GetBooks() {
		book1 := &Book{
			ID:          book.GetId(),
			Title:       book.GetTitle(),
			Description: book.GetDescription(),
			Price:       int(book.GetPrice()),
			Author:      book.GetAuthor(),
		}

		bks = append(bks, book1)
	}

	return bks, nil
}

// GetByID returns the resource with provided id
func (c *GRPCClient) GetByID(id string) (*Book, error) {
	res, err := c.client.GetByID(context.Background(), &bookiePb.GetByIDRequest{Id: id})
	if err != nil {
		return nil, err
	}

	book := &Book{
		ID:          res.GetBook().GetId(),
		Title:       res.GetBook().GetTitle(),
		Description: res.GetBook().GetDescription(),
		Price:       int(res.GetBook().GetPrice()),
		Author:      res.GetBook().GetAuthor(),
	}

	return book, nil
}
