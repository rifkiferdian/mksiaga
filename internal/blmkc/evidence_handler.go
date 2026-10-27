package blmkc

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
)

const evidenceStorageRoot = "storage/uploads/blmkc"
const maxPhotoSize int64 = 10 << 20
const maxVideoSize int64 = 150 << 20

var errInvalidEvidence = errors.New("invalid BLMKC evidence")
var errEvidenceTooLarge = errors.New("BLMKC evidence too large")

func (h *Handler) UploadEvidence(c *gin.Context) {
	caseID, ok := caseRouteID(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	subjectID, err := h.service.EvidenceUploadContext(c.Request.Context(), caseID)
	if errors.Is(err, sql.ErrNoRows) {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	header, err := c.FormFile("evidence_file")
	if err != nil {
		h.redirectEvidenceError(c, caseID, errInvalidEvidence)
		return
	}
	file, err := header.Open()
	if err != nil {
		h.redirectEvidenceError(c, caseID, err)
		return
	}
	defer file.Close()
	sniff := make([]byte, 512)
	readCount, readErr := io.ReadFull(file, sniff)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		h.redirectEvidenceError(c, caseID, errInvalidEvidence)
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		h.redirectEvidenceError(c, caseID, err)
		return
	}
	originalName := filepath.Base(strings.TrimSpace(header.Filename))
	mimeType, category, extension, maxSize, valid := evidenceType(originalName, sniff[:readCount])
	if !valid || originalName == "." || utf8.RuneCountInString(originalName) > 255 || header.Size <= 0 {
		h.redirectEvidenceError(c, caseID, errInvalidEvidence)
		return
	}
	if strings.HasPrefix(mimeType, "image/") {
		requestedCategory := strings.TrimSpace(c.PostForm("category"))
		allowedPhotoCategories := map[string]bool{
			"incident_photo": true, "subject_photo": true, "stolen_item_photo": true, "vehicle_photo": true,
		}
		if !allowedPhotoCategories[requestedCategory] {
			h.redirectEvidenceError(c, caseID, errInvalidEvidence)
			return
		}
		category = requestedCategory
	}
	if header.Size > maxSize {
		h.redirectEvidenceError(c, caseID, errEvidenceTooLarge)
		return
	}
	capturedAt, err := optionalDateTime(c.PostForm("captured_at"))
	description := strings.TrimSpace(c.PostForm("evidence_description"))
	if description == "" {
		description = strings.TrimSpace(c.PostForm("description"))
	}
	if err != nil || utf8.RuneCountInString(description) > 500 {
		h.redirectEvidenceError(c, caseID, errInvalidEvidence)
		return
	}
	token, err := randomToken()
	if err != nil {
		h.redirectEvidenceError(c, caseID, err)
		return
	}
	filename := evidenceFilename(category, token, extension)
	relativePath := filepath.Join(evidenceStorageRoot, strconv.FormatUint(caseID, 10), filename)
	absolutePath, err := filepath.Abs(relativePath)
	if err != nil {
		h.redirectEvidenceError(c, caseID, err)
		return
	}
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0750); err != nil {
		h.redirectEvidenceError(c, caseID, err)
		return
	}
	destination, err := os.OpenFile(absolutePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0640)
	if err != nil {
		h.redirectEvidenceError(c, caseID, err)
		return
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(destination, hash), io.LimitReader(file, maxSize+1))
	closeErr := destination.Close()
	if copyErr != nil || closeErr != nil || written != header.Size || written > maxSize {
		_ = os.Remove(absolutePath)
		if written > maxSize {
			h.redirectEvidenceError(c, caseID, errEvidenceTooLarge)
		} else {
			h.redirectEvidenceError(c, caseID, errInvalidEvidence)
		}
		return
	}
	current, _ := auth.CurrentUser(c)
	userID, parseErr := strconv.ParseUint(current.ID, 10, 64)
	if parseErr != nil {
		_ = os.Remove(absolutePath)
		h.internalError(c, parseErr)
		return
	}
	var vehicle *VehicleInput
	if category == "vehicle_photo" {
		vehicle = &VehicleInput{
			Type: c.PostForm("vehicle_type"), Brand: c.PostForm("vehicle_brand"), Model: c.PostForm("vehicle_model"),
			Color: c.PostForm("vehicle_color"), Plate: c.PostForm("vehicle_plate"), Notes: c.PostForm("vehicle_notes"),
		}
	}
	err = h.service.CreateEvidenceWithVehicle(c.Request.Context(), EvidenceInput{
		CaseID: caseID, SubjectID: subjectID, UserID: userID, FileSize: uint64(written), Category: category,
		Filename: filename, StoragePath: filepath.ToSlash(relativePath), MIMEType: mimeType,
		Description: description, FileHash: hash.Sum(nil), CapturedAt: capturedAt,
	}, vehicle)
	if err != nil {
		_ = os.Remove(absolutePath)
		h.redirectEvidenceError(c, caseID, err)
		return
	}
	if c.PostForm("return_to") == "evidence-step" {
		c.Redirect(http.StatusSeeOther, "/blmkc/cases/"+strconv.FormatUint(caseID, 10)+"/evidence/new?success=evidence-uploaded")
		return
	}
	h.redirectCase(c, caseID, "success", "evidence-uploaded")
}

