package storage

import (
	"github.com/sgladkov/tortuga/internal/models"
	"testing"
)

func TestTestStorage_test(t *testing.T) {
	testStorage = NewTestStorage([]models.User{}, []models.Project{}, []models.Bid{}, []models.Rate{})
	t.Run("Users", testUsers)
	t.Run("Projects", testProjects)
	t.Run("Bids", testBids)
	t.Run("Rates", testRates)
	t.Run("Transactions", testTransactions)
}
