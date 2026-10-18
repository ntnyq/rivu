package updater

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/egoist/mygo"
)

func checkWithRetry(ctx context.Context, timeout time.Duration, delays []time.Duration, check func(context.Context) (*mygo.Update, error)) (*mygo.Update, error) {
	for attempt := 0; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		attemptCtx, cancel := context.WithTimeout(ctx, timeout)
		update, err := check(attemptCtx)
		cancel()
		if err == nil {
			return update, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if attempt >= len(delays) || !retryableCheckError(err) {
			return nil, fmt.Errorf("update check failed after %d attempt(s): %w", attempt+1, err)
		}
		timer := time.NewTimer(delays[attempt])
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

func retryableCheckError(err error) bool {
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
		errors.Is(err, context.DeadlineExceeded) || errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNABORTED) || errors.Is(err, syscall.EPIPE) || platformRetryableCheckError(err) {
		return true
	}
	var networkError net.Error
	return errors.As(err, &networkError) && networkError.Timeout()
}
