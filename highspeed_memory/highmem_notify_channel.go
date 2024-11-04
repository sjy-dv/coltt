package highspeedmemory

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type NotifyChannel struct {
	attempts     uint
	sleep        time.Duration
	maxSleepTime time.Duration
}

var backlogChannel = make(chan string)

func BackLogging() {
	for backlog := range backlogChannel {
		log.Info().Msg(backlog)
	}
}

func copyNotifyCh() *NotifyChannel {
	return &NotifyChannel{
		attempts:     uint(10),
		sleep:        200 * time.Millisecond,
		maxSleepTime: 3 * time.Second,
	}
}

type Option func(*NotifyChannel)

func attempts(attempts uint) Option {
	return func(nc *NotifyChannel) {
		nc.attempts = attempts
	}
}

func sleep(sleep time.Duration) Option {
	return func(nc *NotifyChannel) {
		nc.sleep = sleep
		if nc.sleep*2 > nc.maxSleepTime {
			nc.maxSleepTime = 2 * nc.sleep
		}
	}
}

func maxSleepTime(maxSleepTime time.Duration) Option {
	return func(nc *NotifyChannel) {
		if nc.sleep*2 > maxSleepTime {
			nc.maxSleepTime = 2 * nc.sleep
		} else {
			nc.maxSleepTime = maxSleepTime
		}
	}
}

func checkCtxValid(ctx context.Context) bool {
	return ctx.Err() != context.DeadlineExceeded && ctx.Err() != context.Canceled
}

// using save/load cdat, json, bin, tensor
func binaryCommitRetry(ctx context.Context, collectionName string, funcName string,
	fn func(string) error, opts ...Option,
) error {
	if !checkCtxValid(ctx) {
		return ctx.Err()
	}
	nc := copyNotifyCh()
	for _, opt := range opts {
		opt(nc)
	}

	var el error

	for i := uint(0); i < nc.attempts; i++ {
		if err := fn(collectionName); err != nil {
			if i%4 == 0 {
				log.Error().Err(err).Msgf("func:%s\nretry time: %d failed\nError:%v", funcName, i, err)
			}
			err = errors.Wrapf(err, "attempt #%d", i)
			el = combine(el, err)

			if !isRecoverable(err) {
				return el
			}

			deadline, ok := ctx.Deadline()
			if ok && time.Until(deadline) < nc.sleep {
				return el
			}
			select {
			case <-time.After(nc.sleep):
			case <-ctx.Done():
				return combine(el, ctx.Err())
			}

			nc.sleep *= 2
			if nc.sleep > nc.maxSleepTime {
				nc.sleep = nc.maxSleepTime
			}
		} else {
			return nil
		}
	}
	return el
}

type multiErrors struct {
	errs []error
}

func (e multiErrors) Unwrap() error {
	if len(e.errs) <= 1 {
		return nil
	}
	// To make merr work for multi errors,
	// we need cause of multi errors, which defined as the last error
	if len(e.errs) == 2 {
		return e.errs[1]
	}

	return multiErrors{
		errs: e.errs[1:],
	}
}

func (e multiErrors) Error() string {
	final := e.errs[0]
	for i := 1; i < len(e.errs); i++ {
		final = errors.Wrap(e.errs[i], final.Error())
	}
	return final.Error()
}

func (e multiErrors) Is(err error) bool {
	for _, item := range e.errs {
		if errors.Is(item, err) {
			return true
		}
	}
	return false
}

func combine(errs ...error) error {
	errs = lo.Filter(errs, func(err error, _ int) bool { return err != nil })
	if len(errs) == 0 {
		return nil
	}
	return multiErrors{
		errs,
	}
}

// Unrecoverable method wrap an error to unrecoverableError. This will make retry
// quick return.
func unrecoverable(err error) error {
	return combine(err, errUnrecoverable)
}

// IsRecoverable is used to judge whether the error is wrapped by unrecoverableError.
func isRecoverable(err error) bool {
	return !errors.Is(err, errUnrecoverable)
}
