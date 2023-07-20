package utils

import (
	"github.com/galihfebrizki/dbo-api/config"

	"github.com/bwmarrin/snowflake"
)

var nodeOrder *snowflake.Node
var nodeOrderItem *snowflake.Node
var nodeUser *snowflake.Node

// InitSnowflakeOrder initiate Snowflake node singleton.
func InitSnowflakeOrder() error {
	var err error

	// Get node number from env
	nodeNo := config.Get().Snowflake.Order
	if nodeNo > 0 {
		// Create snowflake node
		n, err := snowflake.NewNode(nodeNo)
		if err != nil {
			return err
		}
		// Set node
		nodeOrder = n
	}

	if nodeOrder == nil {
		return err
	}

	return nil
}

// GenerateSnowflakeOrder generate Snowflake ID
func GenerateSnowflakeOrder() string {
	return nodeOrder.Generate().String()
}

// InitSnowflakeOrderItem initiate Snowflake node singleton.
func InitSnowflakeOrderItem() error {
	var err error

	// Get node number from env
	nodeNo := config.Get().Snowflake.OrderItem
	if nodeNo > 0 {
		// Create snowflake node
		n, err := snowflake.NewNode(nodeNo)
		if err != nil {
			return err
		}
		// Set node
		nodeOrderItem = n
	}

	if nodeOrderItem == nil {
		return err
	}

	return nil
}

// GenerateSnowflakeOrder generate Snowflake ID
func GenerateSnowflakeOrderItem() string {
	return nodeOrderItem.Generate().String()
}

// InitSnowflakeUser initiate Snowflake node singleton.
func InitSnowflakeUser() error {
	var err error

	// Get node number from env
	nodeNo := config.Get().Snowflake.User
	if nodeNo > 0 {
		// Create snowflake node
		n, err := snowflake.NewNode(nodeNo)
		if err != nil {
			return err
		}
		// Set node
		nodeUser = n
	}

	if nodeUser == nil {
		return err
	}

	return nil
}

// GenerateSnowflakeOrder generate Snowflake ID
func GenerateSnowflakeUser() string {
	return nodeUser.Generate().String()
}
