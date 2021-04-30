package pqueue

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const (
	rateLimitPeriod = time.Minute
)

var (
	invalidRedisKey = fmt.Errorf("invalid redis key")
)

type (
	client struct {
		rdb         *redis.ClusterClient
		delayedKey  string
		queuedKey   string
		rateLimit   int
		handlerFunc func(msg string)
		logger      *logrus.Logger
		wg          *sync.WaitGroup
	}
	PQueue struct {
		RetryCount int         `json:"retry_count"`
		Score      int64       `json:"score"`
		Msg        interface{} `json:"msg"`
	}
	Client interface {
		Push(request *PQueue) error
		DelayedLength() (int, error)
		QueuedLength() (int, error)
		Wait()
	}
)

func NewRedisPriorityQueue(logger *logrus.Logger, address []string, key string, limit int, handlerFunc func(msg string)) (Client, error) {
	logger.Info("initializing new redis priority queue")

	ctx := context.Background()
	defer ctx.Done()

	rdb := redis.NewClusterClient(
		&redis.ClusterOptions{
			Addrs:        address,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			MinIdleConns: 16,
			MaxConnAge:   5 * time.Second,
		})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		logger.Errorf("an error occurred when try to ping rdb. err: %v", err)
		return nil, err
	}

	if key == "" {
		logger.Errorf("invalid redis key. err: %v", invalidRedisKey)
		return nil, invalidRedisKey
	}

	var rateLimit = 1000
	if limit > 0 {
		rateLimit = limit
	}

	delayedKey := delayedKey.String(key)
	queuedKey := queuedKey.String(key)

	client := &client{
		rdb:         rdb,
		delayedKey:  delayedKey,
		queuedKey:   queuedKey,
		rateLimit:   rateLimit,
		handlerFunc: handlerFunc,
		logger:      logger,
		wg:          &sync.WaitGroup{},
	}

	go client.delayedQueueConsume()
	go client.queueConsume()

	return client, nil
}

func (c *client) delayedQueueConsume() {
	ctx := context.Background()

	for {
		items, err := c.rdb.ZRangeWithScores(ctx, c.delayedKey, 0, 0).Result()
		if err != nil {
			c.logger.Errorf("an error occurred when try to fetch first message of member. err: %v", err)
			time.Sleep(1)
			continue
		}

		if len(items) < 1 {
			time.Sleep(1)
			continue
		}

		if items[0].Score > float64(time.Now().Unix()) {
			time.Sleep(1)
			continue
		}

		item := items[0]
		if err := c.rdb.ZRem(ctx, c.delayedKey, item.Member).Err(); err != nil {
			c.logger.Errorf("an error occurred when try to remove member. err: %v", err)
			time.Sleep(1)
			continue
		}

		if err := c.rdb.RPush(ctx, c.queuedKey, item.Member).Err(); err != nil {
			c.logger.Errorf("an error occurred when try to push value to list. err: %v", err)

			c.rdb.ZAddNX(ctx, c.delayedKey, &item)
			time.Sleep(1)
			continue
		}
	}
}

func (c *client) queueConsume() {
	ctx := context.Background()

	quotas := make(chan time.Time, c.rateLimit)
	go func() {
		tick := time.NewTicker(rateLimitPeriod / time.Duration(c.rateLimit))
		defer tick.Stop()
		for t := range tick.C {
			select {
			case quotas <- t:
			default:
			}
		}
	}()

	for {
		<-quotas

		msg, err := c.rdb.LPop(ctx, c.queuedKey).Result()
		if err != nil {
			continue
		}

		c.wg.Add(1)
		go c.handlerFunc(msg)
	}
}

func (c *client) Push(request *PQueue) error {
	ctx := context.Background()

	msg, err := json.Marshal(request)
	if err != nil {
		c.logger.Errorf("an error occurred when try to marshal request. err: %v", err)
		return err
	}

	z := &redis.Z{
		Score:  float64(request.Score),
		Member: string(msg),
	}

	if err := c.rdb.ZAddNX(ctx, c.delayedKey, z).Err(); err != nil {
		c.logger.Errorf("an error occurred when try to add redis member. err: %v", err)
		return err
	}

	return nil
}

func (c *client) DelayedLength() (int, error) {
	delayedMsg, err := c.rdb.ZRange(context.Background(), c.delayedKey, 0, -1).Result()
	return len(delayedMsg), err
}

func (c *client) QueuedLength() (int, error) {
	queuedMsg, err := c.rdb.LRange(context.Background(), c.queuedKey, 0, -1).Result()
	return len(queuedMsg), err
}

func (c *client) Wait() {
	c.wg.Wait()
}
