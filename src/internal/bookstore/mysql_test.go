package bookstore

import (
	"testing"
	"uuid"

	"github.com/sadhakbj/bookie-grpc/src/internal/bookstore/sqlcgen"
)

const (
	seededHarryPotterID        = "0199e26a-2951-7000-8000-000000000001"
	seededNineteenEightyFourID = "0199e26a-2951-7000-8000-000000000002"
)

func TestBookFromRowMapsProtoFields(t *testing.T) {
	seedID := uuid.MustParse(seededHarryPotterID)
	row := sqlcgen.Book{
		ID:          seedID[:],
		Title:       "Harry Potter and the Philosopher's Stone",
		Description: "The boy who lived discovers he is a wizard.",
		Author:      "J.K. Rowling",
		Price:       150,
	}

	got := bookFromRow(row)

	if got.GetId() != seededHarryPotterID {
		t.Errorf("id = %q, want canonical seed id %q", got.GetId(), seededHarryPotterID)
	}
	if got.GetTitle() != row.Title || got.GetDescription() != row.Description ||
		got.GetAuthor() != row.Author || got.GetPrice() != row.Price {
		t.Errorf("book = %+v, want fields from %+v", got, row)
	}
}

func TestParseUUIDBytesRoundTrips(t *testing.T) {
	raw, err := parseUUIDBytes(seededNineteenEightyFourID)
	if err != nil {
		t.Fatalf("parseUUIDBytes() error = %v", err)
	}
	if got := uuidString(raw); got != seededNineteenEightyFourID {
		t.Errorf("round trip = %q, want %q", got, seededNineteenEightyFourID)
	}
}

func TestParseUUIDBytesRejectsMalformedID(t *testing.T) {
	if _, err := parseUUIDBytes("not-a-uuid"); err == nil {
		t.Fatal("parseUUIDBytes() accepted a malformed id")
	}
}

func TestUUIDStringRejectsWrongLength(t *testing.T) {
	if got := uuidString([]byte{1, 2, 3}); got != "" {
		t.Errorf("uuidString() = %q, want empty string", got)
	}
}
