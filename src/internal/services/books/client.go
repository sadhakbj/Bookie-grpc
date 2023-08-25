package books

import (
	"context"

	bookiePb "github.com/sadhakbj/bookie-grpc/protos/bookie"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Price       int    `json:"price"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

type GRPCClient struct {
	conn   *grpc.ClientConn
	client bookiePb.BookieClient
}

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

func (c *GRPCClient) Close() error {
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}

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
