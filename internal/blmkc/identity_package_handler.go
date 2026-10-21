package blmkc

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
)

type identityDocument struct {
	Field, Category, Description string
	PhotoOnly                    bool
}

var identityDocuments = []identityDocument{
	{"subject_photo", "subject_photo", "Foto pelaku", true},
	{"identity_card", "identity_card", "Foto atau salinan KTP", false},
}

var resolutionDocuments = []identityDocument{
	{"incident_photo", "incident_photo", "Foto bukti tindak kejahatan", true},
	{"statement_letter", "statement_letter", "Surat pernyataan", false},
	{"settlement_letter", "settlement_letter", "Surat perdamaian", false},
	{"case_resolution", "case_resolution", "Berkas penyelesaian kasus", false},
}

func (h *Handler) IdentityPackageForm(c *gin.Context) {
	id, ok := caseRouteID(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	data, err := h.service.IdentityPackageForm(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	returnTo := c.Query("return_to")
	if returnTo == "" {
		if referer, err := url.Parse(c.Request.Referer()); err == nil {
			returnTo = referer.Path
		}
	}
	data.ReturnTo = identityPackageReturnTo(returnTo, id)
	h.render(c, "blmkc/identity_package_form.html", "Lengkapi Identity BLMKC", "blmkc-identity", gin.H{
		"Data": data, "SuccessMessage": caseMessage(c.Query("success")), "ErrorMessage": caseMessage(c.Query("error")),
	})
}

func (h *Handler) SaveIdentityPackage(c *gin.Context) {
	id, ok := caseRouteID(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	redirect := "/blmkc/cases/" + strconv.FormatUint(id, 10) + "/identity-package"
	returnTo := identityPackageReturnTo(c.PostForm("return_to"), id)
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		c.Redirect(http.StatusSeeOther, redirect+"?error=csrf&return_to="+url.QueryEscape(returnTo))
		return
	}
	current, _ := auth.CurrentUser(c)
	userID, err := strconv.ParseUint(current.ID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	input := IdentityPackageInput{
		NIK: c.PostForm("nik"), FullName: c.PostForm("full_name"), Aliases: c.PostForm("aliases"),
		BirthPlace: c.PostForm("birth_place"), BirthDate: c.PostForm("birth_date"), Gender: c.PostForm("gender"),
		Address: c.PostForm("address"), Phone: c.PostForm("phone"), Occupation: c.PostForm("occupation"),
		IdentificationNotes: c.PostForm("identification_notes"),
	}
	if err := h.service.SaveIdentityPackage(c.Request.Context(), id, userID, input); err != nil {
		h.redirectIdentityPackageError(c, redirect, returnTo, err)
		return
	}
	form, err := h.service.IdentityPackageForm(c.Request.Context(), id)
	if err != nil {
		h.redirectIdentityPackageError(c, redirect, returnTo, err)
		return
	}
	for _, document := range identityDocuments {
		header, err := c.FormFile(document.Field)
		if errors.Is(err, http.ErrMissingFile) {
			continue
		}
		if err != nil {
			h.redirectIdentityPackageError(c, redirect, returnTo, errInvalidEvidence)
			return
		}
		if err := h.saveCaseDocument(c, form.CaseID, form.SubjectID, userID, document, header); err != nil {
			h.redirectIdentityPackageError(c, redirect, returnTo, err)
			return
		}
	}
	c.Redirect(http.StatusSeeOther, returnTo+"?success=identity-completed")
}

func (h *Handler) saveCaseDocument(c *gin.Context, caseID, subjectID, userID uint64, document identityDocument, header *multipart.FileHeader) error {
	if header.Size <= 0 || header.Size > 20<<20 || utf8.RuneCountInString(filepath.Base(header.Filename)) > 255 {
		return errEvidenceTooLarge
	}
	file, err := header.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	sniff := make([]byte, 512)
	n, readErr := io.ReadFull(file, sniff)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return errInvalidEvidence
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	mimeType := http.DetectContentType(sniff[:n])
	extension := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "application/pdf": ".pdf"}[mimeType]
	if extension == "" || (document.PhotoOnly && mimeType == "application/pdf") || (strings.HasPrefix(mimeType, "image/") && header.Size > maxPhotoSize) {
		return errInvalidEvidence
	}
	token, err := randomToken()
	if err != nil {
		return err
	}
	relativePath := filepath.Join(evidenceStorageRoot, strconv.FormatUint(caseID, 10), token+extension)
	absolutePath, err := filepath.Abs(relativePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0750); err != nil {
		return err
	}
	destination, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		return err
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(destination, hash), io.LimitReader(file, (20<<20)+1))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil || written != header.Size || written > 20<<20 {
		_ = os.Remove(absolutePath)
		return errInvalidEvidence
	}
	err = h.service.CreateEvidence(c.Request.Context(), EvidenceInput{
		CaseID: caseID, SubjectID: subjectID, UserID: userID, FileSize: uint64(written),
		Category: document.Category, Filename: filepath.Base(header.Filename), StoragePath: filepath.ToSlash(relativePath),
		MIMEType: mimeType, Description: document.Description, FileHash: hash.Sum(nil),
	})
	if err != nil {
		_ = os.Remove(absolutePath)
	}
	return err
}

func (h *Handler) redirectIdentityPackageError(c *gin.Context, path, returnTo string, err error) {
	code := "failed"
	switch {
	case errors.Is(err, ErrInvalid):
		code = "invalid-identity"
	case errors.Is(err, ErrIdentityPackageUnavailable), errors.Is(err, sql.ErrNoRows):
		code = "not-found"
	case errors.Is(err, errInvalidEvidence):
		code = "invalid-identity-document"
	case errors.Is(err, errEvidenceTooLarge):
		code = "identity-document-too-large"
	default:
		slog.Error("save BLMKC identity package", "error", err)
	}
	c.Redirect(http.StatusSeeOther, path+"?error="+code+"&return_to="+url.QueryEscape(returnTo))
}

func identityPackageReturnTo(value string, caseID uint64) string {
	value = strings.TrimSpace(value)
	for _, prefix := range []string{"/blmkc/identity/", "/blmkc/cases/"} {
		if strings.HasPrefix(value, prefix) {
			id, err := strconv.ParseUint(strings.TrimPrefix(value, prefix), 10, 64)
			if err == nil && id > 0 {
				return value
			}
		}
	}
	return "/blmkc/cases/" + strconv.FormatUint(caseID, 10)
}
