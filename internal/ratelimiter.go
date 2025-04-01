package internal

type IRateLimiter interface {
	Allow() bool
	Wait()
}
