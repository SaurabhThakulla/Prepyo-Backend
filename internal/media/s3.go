package media

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// S3 talks to Backblaze B2 through its S3-compatible API.
// The bucket stays private: files are only fetched by the server and passed
// on to their owner, so no public URL or presigned link exists.
type S3 struct {
	base      string // https://<endpoint>/<bucket>
	region    string
	accessKey string
	secretKey string
	client    *http.Client
	now       func() time.Time
}

// NewS3 takes the bucket's endpoint, such as "s3.us-east-005.backblazeb2.com".
// An empty region is read from that endpoint.
func NewS3(endpoint, region, bucket, accessKey, secretKey string) *S3 {
	host := strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://"), "/")
	if region == "" {
		region = "us-east-1"
		if parts := strings.Split(host, "."); len(parts) > 2 && parts[0] == "s3" && strings.HasSuffix(host, ".backblazeb2.com") {
			region = parts[1]
		}
	}
	return &S3{
		base:      "https://" + host + "/" + bucket,
		region:    region,
		accessKey: accessKey,
		secretKey: secretKey,
		client:    &http.Client{Timeout: 60 * time.Second},
		now:       time.Now,
	}
}

func (s *S3) Put(ctx context.Context, key, contentType string, body []byte) error {
	resp, err := s.do(ctx, http.MethodPut, key, body, contentType)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return s.failure("put", key, resp)
	}
	return nil
}

func (s *S3) Get(ctx context.Context, key string) (io.ReadCloser, string, error) {
	resp, err := s.do(ctx, http.MethodGet, key, nil, "")
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, "", ErrNotFound
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		return nil, "", s.failure("get", key, resp)
	}
	return resp.Body, resp.Header.Get("Content-Type"), nil
}

// Delete removes a file. On B2 this hides it; the bucket's "keep only the last
// version" lifecycle rule then deletes it for good.
func (s *S3) Delete(ctx context.Context, key string) error {
	resp, err := s.do(ctx, http.MethodDelete, key, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 && resp.StatusCode != http.StatusNotFound {
		return s.failure("delete", key, resp)
	}
	return nil
}

func (s *S3) do(ctx context.Context, method, key string, body []byte, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.base+"/"+key, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	sign(req, body, s.accessKey, s.secretKey, s.region, s.now())
	return s.client.Do(req)
}

// failure keeps the store's error body (XML) but never echoes credentials.
func (s *S3) failure(op, key string, resp *http.Response) error {
	detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("media: s3 %s %s: %s: %s", op, key, resp.Status, strings.TrimSpace(string(detail)))
}

// sign adds AWS Signature Version 4 headers. Every header already on the
// request is signed, plus host; requests carry no query string.
func sign(req *http.Request, payload []byte, accessKey, secretKey, region string, now time.Time) {
	amzDate := now.UTC().Format("20060102T150405Z")
	day := amzDate[:8]
	payloadHash := hexSHA256(payload)
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	values := map[string]string{"host": req.URL.Host}
	for name, v := range req.Header {
		values[strings.ToLower(name)] = strings.TrimSpace(strings.Join(v, ","))
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	var headers strings.Builder
	for _, name := range names {
		headers.WriteString(name + ":" + values[name] + "\n")
	}
	signed := strings.Join(names, ";")

	canonical := strings.Join([]string{req.Method, req.URL.EscapedPath(), req.URL.RawQuery, headers.String(), signed, payloadHash}, "\n")
	scope := day + "/" + region + "/s3/aws4_request"
	toSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hexSHA256([]byte(canonical))

	k := hmacSHA256([]byte("AWS4"+secretKey), day)
	k = hmacSHA256(k, region)
	k = hmacSHA256(k, "s3")
	k = hmacSHA256(k, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(k, toSign))
	req.Header.Set("Authorization", fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, scope, signed, signature))
}

func hexSHA256(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(data))
	return m.Sum(nil)
}
