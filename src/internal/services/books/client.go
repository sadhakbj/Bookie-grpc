// Package books is the BFF client / http server exposed to external clients.
package books

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

	conn, err := grpc.Dial("localhost:8020", opts...)
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
