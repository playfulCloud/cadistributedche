package store

import "time"

type KeyValueEntry struct {
	key       string
	value     string
	createdAt time.Time
	ttl       time.Duration
}

func (k *KeyValueEntry) Key() string {
	return k.key
}

func (k *KeyValueEntry) Value() string {
	return k.value
}

func (k *KeyValueEntry) CreatedAt() time.Time {
	return k.createdAt
}

func (k *KeyValueEntry) Ttl() time.Duration {
	return k.ttl
}

func (k *KeyValueEntry) isExpired(now time.Time) bool {
	if k.ttl <= 0 {
		return false
	}

	diff := now.Sub(k.createdAt)
	return diff >= k.ttl
}
