package service

import (
	"context"
	"github.com/sgladkov/tortuga/internal/models"
	"github.com/sgladkov/tortuga/internal/storage"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func prepareMarketplace() *Marketplace {
	testStorage := storage.NewTestStorage([]models.User{}, []models.Project{}, []models.Bid{}, []models.Rate{})
	return NewMarketplace(testStorage)
}

func TestMarketplace_Users(t *testing.T) {
	marketplace := prepareMarketplace()
	ctx := context.Background()
	require.NotNil(t, marketplace)

	ul, err := marketplace.GetUserList(ctx)
	require.NoError(t, err)
	require.Len(t, ul, 0)
	u := models.User{
		Id:          "test",
		Nickname:    "test",
		Description: "some text here",
		Nonce:       0,
		Registered:  time.Now(),
		Status:      0,
		Tags:        models.Tags{"tag1", "tag2"},
		Rating:      0.0,
		Account:     0,
	}
	require.Error(t, marketplace.AddUser(ctx, "wrong", u))
	require.NoError(t, marketplace.AddUser(ctx, "test", u))
	ul, err = marketplace.GetUserList(ctx)
	require.NoError(t, err)
	require.Len(t, ul, 1)
	require.True(t, u.Equal(ul[0]))
	u2, err := marketplace.GetUser(ctx, u.Id)
	require.NoError(t, err)
	require.True(t, u.Equal(u2))
	require.Error(t, marketplace.AddUser(ctx, "test", u))
	_, err = marketplace.GetUser(ctx, "wrong")
	require.Error(t, err)
	nonce := u.Nonce + 1
	require.Error(t, marketplace.UpdateUserNonce(ctx, "wrong", nonce))
	require.NoError(t, marketplace.UpdateUserNonce(ctx, "test", nonce))
	u2, err = marketplace.GetUser(ctx, u.Id)
	require.NoError(t, err)
	require.Equal(t, nonce, u2.Nonce)
}

func TestMarketplace_Projects(t *testing.T) {
	marketplace := prepareMarketplace()
	ctx := context.Background()
	require.NotNil(t, marketplace)

	owner := models.User{
		Id:          "owner",
		Nickname:    "owner",
		Description: "some text here",
		Nonce:       0,
		Registered:  time.Now(),
		Status:      0,
		Tags:        models.Tags{"tag1", "tag2"},
		Rating:      0.0,
		Account:     0,
	}
	require.NoError(t, marketplace.AddUser(ctx, "owner", owner))
	contractor := models.User{
		Id:          "contractor",
		Nickname:    "contractor",
		Description: "some text here",
		Nonce:       0,
		Registered:  time.Now(),
		Status:      0,
		Tags:        models.Tags{"tag1", "tag2"},
		Rating:      0.0,
		Account:     0,
	}
	require.NoError(t, marketplace.AddUser(ctx, "contractor", contractor))

	pid, err := marketplace.CreateProject(ctx, "owner", "some project", "some description",
		models.Tags{}, "owner", time.Hour, 1000000)
	require.NoError(t, err)
	err = marketplace.DeleteProject(ctx, "contractor", pid)
	require.Error(t, err)
	err = marketplace.DeleteProject(ctx, "owner", pid)
	require.NoError(t, err)

	pid, err = marketplace.CreateProject(ctx, "owner", "some project", "some description",
		models.Tags{}, "owner", time.Hour, 1000000)
	require.NoError(t, err)

	bidid, err := marketplace.CreateBid(ctx, "contractor", pid, "contractor", 100000, time.Hour, "message")
	require.NoError(t, err)

	_, err = marketplace.AcceptBid(ctx, "contractor", bidid)
	require.Error(t, err)

	pid2, err := marketplace.AcceptBid(ctx, "owner", bidid)
	require.NoError(t, err)
	require.Equal(t, pid, pid2)

	_, err = marketplace.GetBid(ctx, bidid)
	require.Error(t, err)

	err = marketplace.DeleteProject(ctx, "owner", pid)
	require.Error(t, err)

	err = marketplace.UpdateProject(ctx, "owner", pid, "test", "test", models.Tags{},
		time.Second, 1)
	require.Error(t, err)

	err = marketplace.SetProjectReady(ctx, "owner", pid)
	require.Error(t, err)
	err = marketplace.SetProjectReady(ctx, "contractor", pid)
	require.NoError(t, err)

	err = marketplace.AcceptProject(ctx, "contractor", pid)
	require.Error(t, err)
	err = marketplace.AcceptProject(ctx, "owner", pid)
	require.NoError(t, err)
}
