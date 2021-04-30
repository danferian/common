package PriorityQueue

import "fmt"

type keyPostfix string

const (
	delayedKey = keyPostfix("delayed")
	queuedKey  = keyPostfix("queued")
)

func (kp keyPostfix) String(key string) string {
	return fmt.Sprintf("%s:%s", key, kp)
}
