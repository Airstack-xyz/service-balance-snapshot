package repository

import (
	"fmt"
	"testing"

	"github.com/airstack-xyz/service-balance-snapshot/pkg/constants"
	"github.com/stretchr/testify/assert"
)

func TestGetDB(t *testing.T) {

	db := "TOKENDB"

	t.Run("Get DB Name based on chain ID", func(t *testing.T) {
		chainIds := []string{"1", "137"}
		expectedDBName := []string{db, fmt.Sprint("POLYGON_", db)}
		for i, chainId := range chainIds {
			t.Setenv(constants.CHAINID, chainId)
			dbName := getDB(db)
			assert.Equal(t, expectedDBName[i], dbName)
		}
	})
}
