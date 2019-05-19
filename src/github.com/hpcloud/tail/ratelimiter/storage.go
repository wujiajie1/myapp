package ratelimiter

import "vendor"

type Storage interface {
	GetBucketFor(string) (*vendor.LeakyBucket, error)
	SetBucketFor(string, vendor.LeakyBucket) error
}
