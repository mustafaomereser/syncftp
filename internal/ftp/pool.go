package ftp

import (
	"fmt"
	"sync"

	"syncftp/internal/config"
)

// UploadTask is a single file to be uploaded.
type UploadTask struct {
	LocalPath string
	RelPath   string
	Hash      string
}

// UploadResult holds the outcome of one upload task.
type UploadResult struct {
	RelPath   string
	Hash      string
	Err       error
	Attempts  int
	RetryErrs []error // her başarısız denemenin hatası (son hariç)
}

// Pool manages N concurrent FTP connections to the same server.
type Pool struct {
	clients    []*Client
	maxRetries int
}

// NewPool opens up to srv.MaxConnections parallel FTP connections.
// If some connections fail but at least one succeeds, the pool continues
// with the available connections and prints a warning.
// Returns an error only if zero connections could be established.
func NewPool(srv config.Server) (*Pool, error) {
	n := srv.MaxConnections
	if n <= 0 {
		n = 1
	}
	maxRetries := srv.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	if maxRetries == 0 && srv.MaxRetries == 0 {
		maxRetries = 2
	}

	clients := make([]*Client, 0, n)
	for i := 0; i < n; i++ {
		c, err := Connect(srv)
		if err != nil {
			if i == 0 {
				return nil, err
			}
			fmt.Printf("  ! Bağlantı %d/%d açılamadı, %d bağlantı ile devam ediliyor: %v\n", i+1, n, i, err)
			break
		}
		clients = append(clients, c)
	}

	return &Pool{clients: clients, maxRetries: maxRetries}, nil
}

// Close shuts down all connections in the pool.
func (p *Pool) Close() {
	for _, c := range p.clients {
		c.Close()
	}
}

// UploadCh distributes tasks across workers and streams results one-by-one via a channel.
// The channel is closed when all uploads finish.
func (p *Pool) UploadCh(tasks []UploadTask) <-chan UploadResult {
	ch := make(chan UploadResult, len(p.clients))
	if len(tasks) == 0 {
		close(ch)
		return ch
	}

	jobs := make(chan UploadTask, len(tasks))
	var wg sync.WaitGroup
	for _, c := range p.clients {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			for task := range jobs {
				ch <- p.uploadWithRetry(client, task)
			}
		}(c)
	}
	for _, t := range tasks {
		jobs <- t
	}
	close(jobs)
	go func() {
		wg.Wait()
		close(ch)
	}()
	return ch
}

// Upload distributes tasks across all workers and returns one result per task.
// Failed uploads are retried up to MaxRetries times on the same worker before
// being reported as failed.
func (p *Pool) Upload(tasks []UploadTask) []UploadResult {
	if len(tasks) == 0 {
		return nil
	}

	jobs := make(chan UploadTask, len(tasks))
	results := make(chan UploadResult, len(tasks))

	var wg sync.WaitGroup
	for _, c := range p.clients {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			for task := range jobs {
				result := p.uploadWithRetry(client, task)
				results <- result
			}
		}(c)
	}

	for _, t := range tasks {
		jobs <- t
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	out := make([]UploadResult, 0, len(tasks))
	for r := range results {
		out = append(out, r)
	}
	return out
}

func (p *Pool) uploadWithRetry(client *Client, task UploadTask) UploadResult {
	var lastErr error
	var retryErrs []error
	maxAttempts := p.maxRetries + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		lastErr = client.Upload(task.LocalPath, task.RelPath)
		if lastErr == nil {
			return UploadResult{
				RelPath:   task.RelPath,
				Hash:      task.Hash,
				Attempts:  attempt,
				RetryErrs: retryErrs,
			}
		}
		retryErrs = append(retryErrs, lastErr)
	}
	return UploadResult{
		RelPath:   task.RelPath,
		Err:       lastErr,
		Attempts:  maxAttempts,
		RetryErrs: retryErrs[:len(retryErrs)-1], // son hata zaten Err'da, RetryErrs'da tekrar etme
	}
}
