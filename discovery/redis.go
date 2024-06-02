package discovery

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	discoveryKey  string = "dfsgo:nodes"
	infoKeyPrefix string = "dfsgo:nodes:info"
)

type RedisDiscoverySrv struct {
	client *redis.Client
	ctx    context.Context
	logger *zap.Logger
}

func NewRedisDiscoverySrv(ctx context.Context, redisURL string) (*RedisDiscoverySrv, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opts)
	return &RedisDiscoverySrv{
		client: client,
		ctx:    ctx,
	}, nil
}

func (srv *RedisDiscoverySrv) Close() error {
	return srv.client.Close()
}

func (srv *RedisDiscoverySrv) AddNode(node Node) error {
	b, err := json.Marshal(node)
	if err != nil {
		return err
	}
	return srv.client.HSet(srv.ctx, discoveryKey, node.Address, string(b)).Err()
}

func (srv *RedisDiscoverySrv) GetNodes() ([]Node, error) {
	nodesStr, err := srv.client.HGetAll(srv.ctx, discoveryKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("there is no node in discovery service")
	}
	if err != nil {
		return nil, err
	}
	var nodes []Node
	var node Node

	for addr, nodeStr := range nodesStr {
		err := json.Unmarshal([]byte(nodeStr), &node)
		if err != nil {
			srv.logger.Error("could not parse node info", zap.String("node_addr", addr))
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes, nil

}

func (srv *RedisDiscoverySrv) RemoveDeadNode(n Node) error {
	err := srv.client.HDel(srv.ctx, discoveryKey, n.Address).Err()
	if err != nil {
		srv.logger.Error("could not delete node from discovery service", zap.String("node_addr", n.Address))
		return err
	}
	return nil
}
