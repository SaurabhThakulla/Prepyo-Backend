package billing

import (
	"encoding/base64"
	"testing"
)

// The type a learner's data URL claims is ignored: the bytes decide. An SVG
// labelled as a PNG used to be stored and later served as runnable markup.
func TestDecodeProofSniffsTheBytes(t *testing.T) {
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00")
	svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)

	data, mime, err := decodeProof("data:image/gif;base64," + base64.StdEncoding.EncodeToString(png))
	if err != nil {
		t.Fatalf("real PNG rejected: %v", err)
	}
	if mime != "image/png" || len(data) != len(png) {
		t.Fatalf("got %q (%d bytes), want the sniffed image/png", mime, len(data))
	}

	for _, label := range []string{"image/png", "image/svg+xml"} {
		if _, _, err := decodeProof("data:" + label + ";base64," + base64.StdEncoding.EncodeToString(svg)); err == nil {
			t.Fatalf("SVG labelled %s was accepted", label)
		}
	}
}

// A screenshot at the advertised 4 MB limit must fit in the request body once
// base64 has grown it; the default 1 MiB limit rejected ordinary screenshots.
func TestConfirmBodyFitsAMaxSizeProof(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString(make([]byte, maxProofBytes))
	body := `{"planId":"pro","phoneNumber":"9800000000","transactionId":"TX1","proofImage":"data:image/png;base64,` +
		encoded + `"}`
	if int64(len(body)) > maxConfirmBodyBytes {
		t.Fatalf("a %d-byte proof needs %d bytes of body, limit is %d", maxProofBytes, len(body), int64(maxConfirmBodyBytes))
	}
}
