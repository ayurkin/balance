package tests

import (
	"balance/internal/adapters/postgres"
	"balance/internal/domain/balance"
	e "balance/internal/domain/errors"
	"balance/internal/domain/models"
	"balance/internal/ports"
	"balance/internal/utils"
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName = "app"
	dbUser = "app"
	dbPass = "secret"
)

func TestBalanceRun(t *testing.T) {
	suite.Run(t, new(ApproveSuite))
}

type ApproveSuite struct {
	suite.Suite
	pgContainer testcontainers.Container
	balance     ports.BalancePort
}

func (suite *ApproveSuite) SetupSuite() {
	ctx := context.Background()

	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:14",
			ExposedPorts: []string{"5432"},
			Env: map[string]string{
				"POSTGRES_DB":       dbName,
				"POSTGRES_USER":     dbUser,
				"POSTGRES_PASSWORD": dbPass,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections"),
			SkipReaper: true,
			AutoRemove: true,
		},
		Started: true,
	})
	suite.Require().NoError(err)

	// with a second delay migrations work properly
	time.Sleep(time.Second * 10)

	ip, err := dbContainer.Host(ctx)
	suite.Require().NoError(err)
	port, err := dbContainer.MappedPort(ctx, "5432")
	suite.T().Log(fmt.Sprintf("Postgres container port: %v", port))
	suite.Require().NoError(err)

	cfg := &pgx.ConnConfig{
		Config: pgconn.Config{
			Host:     ip,
			Port:     uint16(port.Int()),
			Database: dbName,
			User:     dbUser,
			Password: dbPass,
		},
	}

	connString := fmt.Sprintf(`postgres://%s:%s@%s:%d/%s?sslmode=%s`,
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		"disable",
	)
	err = utils.ApplyMigrations(connString, "../../db/changelog/")
	suite.T().Log("Migrations finished")
	suite.Require().NoError(err)

	db, err := postgres.New(ctx, connString)

	suite.Require().NoError(err)

	logger, _ := zap.NewProduction()
	balanceS := balance.New(db, logger.Sugar())
	suite.balance = balanceS
	suite.pgContainer = dbContainer

	suite.T().Log("Suite setup is done")
	time.Sleep(time.Second * 5)
}

func (suite *ApproveSuite) TearDownSuite() {
	err := suite.pgContainer.Terminate(context.Background())
	if err != nil {
		suite.T().Error("Terminate container failed")
	}
	suite.T().Log("Suite stop is done")
}

func (suite *ApproveSuite) Test1Income() {
	ctx := context.Background()

	incomeValue := decimal.NewFromFloat32(10.15)
	userId := int64(1)

	income := models.BalanceWithDesc{UserId: userId, Value: incomeValue, Description: "salary"}
	err := suite.balance.AddIncome(ctx, income)
	suite.Require().NoError(err)

	userBalance, err := suite.balance.GetBalance(ctx, userId)
	suite.Require().NoError(err)

	userBalanceExpected := models.Balance{UserId: userId, Value: userBalance.Value}

	a := assert.New(suite.T())
	a.EqualValues(userBalance, userBalanceExpected)
}

func (suite *ApproveSuite) Test2Expense() {
	ctx := context.Background()

	incomeValue := decimal.NewFromFloat32(10.15)
	userId := int64(2)

	income := models.BalanceWithDesc{UserId: userId, Value: incomeValue, Description: "salary"}
	err := suite.balance.AddIncome(ctx, income)
	suite.Require().NoError(err)

	expenseValue := decimal.NewFromFloat32(2.15)

	expense := models.BalanceWithDesc{UserId: userId, Value: expenseValue, Description: "cinema"}
	err = suite.balance.AddExpense(ctx, expense)
	suite.Require().NoError(err)

	userBalance, err := suite.balance.GetBalance(ctx, userId)
	suite.Require().NoError(err)

	userBalanceExpected := models.Balance{UserId: userBalance.UserId, Value: incomeValue.Sub(expenseValue)}

	a := assert.New(suite.T())
	a.EqualValues(userBalance, userBalanceExpected)
}

