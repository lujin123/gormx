package gormx

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

type User struct {
	Id       int64
	Nickname string
	Age      int64
}

func (User) TableName() string {
	return "test_users"
}

type StoreTestSuite struct {
	suite.Suite

	db *Gormx
}

func (suite *StoreTestSuite) SetupTest() {
	conf := &Config{
		Dialector:   nil, //fill driver
		MaxIdleConn: 10,
		MaxOpenConn: 10,
		MaxLifetime: 1000,
		Debug:       false,
	}
	db, err := New(conf)
	suite.Assert().Nil(err)
	suite.db = db
	suite.initUsers()
}

func (suite *StoreTestSuite) TearDownTest() {
	suite.db.Exec("drop table test_users;")
}

func (suite *StoreTestSuite) initUsers() {
	suite.db.Exec("create table test_users (id serial primary key not null, nickname varchar(64) not null, age integer default 0);")
	var users []User
	for i := 0; i < 2; i++ {
		users = append(users, User{
			Nickname: fmt.Sprintf("hello %d", i),
			Age:      int64(i),
		})
	}

	suite.Assert().Nil(suite.db.Insert(users))
}

func (suite *StoreTestSuite) TestFindOne() {
	var user = User{
		Id: 1,
	}
	err := suite.db.FindOne(&user)
	if suite.Assert().Nil(err) {
		suite.Equal("hello 0", user.Nickname)
	}
}

func (suite *StoreTestSuite) TestExists() {
	var (
		exists bool
		err    error
	)
	exists, err = suite.db.Exists(&User{
		Id: 1,
	})
	if suite.Assert().Nil(err) {
		suite.Assert().Equal(true, exists)
	}

	exists, err = suite.db.Exists(&User{
		Id: -1,
	})
	if suite.Assert().Nil(err) {
		suite.Assert().Equal(false, exists)
	}
}

func (suite *StoreTestSuite) TestInsert() {
	user := User{
		Nickname: "hello",
	}
	err := suite.db.Insert(&user)
	if suite.Assert().Nil(err) {
		suite.EqualValues(3, user.Id)
	}
}

func (suite *StoreTestSuite) TestFindMany() {
	var user []User
	err := suite.db.FindMany(&user, Pagination(1, 2))
	if !suite.Assert().Nil(err) {
		return
	}

	suite.Equal(2, len(user))
}

func (suite *StoreTestSuite) TestPluck() {
	var ids []int64
	err := suite.db.Model(&User{}).Pluck("id", &ids, Pagination(1, 2))
	if !suite.Assert().Nil(err) {
		return
	}

	suite.Equal([]int64{1, 2}, ids)
}

func (suite *StoreTestSuite) TestCount() {
	total, err := suite.db.Model(&User{}).Count()
	if !suite.Assert().Nil(err) {
		return
	}

	suite.EqualValues(2, total)
}

func (suite *StoreTestSuite) TestUpdate() {
	err := suite.db.Model(&User{Id: -1}).Update("nickname", "hello world")
	suite.ErrorIs(err, ErrNoRowsAffected)

	err = suite.db.Model(&User{Id: 1}).Update("nickname", "hello world")
	if suite.Assert().Nil(err) {
		var user User
		err = suite.db.FindOne(&user, WithId(1))
		if suite.Assert().Nil(err) {
			suite.Equal("hello world", user.Nickname)
		}
	}
}

func (suite *StoreTestSuite) TestUpdates() {
	err := suite.db.Updates(&User{
		Id:       1,
		Nickname: "hello struct",
		Age:      133,
	})
	if suite.Assert().Nil(err) {
		var user User
		err = suite.db.FindOne(&user, WithId(1))
		if suite.Assert().Nil(err) {
			suite.EqualValues(&User{
				Id:       1,
				Nickname: "hello struct",
				Age:      133,
			}, &user)
		}
	}

	err = suite.db.Model(&User{Id: 1}).Updates(map[string]interface{}{
		"nickname": "hello map",
		"age":      "133",
	})
	if suite.Assert().Nil(err) {
		user := User{
			Id: 1,
		}
		err = suite.db.FindOne(&user)
		if suite.Assert().Nil(err) {
			suite.EqualValues(&User{
				Id:       1,
				Nickname: "hello map",
				Age:      133,
			}, &user)
		}
	}

	err = suite.db.Updates(&User{
		Id:       -1,
		Nickname: "hello no rows",
		Age:      133,
	})
	suite.Assert().ErrorIs(err, ErrNoRowsAffected)
}

func (suite *StoreTestSuite) TestDelete() {
	user := User{Id: 1}
	err := suite.db.Delete(&user)
	if suite.Nil(err) {
		err = suite.db.FindOne(&user)
		suite.ErrorIs(err, gorm.ErrRecordNotFound)
	}
}
func (suite *StoreTestSuite) TestRaw() {
	var (
		query = `select * from test_users where id=?`
		args  = []interface{}{1}
		user  User
		err   error
	)

	err = suite.db.Raw(query, args...).FindOne(&user)
	if suite.Assert().Nil(err) {
		suite.EqualValues(&User{
			Id:       1,
			Nickname: "hello 0",
			Age:      0,
		}, &user)
	}

	var myUser struct {
		Id       int64
		Nickname string
	}
	err = suite.db.Raw(query, args...).Scan(&myUser)
	if suite.Assert().Nil(err) {
		suite.Equal(myUser.Nickname, "hello 0")
	}
}

func (suite *StoreTestSuite) TestTx() {
	var id int64
	err := suite.db.WithContext(context.Background()).Tx(func(tx *Gormx) error {
		if err := tx.Model(&User{Id: 1}).Update("nickname", "hello tx"); err != nil {
			return err
		}
		user := User{
			Nickname: "hello tx insert",
			Age:      133,
		}
		if err := tx.Insert(&user); err != nil {
			return err
		}
		id = user.Id
		return nil
	})
	if suite.Nil(err) {
		user := User{
			Id: id,
		}
		err = suite.db.FindOne(&user)
		if suite.Nil(err) {
			suite.EqualValues(&User{
				Id:       id,
				Nickname: "hello tx insert",
				Age:      133,
			}, &user)
		}
	}
}
