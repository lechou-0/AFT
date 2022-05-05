package storage

import (
	"fmt"
	"os"
	"sync"
	"context"
	"time"

	rdslib "github.com/go-redis/redis"
	"github.com/golang/protobuf/proto"
	pb "github.com/lechou-0/AFT/proto/aft"
)

type RedisStorageManager struct {
	client *rdslib.Client
}

func NewRedisStorageManager(address string, password string) *RedisStorageManager {
	rc := rdslib.NewClient(&rdslib.Options{
		Addr: "192.168.1.104:6379",
		Password: "1234",
		DB: 0,
		PoolSize: 100,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rc.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("Unexpected error while connecting to Redis client:\n%v\n", err)
		os.Exit(1)
	}

	return &RedisStorageManager{client: rc}
}

func (redis *RedisStorageManager) StartTransaction(id string) error {
	return nil
}

func (redis *RedisStorageManager) CommitTransaction(transaction *pb.TransactionRecord) error {
	ctx := context.Background()
	key := fmt.Sprintf(TransactionKey, transaction.Id, transaction.Timestamp)
	serialized, err := proto.Marshal(transaction)
	if err != nil {
		return err
	}

	return redis.client.Set(ctx, key, serialized, 0).Err()
}

func (redis *RedisStorageManager) AbortTransaction(transaction *pb.TransactionRecord) error {
	// TODO: Delete the aborted keys.
	return nil
}

func (redis *RedisStorageManager) Get(key string) (*pb.KeyValuePair, error) {
	ctx := context.Background()
	result := &pb.KeyValuePair{}

	val, err := redis.client.Get(ctx, key).Result()
	if err != nil {
		return result, err
	}

	err = proto.Unmarshal([]byte(val), result)
	return result, err
}

func (redis *RedisStorageManager) GetTransaction(transactionKey string) (*pb.TransactionRecord, error) {
	ctx := context.Background()
	result := &pb.TransactionRecord{}

	val, err := redis.client.Get(ctx, transactionKey).Result()
	if err != nil {
		return result, err
	}

	err = proto.Unmarshal([]byte(val), result)
	return result, err
}

func (redis *RedisStorageManager) MultiGetTransaction(transactionKeys *[]string) (*[]*pb.TransactionRecord, error) {
	ctx := context.Background()
	results := make([]*pb.TransactionRecord, len(*transactionKeys))

	for index, key := range *transactionKeys {
		txn, err := redis.GetTransaction(ctx, key)
		if err != nil {
			return &[]*pb.TransactionRecord{}, err
		}

		results[index] = txn
	}

	return &results, nil
}

func (redis *RedisStorageManager) Put(key string, val *pb.KeyValuePair) error {
	ctx := context.Background()
	serialized, err := proto.Marshal(val)
	if err != nil {
		return err
	}

	return redis.client.Set(ctx, key, serialized, 0).Err()
}

func (redis *RedisStorageManager) MultiPut(data *map[string]*pb.KeyValuePair) error {
	ctx := context.Background()
	for key, val := range *data {
		err := redis.Put(ctx, key, val)
		if err != nil {
			return err
		}
	}

	return nil
}

func (redis *RedisStorageManager) Delete(key string) error {
	ctx := context.Background()
	return redis.client.Del(ctx, key).Err()
}

func (redis *RedisStorageManager) MultiDelete(keys *[]string) error {
	ctx := context.Background()
	for _, key := range *keys {
		err := redis.Delete(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (redis *RedisStorageManager) List(prefix string) ([]string, error) {
	result := []string{}
	redisPrefix := fmt.Sprintf("%s*", prefix)
	mtx := &sync.Mutex{}
	ctx := context.Background()
	err := redis.client.ForEachMaster(ctx, func(master *rdslib.Client) error {
		cursor := uint64(0)
		additionalKeys := true
		ctx1 := context.Background()
		for additionalKeys {
			var scanKeys []string
			var err error
			scanKeys, cursor, err = master.Scan(ctx1, cursor, redisPrefix, 100).Result()

			if err != nil {
				return err
			}

			mtx.Lock()
			result = append(result, scanKeys...)
			mtx.Unlock()

			if cursor == 0 {
				additionalKeys = false
			}
		}

		return nil
	})

	if err != nil {
		return []string{}, err
	}

	return result, nil
}
