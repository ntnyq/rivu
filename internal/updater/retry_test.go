package updater

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/egoist/mygo"
)

func TestCheckRetryClassification(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		want bool
	}{
		{"EOF", io.EOF, true},
		{"short JSON body", fmt.Errorf("decode: %w", io.ErrUnexpectedEOF), true},
		{"connection reset", &net.OpError{Op: "read", Err: syscall.ECONNRESET}, true},
		{"socket timeout", &net.DNSError{IsTimeout: true}, true},
		{"attempt timeout", context.DeadlineExceeded, true},
		{"caller cancellation", context.Canceled, false},
		{"unknown certificate", x509.UnknownAuthorityError{}, false},
		{"unknown host", &net.DNSError{IsNotFound: true}, false},
		{"invalid manifest", errors.New("invalid JSON"), false},
		{"disabled updates", mygo.ErrUpdatesDisabled, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := retryableCheckError(test.err); got != test.want {
				t.Fatalf("retryable=%v, want %v", got, test.want)
			}
		})
	}
}

func TestCheckRetryTimeoutRecovery(t *testing.T) {
	calls := 0
	want := &mygo.Update{Version: "2.0.0"}
	got, err := checkWithRetry(context.Background(), 10*time.Millisecond, []time.Duration{0}, func(ctx context.Context) (*mygo.Update, error) {
		calls++
		if calls == 1 {
			<-ctx.Done()
			return nil, ctx.Err()
		}
		return want, nil
	})
	if err != nil || got != want || calls != 2 {
		t.Fatalf("timeout recovery: update=%v error=%v calls=%d", got, err, calls)
	}
}

func TestCheckRetryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	_, err := checkWithRetry(ctx, time.Second, []time.Duration{time.Hour}, func(context.Context) (*mygo.Update, error) {
		calls++
		cancel()
		return nil, io.EOF
	})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("cancellation: error=%v calls=%d", err, calls)
	}
}

func TestCheckRetryBudgetAndCause(t *testing.T) {
	calls := 0
	_, err := checkWithRetry(context.Background(), time.Second, []time.Duration{0, 0}, func(context.Context) (*mygo.Update, error) {
		calls++
		return nil, fmt.Errorf("request: %w", io.EOF)
	})
	if !errors.Is(err, io.EOF) || calls != 3 {
		t.Fatalf("retry limit or cause lost: error=%v calls=%d", err, calls)
	}
}

func TestCheckRetryCancelsBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	_, err := checkWithRetry(ctx, time.Second, []time.Duration{time.Hour}, func(context.Context) (*mygo.Update, error) {
		calls++
		time.AfterFunc(10*time.Millisecond, cancel)
		return nil, io.EOF
	})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("backoff ignored cancellation: error=%v calls=%d", err, calls)
	}
}
