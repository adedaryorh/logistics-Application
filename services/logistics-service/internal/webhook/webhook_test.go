package webhook

import "testing"

func TestSignatureIsStableAndBodyBound(t *testing.T) {
	got := Signature("secret", 1700000000, []byte(`{"status":"picked_up"}`))
	want := "sha256=429a57d5d8ca89d7cf415995dd23eaba7754d34c83dba4dbf7c392f1a89425b1"
	if got != want {
		t.Fatalf("got %s", got)
	}
	if Signature("secret", 1700000000, []byte(`{"status":"delivered"}`)) == got {
		t.Fatal("signature must change with body")
	}
}
