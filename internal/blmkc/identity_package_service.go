package blmkc

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrIdentityPackageUnavailable = errors.New("identity package is only available after apprehension")

var nikPattern = regexp.MustCompile(`^[0-9]{16}$`)

func (s *Service) IdentityPackageForm(ctx context.Context, caseID uint64) (IdentityPackageForm, error) {
	return s.repository.IdentityPackageForm(ctx, caseID)
}

func (s *Service) SaveIdentityPackage(ctx context.Context, caseID, userID uint64, input IdentityPackageInput) error {
	input.NIK = strings.TrimSpace(input.NIK)
	input.FullName = strings.TrimSpace(input.FullName)
	input.Aliases = strings.TrimSpace(input.Aliases)
	input.BirthPlace = strings.TrimSpace(input.BirthPlace)
	input.Gender = strings.TrimSpace(input.Gender)
	input.Address = strings.TrimSpace(input.Address)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Occupation = strings.TrimSpace(input.Occupation)
	input.IdentificationNotes = strings.TrimSpace(input.IdentificationNotes)
	if caseID == 0 || userID == 0 || input.FullName == "" || utf8.RuneCountInString(input.FullName) > 191 ||
		utf8.RuneCountInString(input.BirthPlace) > 150 || utf8.RuneCountInString(input.Phone) > 30 ||
		utf8.RuneCountInString(input.Occupation) > 150 {
		return ErrInvalid
	}
	if input.Gender != "" && input.Gender != "male" && input.Gender != "female" && input.Gender != "other" {
		return ErrInvalid
	}
	if input.BirthDate != "" {
		if _, err := time.Parse("2006-01-02", input.BirthDate); err != nil {
			return ErrInvalid
		}
	}
	if input.NIK != "" {
		if !nikPattern.MatchString(input.NIK) || len(s.dataKey) < 32 {
			return ErrInvalid
		}
		encrypted, hash, err := encryptNIK(input.NIK, s.dataKey)
		if err != nil {
			return err
		}
		input.NIKEncrypted, input.NIKHash = encrypted, hash
	}
	return s.repository.SaveIdentityPackage(ctx, caseID, userID, input)
}

func encryptNIK(nik, secret string) ([]byte, []byte, error) {
	key := sha256.Sum256([]byte(secret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(nik), nil)
	hash := sha256.Sum256([]byte(nik))
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(ciphertext)))
	base64.StdEncoding.Encode(encoded, ciphertext)
	return encoded, hash[:], nil
}
