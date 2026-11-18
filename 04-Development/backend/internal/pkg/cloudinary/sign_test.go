package cloudinary

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSign(t *testing.T) {
	c := &Client{apiSecret: "abcd"}
	got := c.sign(map[string]string{"timestamp": "1315060510", "public_id": "sample_image", "eager": "w_400,h_300"})
	sum := sha256.Sum256([]byte("eager=w_400,h_300&public_id=sample_image&timestamp=1315060510abcd"))
	if want := hex.EncodeToString(sum[:]); got != want {
		t.Fatalf("sign = %s, want %s", got, want)
	}
}
