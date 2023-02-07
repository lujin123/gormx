# gormx

`gormx` 是对 `gorm` 的高阶函数操作的简单封装

## 使用

- 初始化

调用 `New` 实例化一个新的数据库对象，自定义的参数通过 `Config` 传入

```go
conf := &Config{
    Dialector:   postgres.Open(dsn), //使用 postgres 驱动
    MaxIdleConn: 10,
    MaxOpenConn: 10,
    MaxLifetime: 1000,
    Debug:       false,
}
db, _ := New(conf, opts...)

// 获取原始 gorm 对象
// gormdb := db.DB()
```

- 高阶函数构建

```go
func WithName(name string) Option {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("name=?", name)
    }
}
```

- 创建记录

```go
// 创建单条记录
user := User{
    Nickname: "hello",
}
db.Insert(&user)

// 创建多条记录
var users []User
for i := 0; i < 2; i++ {
    users = append(users, User{
        Nickname: fmt.Sprintf("hello %d", i),
        Age:      int64(i),
    })
}
db.Insert(users)
```

- 查询单条记录

```go
var user = User{
    Id: 1,
}
// model 中的 ID 需要是主键
db.FindOne(&user)
//SQL: select * from test_users where id=1;

// 带有 where 条件
func WithName(name string) Option {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("name=?", name)
    }
}

var user User
db.FindOne(&user, WithName("hello"))
//SQL: select * from test_users where name='hello';
```

- 查询多条记录

```go
var user []User
db.FindMany(&user, Pagination(1, 2))
//SQL: select * from test_users limit 2;
```

- 更新记录

```go
// 更新不存在的记录，会返回 `ErrNoRowsAffected`，注意检查错误
err = db.Model(&User{Id: -1}).Update("nickname", "hello world")

// 更新单字段
err = db.Model(&User{Id: 1}).Update("nickname", "hello world")

// 使用struct更新多字段
err = db.Updates(&User{
    Id:       1,
    Nickname: "hello struct",
    Age:      133,
})

// 使用 map 更新多字段
err = db.Model(&User{Id: 1}).Updates(map[string]interface{}{
    "nickname": "hello map",
    "age":      "133",
})
```

- 删除记录

```go
user := User{Id: 1}
err = db.Delete(&user)
```

- 检查记录是否存在

```go
exists, err := db.Exists(&User{
    Id: 1,
})
// SQL: select exists(select * from test_users where id=1 limit 1);
```

- 获取记录数

```go
total, err := db.Model(&User{}).Count()
// SQL: SELECT count(*) FROM test_users;
```

- SQL 执行

```go
// 建表
db.Exec("create table test_users (id serial primary key not null, nickname varchar(64) not null, age integer default 0);")

```

- SQL 查询

```go
var (
    query = `select * from test_users where id=?`
    args  = []interface{}{1}
    user  User
)

// 执行查询 SQL 后再将数据映射到model 中
db.Raw(query, args...).FindOne(&user)

var myUser struct {
    Id       int64
    Nickname string
}
// 执行 SQL 后将数据映射到任意自定义对象
db.Raw(query, args...).Scan(&myUser)
```

- 上下文Context设置

```go
// 请求之前都需要设置 Context，否则无法追踪调用链路
db.WithContext(ctx).FindOne(...)
```

- 事务

```go
err := db.WithContext(ctx).Tx(func(tx *db) error {
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
    return nil
})
```

## 最后

`gorm`对简单 `SQL` 操作比较好用，复杂的查询还得使用原生 `SQL`，所以不能满足使用的时候，取出 `gorm` 对象自己操作 `SQL` 就行了
