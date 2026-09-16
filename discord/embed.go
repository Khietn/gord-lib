package discord

import (
	"errors"
	"fmt"
	"time"
)

// Embed Limits according to Discord API documentation.
const (
	MaxEmbedTitleLength       = 256
	MaxEmbedDescriptionLength = 4096
	MaxEmbedFields            = 25
	MaxEmbedFieldNameLength   = 256
	MaxEmbedFieldValueLength  = 1024
	MaxEmbedFooterTextLength  = 2048
	MaxEmbedAuthorNameLength  = 256
	MaxEmbedTotalCharacters   = 6000
)

// Sentinel errors for Embed validation.
var (
	ErrEmbedTitleTooLong       = fmt.Errorf("embed title exceeds %d characters", MaxEmbedTitleLength)
	ErrEmbedDescriptionTooLong = fmt.Errorf("embed description exceeds %d characters", MaxEmbedDescriptionLength)
	ErrEmbedTooManyFields      = fmt.Errorf("embed exceeds maximum %d fields", MaxEmbedFields)
	ErrEmbedFieldNameTooLong   = fmt.Errorf("embed field name exceeds %d characters", MaxEmbedFieldNameLength)
	ErrEmbedFieldValueTooLong  = fmt.Errorf("embed field value exceeds %d characters", MaxEmbedFieldValueLength)
	ErrEmbedFooterTooLong      = fmt.Errorf("embed footer text exceeds %d characters", MaxEmbedFooterTextLength)
	ErrEmbedAuthorNameTooLong  = fmt.Errorf("embed author name exceeds %d characters", MaxEmbedAuthorNameLength)
	ErrEmbedTotalTooLong       = fmt.Errorf("embed total characters exceed %d", MaxEmbedTotalCharacters)
)

// Embed represents a rich Discord embed object.
type Embed struct {
	Title       string          `json:"title,omitempty"`
	Type        string          `json:"type,omitempty"`
	Description string          `json:"description,omitempty"`
	URL         string          `json:"url,omitempty"`
	Timestamp   *time.Time      `json:"timestamp,omitempty"`
	Color       int             `json:"color,omitempty"`
	Footer      *EmbedFooter    `json:"footer,omitempty"`
	Image       *EmbedImage     `json:"image,omitempty"`
	Thumbnail   *EmbedThumbnail `json:"thumbnail,omitempty"`
	Video       *EmbedVideo     `json:"video,omitempty"`
	Provider    *EmbedProvider  `json:"provider,omitempty"`
	Author      *EmbedAuthor    `json:"author,omitempty"`
	Fields      []EmbedField    `json:"fields,omitempty"`
}

// EmbedFooter represents footer information in an embed.
type EmbedFooter struct {
	Text         string `json:"text"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedImage represents image details in an embed.
type EmbedImage struct {
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// EmbedThumbnail represents thumbnail details in an embed.
type EmbedThumbnail struct {
	URL      string `json:"url"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   int    `json:"height,omitempty"`
	Width    int    `json:"width,omitempty"`
}

// EmbedVideo represents video details in an embed.
type EmbedVideo struct {
	URL    string `json:"url,omitempty"`
	Height int    `json:"height,omitempty"`
	Width  int    `json:"width,omitempty"`
}

// EmbedProvider represents provider details in an embed.
type EmbedProvider struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

// EmbedAuthor represents author details in an embed.
type EmbedAuthor struct {
	Name         string `json:"name"`
	URL          string `json:"url,omitempty"`
	IconURL      string `json:"icon_url,omitempty"`
	ProxyIconURL string `json:"proxy_icon_url,omitempty"`
}

// EmbedField represents a single field within an embed.
type EmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

// EmbedBuilder provides a fluent interface for constructing and validating an Embed.
type EmbedBuilder struct {
	embed Embed
}

