package gobatis_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/GoBatis/tests"
)

func TestInit(t *testing.T) {
	defer gobatis.ClearInit()
	exceptederr := errors.New("init error")
	gobatis.Init(func(ctx *gobatis.InitContext) error {
		return exceptederr
	})

	_, err := gobatis.New(&gobatis.Config{DriverName: tests.TestDrv,
		DataSource: tests.TestConnURL,
		XMLPaths: []string{"example/test.xml",
			"../example/test.xml",
			"../../example/test.xml"}})
	if err == nil {
		t.Error("excepted error got ok")
		return
	}

	if !strings.Contains(err.Error(), exceptederr.Error()) {
		t.Error("excepted contains", exceptederr.Error())
		t.Error("got", err.Error())
		return
	}
}

func TestSession(t *testing.T) {
	tests.Run(t, func(_ testing.TB, factory *gobatis.SessionFactory) {
		insertUser := tests.User{
			Name:        "张三",
			Nickname:    "haha",
			Password:    "password",
			Description: "地球人",
			Address:     "沪南路1155号",
			Sex:         "女",
			ContactInfo: `{"QQ":"8888888"}`,
			Birth:       time.Now(),
			CreateTime:  time.Now(),
		}

		user := tests.User{
			Name: "张三",
		}

		t.Run("statementNotFound", func(t *testing.T) {
			_, err := factory.Insert("insertUserNotExists", insertUser)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "insertUserNotExists") {
				t.Error("excepted is insertUserNotExists")
				t.Error("actual   is", err)
			}
		})

		t.Run("statementTypeError", func(t *testing.T) {
			_, err := factory.Insert("selectUser", insertUser)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUser") {
				t.Error("excepted is selectUser")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "type Error") {
				t.Error("excepted is type Error")
				t.Error("actual   is", err)
			}
		})

		t.Run("argumentError", func(t *testing.T) {
			var u int
			err := factory.SelectOne("selectUserTplError").Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUserTplError") {
				t.Error("excepted is selectUserTplError")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "arguments is missing") {
				t.Error("excepted is arguments is missing")
				t.Error("actual   is", err)
			}

			err = factory.SelectOne("selectUserTplError", 1, 2, 3).Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUserTplError") {
				t.Error("excepted is selectUserTplError")
				t.Error("actual   is", err)
			}

			if !strings.Contains(err.Error(), "arguments is exceed 1") {
				t.Error("excepted is arguments is exceed 1")
				t.Error("actual   is", err)
			}
		})

		t.Run("bindError", func(t *testing.T) {
			var u int
			err := factory.SelectOne("selectUser", struct{ s string }{"abc"}).Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUser") {
				t.Error("excepted is selectUser")
				t.Error("actual   is", err)
			}
		})

		t.Run("compileSQLFailAfterTplOk", func(t *testing.T) {
			var u int
			err := factory.SelectOne("selectUserTplError", map[string]interface{}{"id": "abc"}).Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUserTplError") {
				t.Error("excepted is selectUserTplError")
				t.Error("actual   is", err)
			}
		})

		t.Run("bindErrorAfterTplOk", func(t *testing.T) {
			var u int
			err := factory.SelectOne("selectUserTpl3", map[string]interface{}{"id": "abc"}).Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}
			if !strings.Contains(err.Error(), "selectUserTpl3") {
				t.Error("excepted is selectUserTpl3")
				t.Error("actual   is", err)
			}
			if !strings.Contains(err.Error(), "'name'") {
				t.Error("excepted is 'name'")
				t.Error("actual   is", err)
			}
		})
		t.Run("selectUsers", func(t *testing.T) {
			if _, err := factory.DB().Exec(`DELETE FROM gobatis_users`); err != nil {
				t.Error(err)
				return
			}

			_, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}
			_, err = factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			var users []tests.User
			err = factory.Select("selectUsers", user).ScanSlice(&users)
			if err != nil {
				t.Error(err)
				return
			}

			if len(users) != 2 {
				t.Error("excepted size is", 2)
				t.Error("actual size   is", len(users))
				return
			}

			insertUser2 := insertUser
			insertUser2.Birth = insertUser2.Birth.UTC()
			insertUser2.CreateTime = insertUser2.CreateTime.UTC()

			for _, u := range users {
				insertUser2.ID = u.ID
				u.Birth = u.Birth.UTC()
				u.CreateTime = u.CreateTime.UTC()

				tests.AssertUser(t, insertUser2, u)
			}

			results := factory.Reference().Select("selectUsers",
				[]string{"name"},
				[]interface{}{user.Name})
			if results.Err() != nil {
				t.Error(results.Err())
				return
			}
			defer results.Close()

			users = nil
			for results.Next() {
				var u tests.User
				err = results.Scan(&u)
				if err != nil {
					t.Error(err)
					return
				}
				users = append(users, u)
			}

			if results.Err() != nil {
				t.Error(results.Err())
				return
			}

			if len(users) != 2 {
				t.Error("excepted size is", 2)
				t.Error("actual size   is", len(users))
				return
			}

			for _, u := range users {

				insertUser2.ID = u.ID
				u.Birth = u.Birth.UTC()
				u.CreateTime = u.CreateTime.UTC()

				tests.AssertUser(t, insertUser2, u)
			}
		})

		t.Run("selectUser", func(t *testing.T) {
			if _, err := factory.DB().Exec(`DELETE FROM gobatis_users`); err != nil {
				t.Error(err)
				return
			}

			id, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			u := tests.User{Name: insertUser.Name + "abc"}
			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err == nil {
				t.Error("excepted error but got ok")
				return
			}

			if !strings.Contains(err.Error(), sql.ErrNoRows.Error()) {
				t.Error("excepted is", sql.ErrNoRows)
				t.Error("actual   is", err)
			}

			u = tests.User{Name: insertUser.Name}
			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err != nil {
				t.Error(err)
				return
			}

			insertUser.ID = u.ID
			insertUser.Birth = insertUser.Birth.UTC()
			insertUser.CreateTime = insertUser.CreateTime.UTC()
			u.Birth = u.Birth.UTC()
			u.CreateTime = u.CreateTime.UTC()

			tests.AssertUser(t, insertUser, u)

			u2 := tests.User{}
			err = factory.SelectOne("selectUser", map[string]interface{}{"name": insertUser.Name}).
				Scan(&u2)
			if err != nil {
				t.Error(err)
				return
			}

			insertUser.ID = u2.ID
			insertUser.Birth = insertUser.Birth.UTC()
			insertUser.CreateTime = insertUser.CreateTime.UTC()
			u2.Birth = u2.Birth.UTC()
			u2.CreateTime = u2.CreateTime.UTC()

			tests.AssertUser(t, insertUser, u2)

			u2 = tests.User{}
			err = factory.Reference().SelectOne("selectUserTpl", []string{"id"}, []interface{}{id}).
				Scan(&u2)
			if err != nil {
				t.Error(err)
				return
			}

			u2.Birth = u2.Birth.UTC()
			u2.CreateTime = u2.CreateTime.UTC()
			tests.AssertUser(t, insertUser, u2)

			u2 = tests.User{}
			err = factory.Reference().SelectOne("selectUserTpl2", []string{"u"}, []interface{}{&tests.User{ID: id}}).
				Scan(&u2)
			if err != nil {
				t.Error(err)
				return
			}

			u2.Birth = u2.Birth.UTC()
			u2.CreateTime = u2.CreateTime.UTC()
			tests.AssertUser(t, insertUser, u2)

			u2 = tests.User{}
			err = factory.Reference().SelectOne("selectUserTpl3", []string{"id", "name"}, []interface{}{id, insertUser.Name}).
				Scan(&u2)
			if err != nil {
				t.Error(err)
				return
			}

			u2.Birth = u2.Birth.UTC()
			u2.CreateTime = u2.CreateTime.UTC()
			tests.AssertUser(t, insertUser, u2)
		})

		t.Run("updateUser", func(t *testing.T) {
			if _, err := factory.DB().Exec(`DELETE FROM gobatis_users`); err != nil {
				t.Error(err)
				return
			}

			_, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			u := tests.User{Name: insertUser.Name}
			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err != nil {
				t.Error(err)
				return
			}

			updateUser := insertUser
			updateUser.ID = u.ID
			updateUser.Nickname = "test@foxmail.com"
			updateUser.Birth = time.Now()
			updateUser.CreateTime = time.Now()
			_, err = factory.Update("updateUser", updateUser)
			if err != nil {
				t.Error(err)
			}

			updateUser.Birth = updateUser.Birth.UTC()
			updateUser.CreateTime = updateUser.CreateTime.UTC()

			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err != nil {
				t.Error(err)
				return
			}
			u.Birth = u.Birth.UTC()
			u.CreateTime = u.CreateTime.UTC()

			tests.AssertUser(t, updateUser, u)
		})

		t.Run("deleteUser", func(t *testing.T) {
			if _, err := factory.DB().Exec(`DELETE FROM gobatis_users`); err != nil {
				t.Error(err)
				return
			}

			_, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}

			u := tests.User{Name: insertUser.Name}
			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err != nil {
				t.Error(err)
				return
			}

			deleteUser := tests.User{ID: u.ID}
			_, err = factory.Delete("deleteUser", deleteUser)
			if err != nil {
				t.Error(err)
			}

			err = factory.SelectOne("selectUser", u).Scan(&u)
			if err == nil {
				t.Error("DELETE fail")
				return
			}

			if err != sql.ErrNoRows {
				t.Error(err)
			}
		})

		t.Run("deleteUserTpl", func(t *testing.T) {
			if _, err := factory.DB().Exec(`DELETE FROM gobatis_users`); err != nil {
				t.Error(err)
				return
			}

			id1, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}
			t.Log("first id is", id1)

			id2, err := factory.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
			}
			t.Log("first id is", id2)

			var count int64
			err = factory.SelectOne("countUsers").Scan(&count)
			if err != nil {
				t.Error("DELETE fail", err)
				return
			}

			if count != 2 {
				t.Error("count isnot 2, actual is", count)
			}

			_, err = factory.Delete("deleteUserTpl", tests.User{ID: id1})
			if err != nil {
				t.Error(err)
			}

			err = factory.SelectOne("countUsers").Scan(&count)
			if err != nil {
				t.Error("DELETE fail", err)
				return
			}

			if count != 1 {
				t.Error("count isnot 1, actual is", count)
			}

			_, err = factory.Delete("deleteUser", id2)
			if err != nil {
				t.Error(err)
			}

			err = factory.SelectOne("countUsers").Scan(&count)
			if err != nil {
				t.Error("DELETE fail", err)
				return
			}

			if count != 0 {
				t.Error("count isnot 0, actual is", count)
			}
		})

		t.Run("tx", func(t *testing.T) {
			_, err := factory.Delete("deleteAllUsers")
			if err != nil {
				t.Error(err)
				return
			}

			tx, err := factory.Begin()
			if err != nil {
				t.Error(err)
				return
			}

			id, err := tx.Insert("insertUser", insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			if err = tx.Commit(); err != nil {
				t.Error(err)
				return
			}

			_, err = factory.Delete("deleteUser", tests.User{ID: id})
			if err != nil {
				t.Error(err)
				return
			}
			tx, err = factory.Begin()
			if err != nil {
				t.Error(err)
				return
			}

			_, err = tx.Insert("insertUser", &insertUser)
			if err != nil {
				t.Error(err)
				return
			}

			if err = tx.Rollback(); err != nil {
				t.Error(err)
				return
			}

			var c int64
			err = factory.SelectOne("countUsers").Scan(&c)
			if err != nil {
				t.Error(err)
				return
			}
			if c != 0 {
				t.Error("count isnot 0, actual is", c)
			}
		})
	})
}