func evidenceFilename(category, token, extension string) string {
	prefixes := map[string]string{
		"incident_photo":    "foto-kejadian",
		"subject_photo":     "foto-pelaku",
		"stolen_item_photo": "foto-barang-dicuri",
		"vehicle_photo":     "foto-kendaraan",
		"cctv_capture":      "foto-cctv",
		"cctv_video":        "video-cctv",
	}
	prefix := prefixes[category]
	if prefix == "" {
		prefix = "bukti"
	}
	unique := token
	if len(unique) > 8 {
		unique = unique[:8]
	}
	return prefix + "-" + unique + strings.ToLower(extension)
}

func (h *Handler) PreviewEvidence(c *gin.Context)  { h.serveEvidence(c, true) }
func (h *Handler) DownloadEvidence(c *gin.Context) { h.serveEvidence(c, false) }

func (h *Handler) serveEvidence(c *gin.Context, inline bool) {
	caseID, ok := caseRouteID(c)
	evidenceID, evidenceOK := evidenceRouteID(c)
	if !ok || !evidenceOK {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	item, err := h.service.Evidence(c.Request.Context(), caseID, evidenceID)
	if errors.Is(err, sql.ErrNoRows) {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	path, err := safeEvidencePath(item.StoragePath)
	if err != nil {
		h.internalError(c, err)
		return
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || uint64(info.Size()) != item.FileSize {
		h.internalError(c, fmt.Errorf("invalid evidence file on disk"))
		return
	}
	current, _ := auth.CurrentUser(c)
	userID, _ := strconv.ParseUint(current.ID, 10, 64)
	action := "download_evidence"
	disposition := "attachment"
	if inline {
		action = "view_evidence"
		disposition = "inline"
	}
	if err := h.service.LogEvidenceAccess(c.Request.Context(), userID, item.StoreID, caseID, action, c.ClientIP(), c.Request.UserAgent()); err != nil {
		slog.Error("log BLMKC evidence access", "error", err)
	}
	c.Header("Content-Type", item.MIMEType)
	c.Header("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": item.Filename}))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "private, no-store")
	http.ServeContent(c.Writer, c.Request, item.Filename, info.ModTime(), file)
}

func evidenceType(filename string, header []byte) (mimeType, category, extension string, maxSize int64, valid bool) {
	extension = strings.ToLower(filepath.Ext(filename))
	detected := http.DetectContentType(header)
	allowed := map[string]struct {
		mime, category string
		max            int64
	}{
		".jpg":  {"image/jpeg", "cctv_capture", maxPhotoSize},
		".jpeg": {"image/jpeg", "cctv_capture", maxPhotoSize},
		".png":  {"image/png", "cctv_capture", maxPhotoSize},
		".webp": {"image/webp", "cctv_capture", maxPhotoSize},
	}
	definition, exists := allowed[extension]
	if !exists {
		return "", "", "", 0, false
	}
	if detected != definition.mime {
		return "", "", "", 0, false
	}
	return definition.mime, definition.category, extension, definition.max, true
}

func optionalDateTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation("2006-01-02T15:04", value, time.Local)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func evidenceRouteID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("evidenceID"), 10, 64)
	return id, err == nil && id > 0
}

func safeEvidencePath(stored string) (string, error) {
	root, err := filepath.Abs(evidenceStorageRoot)
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(filepath.Clean(filepath.FromSlash(stored)))
	if err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("evidence path escapes storage root")
	}
	return path, nil
}

func (h *Handler) redirectEvidenceError(c *gin.Context, caseID uint64, err error) {
	code := "failed"
	switch {
	case errors.Is(err, errInvalidEvidence), errors.Is(err, ErrInvalid):
		code = "invalid-evidence"
	case errors.Is(err, errEvidenceTooLarge):
		code = "evidence-too-large"
	case errors.Is(err, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("BLMKC evidence operation failed", "error", err)
	}
	if c.PostForm("return_to") == "evidence-step" {
		c.Redirect(http.StatusSeeOther, "/blmkc/cases/"+strconv.FormatUint(caseID, 10)+"/evidence/new?error="+url.QueryEscape(code))
		return
	}
	if c.PostForm("return_to") == "verification" || strings.Contains(c.Request.Referer(), "/blmkc/verification/") {
		c.Redirect(http.StatusSeeOther, "/blmkc/verification/"+strconv.FormatUint(caseID, 10)+"?error="+url.QueryEscape(code))
		return
	}
	h.redirectCase(c, caseID, "error", code)
}
