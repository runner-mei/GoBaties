# 事务的支持


事务的支持比较简单，有两种方式

## 1. 通过 context 参数传递

你可以在 接口的方法上增加一个 context.Context 参数，然后用 gobatis.WithTx 传递事务

````go

  type UserDao interface  {
    Update(ctx context.Context, id int64, user User) error
  }
  // ...
  userDao := NewUserDao(sessionRef)
  // ...
  db, _ := sql.Open("mysql", "test:@tcp(localhost:3306)/golang?autocommit=true&parseTime=true&multiStatements=true")


  // ...
  tx := db.Begin()
  defer tx.Rollback()
  ctx := WithTx(context.Background(), tx)
  userDao.Update(ctx, 1, u1)
  userDao.Update(ctx, 2, u2)
  tx.Commit()


````

完整例子请看 https://github.com/runner-mei/GoBatis/blob/master/example/transaction_test.go 中的 ExampleTx1

## 2. 另一种方法

````go

  // 下面的 Connection， Tx 和 NewConnection 请看
  // https://github.com/runner-mei/GoBatis/blob/master/example/connection.go

  conn := NewConnection(factory)

  tx := conn.Begin()
  defer tx.Rollback()
  tx.Users().Update(ctx, 1, u1)
  tx.Users().Update(ctx, 2, u2)
  tx.Commit()
  
````
完整例子请看 https://github.com/runner-mei/GoBatis/blob/master/example/transaction_test.go 中的 ExampleTx2
