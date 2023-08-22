package storage

import (
	"context"
	"github.com/sgladkov/tortuga/internal/models"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var testStorage Storage

func compareProjects(t *testing.T, p1 models.Project, p2 models.Project) {
	require.Equal(t, p1.Id, p2.Id)
	require.Equal(t, p1.Title, p2.Title)
	require.Equal(t, p1.Description, p2.Description)
	require.Equal(t, p1.Tags, p2.Tags)
	require.Equal(t, p1.Created.Round(time.Second), p2.Created.Round(time.Second))
	require.Equal(t, p1.Status, p2.Status)
	require.Equal(t, p1.Owner, p2.Owner)
	require.Equal(t, p1.Contractor, p2.Contractor)
	require.Equal(t, p1.Started.Round(time.Second), p2.Started.Round(time.Second))
	require.Equal(t, p1.Deadline, p2.Deadline)
	require.Equal(t, p1.Price, p2.Price)
}

func testUsers(t *testing.T) {
	ctx := context.Background()
	require.NotNil(t, testStorage)
	ul, err := testStorage.GetUserList(ctx)
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
	require.NoError(t, testStorage.CreateUser(ctx, u))
	ul, err = testStorage.GetUserList(ctx)
	require.NoError(t, err)
	require.Len(t, ul, 1)
	require.True(t, u.Equal(ul[0]))
	u2, err := testStorage.GetUser(ctx, u.Id)
	require.NoError(t, err)
	require.True(t, u.Equal(u2))
	require.Error(t, testStorage.CreateUser(ctx, u))
	_, err = testStorage.GetUser(ctx, "wrong")
	require.Error(t, err)
	u.Id = "wrong"
	u.Nickname = "test2"
	require.Error(t, testStorage.UpdateUser(ctx, u))
	u.Id = "test"
	require.NoError(t, testStorage.UpdateUser(ctx, u))
	u2, err = testStorage.GetUser(ctx, u.Id)
	require.NoError(t, err)
	require.True(t, u.Equal(u2))
	require.Error(t, testStorage.DeleteUser(ctx, "wrong"))
	require.NoError(t, testStorage.DeleteUser(ctx, "test"))
	_, err = testStorage.GetUser(ctx, "test")
	require.Error(t, err)
}

func testProjects(t *testing.T) {
	ctx := context.Background()
	require.NotNil(t, testStorage)
	pl, err := testStorage.GetProjectList(ctx)
	require.NoError(t, err)
	require.Len(t, pl, 0)
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
	require.NoError(t, testStorage.CreateUser(ctx, u))
	p := models.Project{
		Title:       "test project",
		Description: "some project description",
		Tags:        models.Tags{"tag1", "tag2"},
		Created:     time.Now(),
		Status:      models.Open,
		Owner:       "wrong",
		Deadline:    time.Hour * 24 * 14,
		Price:       10000000000,
	}
	_, err = testStorage.CreateProject(ctx, p)
	require.Error(t, err)
	p.Owner = "test"
	pid, err := testStorage.CreateProject(ctx, p)
	require.NoError(t, err)
	p.Id = pid
	p2, err := testStorage.GetProject(ctx, pid)
	require.NoError(t, err)
	compareProjects(t, p, p2)
	pl, err = testStorage.GetProjectList(ctx)
	require.NoError(t, err)
	require.Len(t, pl, 1)
	compareProjects(t, p, pl[0])
	pl, err = testStorage.GetUserProjects(ctx, u.Id)
	require.NoError(t, err)
	require.Len(t, pl, 1)
	compareProjects(t, p, pl[0])
	p.Id = pid + 1
	p.Description = "test2"
	require.Error(t, testStorage.UpdateProject(ctx, p))
	p.Id = pid
	require.NoError(t, testStorage.UpdateProject(ctx, p))
	p2, err = testStorage.GetProject(ctx, pid)
	require.NoError(t, err)
	compareProjects(t, p, p2)
	require.Error(t, testStorage.DeleteProject(ctx, 100))
	require.NoError(t, testStorage.DeleteProject(ctx, pid))
	_, err = testStorage.GetProject(ctx, pid)
	require.Error(t, err)
	require.NoError(t, testStorage.DeleteUser(ctx, "test"))
}

func testBids(t *testing.T) {
	ctx := context.Background()
	require.NotNil(t, testStorage)
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
	require.NoError(t, testStorage.CreateUser(ctx, u))
	p := models.Project{
		Title:       "test project",
		Description: "some project description",
		Tags:        models.Tags{"tag1", "tag2"},
		Created:     time.Now(),
		Status:      models.Open,
		Owner:       "test",
		Deadline:    time.Hour * 24 * 14,
		Price:       10000000000,
	}
	pid, err := testStorage.CreateProject(ctx, p)
	require.NoError(t, err)
	p.Id = pid
	bl, err := testStorage.GetProjectBids(ctx, pid)
	require.NoError(t, err)
	require.Len(t, bl, 0)
	bid := models.Bid{
		Project:  pid,
		User:     u.Id,
		Deadline: time.Hour * 24 * 13,
		Price:    100000,
	}
	bidid, err := testStorage.CreateBid(ctx, bid)
	require.NoError(t, err)
	bid.Id = bidid
	b2, err := testStorage.GetBid(ctx, bidid)
	require.NoError(t, err)
	require.Equal(t, bid, b2)
	_, err = testStorage.GetBid(ctx, bidid+1)
	require.Error(t, err)
	bl, err = testStorage.GetProjectBids(ctx, pid)
	require.NoError(t, err)
	require.Len(t, bl, 1)
	require.Equal(t, bid, bl[0])
	bid.User = "wrong"
	_, err = testStorage.CreateBid(ctx, bid)
	require.Error(t, err)
	bid.User = u.Id
	bid.Project = pid + 1
	_, err = testStorage.CreateBid(ctx, bid)
	require.Error(t, err)

	require.NoError(t, testStorage.DeleteBid(ctx, bidid))
	require.NoError(t, testStorage.DeleteProject(ctx, pid))
	require.NoError(t, testStorage.DeleteUser(ctx, "test"))
}

func testRates(t *testing.T) {
	ctx := context.Background()
	require.NotNil(t, testStorage)
	owner := models.User{
		Id:          "owner",
		Nickname:    "test",
		Description: "some text here",
		Nonce:       0,
		Registered:  time.Now(),
		Status:      0,
		Tags:        models.Tags{"tag1", "tag2"},
		Rating:      0.0,
		Account:     0,
	}
	require.NoError(t, testStorage.CreateUser(ctx, owner))
	contractor := models.User{
		Id:          "contractor",
		Nickname:    "test",
		Description: "some text here",
		Nonce:       0,
		Registered:  time.Now(),
		Status:      0,
		Tags:        models.Tags{"tag1", "tag2"},
		Rating:      0.0,
		Account:     0,
	}
	require.NoError(t, testStorage.CreateUser(ctx, contractor))
	p := models.Project{
		Title:       "test project",
		Description: "some project description",
		Tags:        models.Tags{"tag1", "tag2"},
		Created:     time.Now(),
		Status:      models.Open,
		Owner:       "owner",
		Contractor:  "contractor",
		Deadline:    time.Hour * 24 * 14,
		Price:       10000000000,
	}
	pid, err := testStorage.CreateProject(ctx, p)
	require.NoError(t, err)
	p.Id = pid
	rl, err := testStorage.GetEvaluatedRates(ctx, "contractor")
	require.NoError(t, err)
	require.Len(t, rl, 0)
	rl, err = testStorage.GetEvaluatedRates(ctx, "owner")
	require.NoError(t, err)
	require.Len(t, rl, 0)
	rl, err = testStorage.GetEvaluatorRates(ctx, "contractor")
	require.NoError(t, err)
	require.Len(t, rl, 0)
	rl, err = testStorage.GetEvaluatorRates(ctx, "owner")
	require.NoError(t, err)
	require.Len(t, rl, 0)
	rate1 := models.Rate{
		Project:   pid,
		Evaluator: "owner",
		Evaluated: "wrong",
	}
	_, err = testStorage.CreateRate(ctx, rate1)
	require.Error(t, err)
	rate1.Evaluated = "contractor"
	rate1.Id, err = testStorage.CreateRate(ctx, rate1)
	require.NoError(t, err)
	r, err := testStorage.GetRate(ctx, rate1.Id)
	require.NoError(t, err)
	require.Equal(t, rate1, r)
	_, err = testStorage.GetRate(ctx, rate1.Id+1)
	require.Error(t, err)
	rate2 := models.Rate{
		Project:   pid,
		Evaluator: "contractor",
		Evaluated: "owner",
	}
	rate2.Id, err = testStorage.CreateRate(ctx, rate2)
	require.NoError(t, err)
	rl, err = testStorage.GetEvaluatedRates(ctx, "contractor")
	require.NoError(t, err)
	require.Len(t, rl, 1)
	require.Equal(t, rate1, rl[0])
	rl, err = testStorage.GetEvaluatedRates(ctx, "owner")
	require.NoError(t, err)
	require.Len(t, rl, 1)
	require.Equal(t, rate2, rl[0])
	rl, err = testStorage.GetEvaluatorRates(ctx, "contractor")
	require.NoError(t, err)
	require.Len(t, rl, 1)
	require.Equal(t, rate2, rl[0])
	rl, err = testStorage.GetEvaluatorRates(ctx, "owner")
	require.NoError(t, err)
	require.Len(t, rl, 1)
	require.Equal(t, rate1, rl[0])

	require.NoError(t, testStorage.DeleteRate(ctx, rate1.Id))
	require.NoError(t, testStorage.DeleteRate(ctx, rate2.Id))
	require.NoError(t, testStorage.DeleteProject(ctx, pid))
	require.NoError(t, testStorage.DeleteUser(ctx, "owner"))
	require.NoError(t, testStorage.DeleteUser(ctx, "contractor"))
}

func testTransactions(t *testing.T) {
	ctx := context.Background()
	require.NotNil(t, testStorage)

	require.NoError(t, testStorage.BeginTx())
	u := models.User{
		Id: "test",
	}
	require.NoError(t, testStorage.CreateUser(ctx, u))
	_, err := testStorage.GetUser(ctx, u.Id)
	require.NoError(t, err)
	require.NoError(t, testStorage.RollbackTx())
	_, err = testStorage.GetUser(ctx, u.Id)
	require.Error(t, err)
	require.Error(t, testStorage.CommitTx())

	require.NoError(t, testStorage.BeginTx())
	require.NoError(t, testStorage.CreateUser(ctx, u))
	require.NoError(t, testStorage.CommitTx())
	_, err = testStorage.GetUser(ctx, u.Id)
	require.NoError(t, err)
	require.NoError(t, testStorage.DeleteUser(ctx, "test"))
}
