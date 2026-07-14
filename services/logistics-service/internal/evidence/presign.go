package evidence

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/url"
	"path"
	"strings"
	"time"
)

type Upload struct {
	UploadURL   string    `json:"upload_url"`
	EvidenceURL string    `json:"evidence_url"`
	ExpiresAt   time.Time `json:"expires_at"`
}
type Presigner interface {
	Presign(orderID, fileName, contentType string, timeToLive time.Duration) (Upload, error)
}
type HMACPresigner struct{ UploadBaseURL, PublicBaseURL, Secret string }

func (p HMACPresigner) Presign(orderID, fileName, contentType string, ttl time.Duration) (Upload, error) {
	if p.UploadBaseURL == "" || p.PublicBaseURL == "" || p.Secret == "" {
		return Upload{}, errors.New("evidence storage is not configured")
	}
	ext := path.Ext(path.Base(fileName))
	key := fmt.Sprintf("delivery-evidence/%s/%s%s", orderID, uuid.NewString(), ext)
	expires := time.Now().UTC().Add(ttl)
	canonical := fmt.Sprintf("PUT\n%s\n%d\n%s", key, expires.Unix(), contentType)
	mac := hmac.New(sha256.New, []byte(p.Secret))
	mac.Write([]byte(canonical))
	query := url.Values{"expires": {fmt.Sprint(expires.Unix())}, "content_type": {contentType}, "signature": {hex.EncodeToString(mac.Sum(nil))}}
	return Upload{UploadURL: strings.TrimRight(p.UploadBaseURL, "/") + "/" + key + "?" + query.Encode(), EvidenceURL: strings.TrimRight(p.PublicBaseURL, "/") + "/" + key, ExpiresAt: expires}, nil
}
