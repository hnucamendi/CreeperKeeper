package database

import (
	"context"

	"github.com/hnucamendi/creeper-keeper/service/database/dynamo"
	"github.com/hnucamendi/creeper-keeper/types"
	"github.com/hnucamendi/creeper-keeper/utils"
)

type DBClient string

const (
	DYNAMODB DBClient = "DYNAMODB"
)

type Database interface {
	RegisterServer(ctx context.Context, tableName string, serverID string, serverType string, serverIP string, serverName string, serverIsRunning bool, serverLastUpdated string) (bool, error)
	ListServers(ctx context.Context, tableName string) ([]types.Server, error)
	ListServer(ctx context.Context, tableName string, serverID string) (*types.Server, error)
	UpsertServer(ctx context.Context, tableName string, serverID string, serverIP string, serverName string) (bool, error)
}

type Client struct {
	Client   Database
	Database *string
	Schema   *string
	Table    *string
}

type Opts func(*Client)

func WithName(name string) Opts {
	return func(c *Client) {
		c.Database = utils.String(name)
	}
}

func WithSchema(schema string) Opts {
	return func(c *Client) {
		c.Schema = utils.String(schema)
	}
}

func WithTable(table string) Opts {
	return func(c *Client) {
		c.Table = utils.String(table)
	}
}

func WithClient(db DBClient) Opts {
	return func(c *Client) {
		switch db {
		case DYNAMODB:
			db, err := dynamo.NewDatabase()
			if err != nil {
				c.Client = nil
			}
			c.Client = db
		default:
			c.Client = nil
		}
	}
}

func NewDatabase(fn ...Opts) *Client {
	c := &Client{}
	for _, f := range fn {
		f(c)
	}
	return c
}
