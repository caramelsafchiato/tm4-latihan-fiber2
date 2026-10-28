package service

import (
	"errors"
	"strings"
)

// Custom error untuk validasi ulasan (Endpoint 4)
var (
	ErrInvalidRating   = errors.New("rating: nilai maksimal 5")
	ErrCommentTooShort = errors.New("comment: minimal 10 karakter")
	ErrTooManyTags     = errors.New("tags: maksimal 5 item")
)

// ValidateReviewInput memeriksa aturan validasi payload ulasan
func ValidateReviewInput(rating int, comment string, tags []string) error {
	if rating < 1 || rating > 5 {
		return ErrInvalidRating
	}
	if len(strings.TrimSpace(comment)) < 10 {
		return ErrCommentTooShort
	}
	if len(tags) > 5 {
		return ErrTooManyTags
	}
	return nil
}

// NormalizeTags merapikan tag sebelum disimpan: huruf kecil, tanpa spasi di
// tepi, tanpa duplikat, urutan pertama dipertahankan.
func NormalizeTags(tags []string) []string {
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		result = append(result, t)
	}
	return result
}