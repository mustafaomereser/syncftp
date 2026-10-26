package ftp

import (
	"context"
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"syncftp/internal/config"
)

// chunkReader veriyi sabit boyutlu parçalarla verir (\r ve \n farklı Read'lere düşsün).
type chunkReader struct {
	data []byte
	size int
}

func (r *chunkReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	n := r.size
	if n > len(p) {
		n = len(p)
	}
	if n > len(r.data) {
		n = len(r.data)
	}
	copy(p, r.data[:n])
	r.data = r.data[n:]
	return n, nil
}

func TestCRLFNormalizerAcrossChunks(t *testing.T) {
	cases := map[string]string{
		"a\r\nb\r\n":   "a\nb\n",
		"a\rb":         "a\rb",
		"a\r":          "a\r",
		"\r\r\n":       "\r\n",
		"x\r\n\r\ny":   "x\n\ny",
		"no newline":   "no newline",
		"\r\n\r\n\r\n": "\n\n\n",
	}
	for in, want := range cases {
		for size := 1; size <= len(in)+1; size++ {
			got, err := io.ReadAll(newCRLFNormalizer(&chunkReader{data: []byte(in), size: size}))
			if err != nil {
				t.Fatalf("%q size %d: %v", in, size, err)
			}
			if string(got) != want {
				t.Fatalf("%q size %d: got %q want %q", in, size, got, want)
			}
		}
		// Tek baytlık çıkış tamponu + tek baytlık okuma
		got, err := io.ReadAll(iotest.OneByteReader(newCRLFNormalizer(iotest.OneByteReader(strings.NewReader(in)))))
		if err != nil || string(got) != want {
			t.Fatalf("%q one-byte: got %q (%v) want %q", in, got, err, want)
		}
	}
}

func TestHostSemaphoreLimitRecomputedAfterClose(t *testing.T) {
	srv := config.Server{Protocol: "ftp", Host: "recompute", Port: 21, User: "u"}
	s := semFor(srv, 1)
	s.unregister(1) // pool kapandı
	s = semFor(srv, 3)
	defer s.unregister(3)
	for i := 0; i < 3; i++ {
		if !s.tryAcquire() {
			t.Fatalf("slot %d should be free with raised limit", i)
		}
	}
	if s.tryAcquire() {
		t.Fatal("limit 3 exceeded")
	}
	// Eş zamanlı daha küçük limitli pool → en küçük geçerli
	s2 := semFor(srv, 2)
	s.release(3)
	s2.acquire(context.Background())
	s2.acquire(context.Background())
	if s2.tryAcquire() {
		t.Fatal("smallest concurrent limit (2) exceeded")
	}
	s2.release(2)
	s2.unregister(2)
}
