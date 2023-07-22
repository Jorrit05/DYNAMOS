package etcd

import (
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// RetryOptions represents the configuration options for retries.
type RetryOptions struct {
	InitialInterval time.Duration
	MaxInterval     time.Duration
	MaxElapsedTime  time.Duration
	AddJsonTrace    bool
}

// DefaultRetryOptions provides a RetryOptions with default values.
var DefaultRetryOptions = RetryOptions{
	InitialInterval: time.Second * 1,
	MaxInterval:     time.Second * 3,
	MaxElapsedTime:  time.Second * 30, // Max retry time
	//Todo, shouldb't be part of retry options, rename to generic options..
	AddJsonTrace: false,
}

// Option is a function that applies a configuration option to RetryOptions.
type Option func(*RetryOptions)

// WithInitialInterval sets the initial interval for retries.
func WithJsonTrace() Option {
	return func(opts *RetryOptions) {
		opts.AddJsonTrace = true
	}
}

// WithInitialInterval sets the initial interval for retries.
func WithInitialInterval(d time.Duration) Option {
	return func(opts *RetryOptions) {
		opts.InitialInterval = d
	}
}

// WithMaxInterval sets the max interval for retries.
func WithMaxInterval(d time.Duration) Option {
	return func(opts *RetryOptions) {
		opts.MaxInterval = d
	}
}

// WithMaxElapsedTime sets the max elapsed time for retries.
func WithMaxElapsedTime(d time.Duration) Option {
	return func(opts *RetryOptions) {
		opts.MaxElapsedTime = d
	}
}

func GetEtcdClient(endpoints string) *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(endpoints, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logger.Sugar().Fatalw("Error in creating ETCD client %v", err)
	}

	return cli
}