func (suite *ApproveSuite) Test3Transfer() {
	ctx := context.Background()

	incomeValue := decimal.NewFromFloat32(10.15)
	transferValue := decimal.NewFromFloat32(2.15)
	userIdFrom := int64(3)
	userIdTo := userIdFrom + 1
	transfer := models.Transaction{
		UserIdFrom:  userIdFrom,
		UserIdTo:    userIdTo,
		Value:       transferValue,
		Time:        time.Now(),
		Description: "credit",
	}

	income := models.BalanceWithDesc{UserId: userIdFrom, Value: incomeValue, Description: "salary"}
	err := suite.balance.AddIncome(ctx, income)
	suite.Require().NoError(err)

	err = suite.balance.DoTransfer(ctx, transfer)
	suite.Require().NoError(err)

	userBalanceFrom, err := suite.balance.GetBalance(ctx, userIdFrom)
	suite.Require().NoError(err)

	userBalanceTo, err := suite.balance.GetBalance(ctx, userIdTo)
	suite.Require().NoError(err)

	userBalanceFromExpected := models.Balance{UserId: userIdFrom, Value: incomeValue.Sub(transferValue)}
	userBalanceToExpected := models.Balance{UserId: userIdTo, Value: transferValue}

	a := assert.New(suite.T())
	a.EqualValues(userBalanceFrom, userBalanceFromExpected)
	a.EqualValues(userBalanceTo, userBalanceToExpected)
}

func (suite *ApproveSuite) Test4ExpenseUnknownUserId() {
	ctx := context.Background()

	userId := int64(5)
	expenseValue := decimal.NewFromFloat32(2.15)

	expense := models.BalanceWithDesc{UserId: userId, Value: expenseValue, Description: "cinema"}
	err := suite.balance.AddExpense(ctx, expense)

	a := assert.New(suite.T())
	a.EqualValues(errors.Is(err, e.UnknownUserIdError), true)
}

func (suite *ApproveSuite) Test5ExpenseNotEnoughBalance() {
	ctx := context.Background()

	userId := int64(6)
	expenseValue := decimal.NewFromFloat32(2.15)
	incomeValue := decimal.NewFromFloat32(1.0)

	income := models.BalanceWithDesc{UserId: userId, Value: incomeValue, Description: "salary"}
	err := suite.balance.AddIncome(ctx, income)
	suite.Require().NoError(err)

	expense := models.BalanceWithDesc{UserId: userId, Value: expenseValue, Description: "cinema"}
	err = suite.balance.AddExpense(ctx, expense)

	a := assert.New(suite.T())
	a.EqualValues(errors.Is(err, e.NotEnoughUserBalanceError), true)
}

func (suite *ApproveSuite) Test6TransferUnknownUserId() {
	ctx := context.Background()

	transferValue := decimal.NewFromFloat32(2.15)
	userIdFrom := int64(7)
	userIdTo := userIdFrom + 1
	transfer := models.Transaction{
		UserIdFrom:  userIdFrom,
		UserIdTo:    userIdTo,
		Value:       transferValue,
		Time:        time.Now(),
		Description: "credit",
	}
	err := suite.balance.DoTransfer(ctx, transfer)

	a := assert.New(suite.T())
	a.EqualValues(errors.Is(err, e.UnknownUserIdError), true)
}

func (suite *ApproveSuite) Test7TransferNotEnoughBalance() {
	ctx := context.Background()

	incomeValue := decimal.NewFromFloat32(2.15)
	transferValue := decimal.NewFromFloat32(10.15)
	userIdFrom := int64(9)
	userIdTo := userIdFrom + 1
	transfer := models.Transaction{
		UserIdFrom:  userIdFrom,
		UserIdTo:    userIdTo,
		Value:       transferValue,
		Time:        time.Now(),
		Description: "credit",
	}

	income := models.BalanceWithDesc{UserId: userIdFrom, Value: incomeValue, Description: "salary"}
	err := suite.balance.AddIncome(ctx, income)
	suite.Require().NoError(err)

	err = suite.balance.DoTransfer(ctx, transfer)
	a := assert.New(suite.T())
	a.EqualValues(errors.Is(err, e.NotEnoughUserBalanceError), true)
}
