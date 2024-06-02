package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRedisDiscoverySrv_AddNode(t *testing.T) {
	var ctx = context.TODO()
	client, mock := redismock.NewClientMock()

	srv := &RedisDiscoverySrv{
		client: client,
		ctx:    ctx,
		logger: zap.NewExample(),
	}

	type args struct {
		node Node
	}

	tests := []struct {
		name        string
		args        args
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successfully add a node",
			args: args{
				node: Node{
					ServerID:          "server_id",
					IP:                "127.0.0.1",
					Hostmane:          "my_machine",
					Port:              3000,
					ConnectedToClient: false,
					Address:           "127.0.0.1:3000",
				},
			},
			wantErr: false,
		},
		{
			name: "error adding a node",
			args: args{
				node: Node{
					ServerID:          "server_id",
					IP:                "127.0.0.1",
					Hostmane:          "my_machine",
					Port:              3000,
					ConnectedToClient: false,
					Address:           "127.0.0.1:3000",
				},
			},
			wantErr:     true,
			expectedErr: fmt.Errorf("error from redis"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			n, err := json.Marshal(tt.args.node)
			require.Nil(t, err)

			expectedInt := mock.ExpectHSet(discoveryKey, tt.args.node.Address, string(n))

			if tt.wantErr {
				expectedInt.SetErr(tt.expectedErr)
			} else {
				expectedInt.SetVal(1)
			}

			err = srv.AddNode(tt.args.node)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestRedisDiscoverySrv_GetNodes(t *testing.T) {
	var ctx = context.TODO()
	client, mock := redismock.NewClientMock()

	srv := &RedisDiscoverySrv{
		client: client,
		ctx:    ctx,
		logger: zap.NewExample(),
	}
	tests := []struct {
		name        string
		want        []Node
		wantErr     bool
		expectedMap map[string]string
		expectedErr error
	}{
		{
			name: "successfully retieve a list of nodes",
			want: []Node{
				{
					ServerID:          "server_id",
					IP:                "127.0.0.1",
					Hostmane:          "my_machine",
					Port:              3000,
					ConnectedToClient: false,
					Address:           "127.0.0.1:3000",
				},
			},
			expectedMap: map[string]string{
				"127.0.0.1:3000": "{\"server_id\":\"server_id\",\"ip\":\"127.0.0.1\",\"hostmane\":\"my_machine\",\"port\":3000,\"created_at\":\"0001-01-01T00:00:00Z\",\"address\":\"127.0.0.1:3000\"}",
			},
			wantErr: false,
		},
		{
			name:        "empty nodes",
			want:        nil,
			wantErr:     true,
			expectedMap: nil,
			expectedErr: redis.Nil,
		},
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			resultMap := mock.ExpectHGetAll(discoveryKey)
			if tt.wantErr {
				resultMap.SetErr(tt.expectedErr)
			}
			resultMap.SetVal(tt.expectedMap)
			nodes, err := srv.GetNodes()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, tt.want, nodes)

		})
	}
}

func TestRedisDiscoverySrv_RemoveDeadNode(t *testing.T) {
	var ctx = context.TODO()
	client, mock := redismock.NewClientMock()

	srv := &RedisDiscoverySrv{
		client: client,
		ctx:    ctx,
		logger: zap.NewExample(),
	}

	type args struct {
		n Node
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		expectedErr error
	}{
		{
			name: "successfully delete a desconnected node",
			args: args{n: Node{
				Address: "127.0.0.1:5000",
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedCmd := mock.ExpectHDel(discoveryKey, tt.args.n.Address)
			if tt.wantErr {
				expectedCmd.SetErr(tt.expectedErr)
			} else {
				expectedCmd.SetVal(1)
			}
			err := srv.RemoveDeadNode(tt.args.n)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.Nil(t, err)
			}

		})
	}
}
