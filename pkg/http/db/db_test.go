package db

import (
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDB(t *testing.T) {
	t.Run("open successfully", func(t *testing.T) {
		assert.NotPanics(t, func() {
			MustOpenDefault(inMemorySqliteCfg)
			CloseDB()
		})
	})

	t.Run("panic on unsupported driver", func(t *testing.T) {
		assert.Panics(t, func() {
			unsupportedDriverCfg := *inMemorySqliteCfg
			unsupportedDriverCfg.Driver = "mysql"
			MustOpenDefault(&unsupportedDriverCfg)
		})
	})

	t.Run("panic on invalid setup test", func(t *testing.T) {
		assert.Panics(t, func() {
			oldValue := inMemorySqliteCfg.Driver
			inMemorySqliteCfg.Driver = "mysql"
			defer func() {
				inMemorySqliteCfg.Driver = oldValue
			}()

			MustSetupTest()
		})
	})
}

func TestSetupTestDB(t *testing.T) {
	assert.NotPanics(t, func() {
		MustSetupTest()
		CloseDB()
	})
}

func TestGetDB(t *testing.T) {

	assert.NotPanics(t, func() {
		MustOpenDefault(inMemorySqliteCfg)

		assert.Equal(t, dbDefault, GetDB())
		CloseDB()
	})

	assert.Panics(t, func() {
		_ = GetDB()
	}, "should panic on uninitialized connection")
}

func TestCloseDB(t *testing.T) {
	MustOpenDefault(inMemorySqliteCfg)
	assert.NotNil(t, dbDefault)

	CloseDB()
	assert.Nil(t, dbDefault)
}

type sampleModel struct {
	ID      int64  `gorm:"PRIMARY_KEY,AUTO_INCREMENT"`
	Message string `gorm:"size:255"`
}

func TestInsertOnSameDB(t *testing.T) {
	maxItemsEachThread := 1000
	threads := 2
	wg := &sync.WaitGroup{}

	// init data
	MustSetupTest()
	err := GetDB().AutoMigrate(&sampleModel{})
	require.NoError(t, err)

	for i := 0; i < threads; i++ {

		wg.Add(1)

		go func(tid int, wg *sync.WaitGroup, numOfItems int) {

			defer wg.Done()
			tag := fmt.Sprintf("thread-%d", tid)

			for j := 0; j < numOfItems; j++ {
				err := GetDB().Create(&sampleModel{Message: fmt.Sprintf("%s - message %d", tag, j)}).Error
				if err != nil {
					log.Print(tag, " error while inserting data: ", err)
				}
			}
		}(i, wg, maxItemsEachThread)
	}

	// wait for all job done
	wg.Wait()

	// check insert data
	var count int64
	GetDB().Model(&sampleModel{}).Count(&count)
	assert.Equal(t, int64(maxItemsEachThread*threads), count)
}

func TestConfigShouldBuildDSN(t *testing.T) {
	c := Config{
		DSN:  "",
		Host: "localhost",
		Port: "5432",
		User: "test",
		Pass: "test",
		Name: "db_test",
	}
	assert.Equal(t, "host=localhost port=5432 user=test dbname=db_test password=test sslmode=disable connect_timeout=5", c.GetDSN())

	// if case DSN is set
	c.DSN = "host=server.com port=5432 user=dev dbname=db_dev password=dev sslmode=disable"
	assert.Equal(t, "host=server.com port=5432 user=dev dbname=db_dev password=dev sslmode=disable", c.GetDSN())
}
