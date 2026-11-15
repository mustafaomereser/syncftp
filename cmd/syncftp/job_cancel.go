package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"syncftp/internal/config"
	ftpclient "syncftp/internal/ftp"
	"syncftp/internal/lang"
)

// connectCtx bağlanır; ctx iptal edilirse bağlantıyı beklemeden döner (geç gelen bağlantı kapatılır).
// Panel işlerindeki "Durdur" cevapsız sunucuda da hemen çalışsın diye.
func connectCtx(ctx context.Context, srv config.Server) (*ftpclient.Client, error) {
	type res struct {
		c   *ftpclient.Client
		err error
	}
	ch := make(chan res, 1)
	go func() {
		c, err := ftpclient.Connect(srv)
		ch <- res{c, err}
	}()
	select {
	case r := <-ch:
		return r.c, r.err
	case <-ctx.Done():
		go func() {
			if r := <-ch; r.c != nil {
				r.c.Close()
			}
		}()
		return nil, ctx.Err()
	}
}

// listProgress uzun listelemede 2 sn'de bir "N dosya listelendi · son yol" satırı yazar (calibrate'teki
// gibi). progress false ise nil: CLI'da satır satır akmasın.
func listProgress(w io.Writer, progress bool) func(int, string) {
	if !progress {
		return nil
	}
	last := time.Now()
	return func(n int, latest string) {
		if time.Since(last) >= 2*time.Second {
			fmt.Fprintf(w, lang.L.ResyncListLiveFmt, n, latest)
			last = time.Now()
		}
	}
}
