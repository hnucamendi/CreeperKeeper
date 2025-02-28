package dynamo

import (
	"context"

	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	cktypes "github.com/hnucamendi/creeper-keeper/types"
)

type DynamoAPI interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
}

type Client struct {
	*dynamodb.Client
}

func (db *Client) RegisterServer(ctx context.Context, tableName string, serverID string, serverType string, serverIP string, serverName string, serverIsRunning bool, serverLastUpdated string) (bool, error) {
	_, err := db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: serverID,
			},
			"SK": &types.AttributeValueMemberS{
				Value: "serverdetails",
			},
			"ServerIP": &types.AttributeValueMemberS{
				Value: serverIP,
			},
			"ServerName": &types.AttributeValueMemberS{
				Value: serverName,
			},
			"LastUpdated": &types.AttributeValueMemberS{
				Value: serverLastUpdated,
			},
			"IsRunning": &types.AttributeValueMemberBOOL{
				Value: serverIsRunning,
			},
		},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *Client) ListServers(ctx context.Context, tableName string) ([]cktypes.Server, error) {
	input := &dynamodb.ScanInput{TableName: aws.String(tableName)}

	out, err := db.Client.Scan(ctx, input)
	if err != nil {
		return nil, err
	}

	var servers []cktypes.Server
	err = attributevalue.UnmarshalListOfMaps(out.Items, &servers)
	if err != nil {
	}

	return servers, nil
}
func (db *Client) ListServer(ctx context.Context, tableName string, serverID string) (*cktypes.Server, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: serverID,
			},
			"SK": &types.AttributeValueMemberS{
				Value: "serverdetails",
			},
		},
	}
	out, err := db.Client.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}

	var server cktypes.Server
	attributevalue.UnmarshalMap(out.Item, &server)

	return &server, nil
}

func (db *Client) UpsertServer(ctx context.Context, tableName string, serverID string, serverIP string, serverName string) (bool, error) {
	zone, err := time.LoadLocation("America/New_York")
	if err != nil {
		return false, err
	}
	lastUpdated := time.Now().In(zone).Format(time.DateTime)

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: serverID,
			},
			"SK": &types.AttributeValueMemberS{
				Value: "serverdetails",
			},
			"ServerIP": &types.AttributeValueMemberS{
				Value: serverIP,
			},
			"ServerName": &types.AttributeValueMemberS{
				Value: serverName,
			},
			"LastUpdated": &types.AttributeValueMemberS{
				Value: lastUpdated,
			},
			"IsRunning": &types.AttributeValueMemberBOOL{
				Value: false,
			},
		},
	}
	_, err = db.Client.PutItem(ctx, input)
	if err != nil {
		return false, err
	}
	return true, nil
}

func NewDatabase() (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: dynamodb.NewFromConfig(cfg),
	}, nil
}
