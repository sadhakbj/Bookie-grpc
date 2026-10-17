// Package bookie additional files are used to generate mocks for this package, see the make mocks rule.
package bookie

//go:generate mockgen -source=book_grpc.pb.go -package=bookie -destination=mock_bookie_client.go