// NewEmbedBuilder creates a new instance of EmbedBuilder.
func NewEmbedBuilder() *EmbedBuilder {
	return &EmbedBuilder{
		embed: Embed{
			Type: "rich",
		},
	}
}

// SetTitle sets the title of the embed.
func (b *EmbedBuilder) SetTitle(title string) *EmbedBuilder {
	b.embed.Title = title
	return b
}

// SetDescription sets the description of the embed.
func (b *EmbedBuilder) SetDescription(desc string) *EmbedBuilder {
	b.embed.Description = desc
	return b
}

// SetURL sets the URL for the embed title.
func (b *EmbedBuilder) SetURL(url string) *EmbedBuilder {
	b.embed.URL = url
	return b
}

// SetColor sets the color code of the embed.
func (b *EmbedBuilder) SetColor(color int) *EmbedBuilder {
	b.embed.Color = color
	return b
}

// SetTimestamp sets the timestamp on the embed.
func (b *EmbedBuilder) SetTimestamp(t time.Time) *EmbedBuilder {
	utc := t.UTC()
	b.embed.Timestamp = &utc
	return b
}

// SetFooter sets the embed footer.
func (b *EmbedBuilder) SetFooter(text string, iconURL ...string) *EmbedBuilder {
	f := &EmbedFooter{Text: text}
	if len(iconURL) > 0 {
		f.IconURL = iconURL[0]
	}
	b.embed.Footer = f
	return b
}

// SetImage sets the main embed image URL.
func (b *EmbedBuilder) SetImage(url string) *EmbedBuilder {
	b.embed.Image = &EmbedImage{URL: url}
	return b
}

// SetThumbnail sets the embed thumbnail URL.
func (b *EmbedBuilder) SetThumbnail(url string) *EmbedBuilder {
	b.embed.Thumbnail = &EmbedThumbnail{URL: url}
	return b
}

// SetAuthor sets the embed author.
func (b *EmbedBuilder) SetAuthor(name, url, iconURL string) *EmbedBuilder {
	b.embed.Author = &EmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return b
}

// AddField appends a field to the embed.
func (b *EmbedBuilder) AddField(name, value string, inline bool) *EmbedBuilder {
	b.embed.Fields = append(b.embed.Fields, EmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return b
}

// Build validates and returns the constructed Embed.
func (b *EmbedBuilder) Build() (Embed, error) {
	totalChars := 0

	if len(b.embed.Title) > MaxEmbedTitleLength {
		return Embed{}, ErrEmbedTitleTooLong
	}
	totalChars += len(b.embed.Title)

	if len(b.embed.Description) > MaxEmbedDescriptionLength {
		return Embed{}, ErrEmbedDescriptionTooLong
	}
	totalChars += len(b.embed.Description)

	if len(b.embed.Fields) > MaxEmbedFields {
		return Embed{}, ErrEmbedTooManyFields
	}

	for _, field := range b.embed.Fields {
		if len(field.Name) > MaxEmbedFieldNameLength {
			return Embed{}, ErrEmbedFieldNameTooLong
		}
		if len(field.Value) > MaxEmbedFieldValueLength {
			return Embed{}, ErrEmbedFieldValueTooLong
		}
		totalChars += len(field.Name) + len(field.Value)
	}

	if b.embed.Footer != nil {
		if len(b.embed.Footer.Text) > MaxEmbedFooterTextLength {
			return Embed{}, ErrEmbedFooterTooLong
		}
		totalChars += len(b.embed.Footer.Text)
	}

	if b.embed.Author != nil {
		if len(b.embed.Author.Name) > MaxEmbedAuthorNameLength {
			return Embed{}, ErrEmbedAuthorNameTooLong
		}
		totalChars += len(b.embed.Author.Name)
	}

	if totalChars > MaxEmbedTotalCharacters {
		return Embed{}, errors.Join(ErrEmbedTotalTooLong, fmt.Errorf("got %d characters", totalChars))
	}

	return b.embed, nil
}
