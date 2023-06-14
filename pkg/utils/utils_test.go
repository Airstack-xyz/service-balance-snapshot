package utils

import (
	"fmt"
	"testing"

	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/test-go/testify/assert"
)

func TestGetTopicName(t *testing.T) {

	topicName := "transfer"

	t.Run("Get Topic Name based on chain ID", func(t *testing.T) {
		chainIds := []string{"1", "137"}
		expectedDBName := []string{topicName, fmt.Sprint("polygon_", topicName)}
		for i, chainId := range chainIds {
			t.Setenv(constants.CHAINID, chainId)
			dbName := GetTopicName(topicName)
			assert.Equal(t, expectedDBName[i], dbName)
		}
	})

	t.Run("Get Topic Name from env", func(t *testing.T) {
		t.Setenv(topicName, "test-topic")
		dbName := GetTopicName(topicName)
		assert.Equal(t, "test-topic", dbName)
	})
}

func TestGetCacheTTL(t *testing.T) {
	t.Run("Default TTL when env not set", func(t *testing.T) {
		ttl := GetCacheTTL()
		assert.Equal(t, constants.DEFAULT_CACHE_TTL, ttl)
	})

	t.Run("Returns Env TTL if it's set", func(t *testing.T) {
		t.Setenv(constants.CACHE_TTL, "100")
		ttl := GetCacheTTL()
		assert.Equal(t, 100, ttl)
	})
}

func TestPtr(t *testing.T) {
	a := 10
	address := Ptr(a)
	assert.Equal(t, &a, address)
}
