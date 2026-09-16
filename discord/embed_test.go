package discord

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestEmbedBuilder_Valid(t *testing.T) {
	now := time.Now()
	embed, err := NewEmbedBuilder().
		SetTitle("Test Embed").
		SetDescription("This is a valid test embed").
		SetURL("https://discord.com").
		SetColor(0x5865F2).
		SetTimestamp(now).
		SetImage("https://example.com/image.png").
		SetThumbnail("https://example.com/thumb.png").
		SetAuthor("Author", "https://example.com", "https://example.com/icon.png").
		SetFooter("Footer text", "https://example.com/footer.png").
		AddField("Field 1", "Value 1", true).
		AddField("Field 2", "Value 2", true).
		Build()

	if err != nil {
		t.Fatalf("expected valid embed build, got error: %v", err)
	}

	if embed.Title != "Test Embed" {
		t.Errorf("expected title 'Test Embed', got %s", embed.Title)
	}
	if embed.URL != "https://discord.com" {
		t.Errorf("expected URL 'https://discord.com', got %s", embed.URL)
	}
	if embed.Color != 0x5865F2 {
		t.Errorf("expected color 0x5865F2, got %d", embed.Color)
	}
	if embed.Image.URL != "https://example.com/image.png" {
		t.Errorf("expected image URL, got %s", embed.Image.URL)
	}
	if embed.Thumbnail.URL != "https://example.com/thumb.png" {
		t.Errorf("expected thumbnail URL, got %s", embed.Thumbnail.URL)
	}
	if embed.Author.Name != "Author" {
		t.Errorf("expected author name 'Author', got %s", embed.Author.Name)
	}
	if embed.Footer.Text != "Footer text" || embed.Footer.IconURL != "https://example.com/footer.png" {
		t.Errorf("expected footer text and icon to match")
	}
	if len(embed.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(embed.Fields))
	}
}

func TestEmbedBuilder_ValidationErrors(t *testing.T) {
	// Title too long
	_, err := NewEmbedBuilder().
		SetTitle(strings.Repeat("a", MaxEmbedTitleLength+1)).
		Build()
	if !errors.Is(err, ErrEmbedTitleTooLong) {
		t.Errorf("expected ErrEmbedTitleTooLong, got %v", err)
	}

	// Description too long
	_, err = NewEmbedBuilder().
		SetDescription(strings.Repeat("b", MaxEmbedDescriptionLength+1)).
		Build()
	if !errors.Is(err, ErrEmbedDescriptionTooLong) {
		t.Errorf("expected ErrEmbedDescriptionTooLong, got %v", err)
	}

	// Too many fields
	builder := NewEmbedBuilder()
	for i := 0; i < MaxEmbedFields+1; i++ {
		builder.AddField("field", "val", false)
	}
	_, err = builder.Build()
	if !errors.Is(err, ErrEmbedTooManyFields) {
		t.Errorf("expected ErrEmbedTooManyFields, got %v", err)
	}

	// Field name too long
	_, err = NewEmbedBuilder().
		AddField(strings.Repeat("f", MaxEmbedFieldNameLength+1), "val", false).
		Build()
	if !errors.Is(err, ErrEmbedFieldNameTooLong) {
		t.Errorf("expected ErrEmbedFieldNameTooLong, got %v", err)
	}

	// Field value too long
	_, err = NewEmbedBuilder().
		AddField("name", strings.Repeat("v", MaxEmbedFieldValueLength+1), false).
		Build()
	if !errors.Is(err, ErrEmbedFieldValueTooLong) {
		t.Errorf("expected ErrEmbedFieldValueTooLong, got %v", err)
	}

	// Footer too long
	_, err = NewEmbedBuilder().
		SetFooter(strings.Repeat("f", MaxEmbedFooterTextLength+1)).
		Build()
	if !errors.Is(err, ErrEmbedFooterTooLong) {
		t.Errorf("expected ErrEmbedFooterTooLong, got %v", err)
	}

	// Author name too long
	_, err = NewEmbedBuilder().
		SetAuthor(strings.Repeat("a", MaxEmbedAuthorNameLength+1), "", "").
		Build()
	if !errors.Is(err, ErrEmbedAuthorNameTooLong) {
		t.Errorf("expected ErrEmbedAuthorNameTooLong, got %v", err)
	}

	// Total characters exceed 6000
	totalBuilder := NewEmbedBuilder().
		SetDescription(strings.Repeat("d", 4000))
	for i := 0; i < 3; i++ {
		totalBuilder.AddField("field", strings.Repeat("x", 800), false)
	}
	_, err = totalBuilder.Build()
	if !errors.Is(err, ErrEmbedTotalTooLong) {
		t.Errorf("expected ErrEmbedTotalTooLong, got %v", err)
	}
}
