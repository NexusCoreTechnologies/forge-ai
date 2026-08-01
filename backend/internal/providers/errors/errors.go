package errors

import "errors"

var (
    ErrAuth        = errors.New("authentication error")
    ErrNetwork     = errors.New("network error")
    ErrTimeout     = errors.New("timeout")
    ErrRateLimit   = errors.New("rate limit")
    ErrQuota       = errors.New("quota exceeded")
    ErrConfig      = errors.New("configuration error")
    ErrUnknown     = errors.New("unknown provider error")
)
