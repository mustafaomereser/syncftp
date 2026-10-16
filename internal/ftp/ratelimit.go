package ftp

import (
	"context"
	"io"

	"golang.org/x/time/rate"
)

// NewLimiter KB/s cinsinden token bucket limiter üretir; kbps <= 0 → nil (sınırsız).
// Aynı limiter bir sunucunun tüm pool bağlantıları arasında paylaşılır →
// sınır bağlantı başına değil, sunucu başına toplam hızdır.
func NewLimiter(kbps int) *rate.Limiter {
	if kbps <= 0 {
		return nil
	}
	bps := kbps * 1024
	burst := bps
	if burst > 64*1024 {
		burst = 64 * 1024
	}
	return rate.NewLimiter(rate.Limit(bps), burst)
}

// limitedReader her Read'de okunan bayt kadar token bekler.
type limitedReader struct {
	r   io.Reader
	lim *rate.Limiter
}

func newLimitedReader(r io.Reader, lim *rate.Limiter) io.Reader {
	if lim == nil {
		return r
	}
	return &limitedReader{r: r, lim: lim}
}

func (l *limitedReader) Read(p []byte) (int, error) {
	// WaitN burst'ten büyük n kabul etmez → okumayı burst ile sınırla
	if b := l.lim.Burst(); len(p) > b {
		p = p[:b]
	}
	n, err := l.r.Read(p)
	if n > 0 {
		if werr := l.lim.WaitN(context.Background(), n); werr != nil && err == nil {
			err = werr
		}
	}
	return n, err
}
