package models

import (
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type User struct {
	Id   int
	Name string
	Pwd  string

	Articles []*Article `orm:"reverse(many)"`
}

type Article struct {
	Id        int       `orm:"pk;auto"`
	Title     string    `orm:"unique;size(50)"`
	Content   string    `orm:"size(500)"`
	Image     string    `orm:"null"`
	Time      time.Time `orm:"type(datetime);auto_now_add"`
	ReadCount int       `orm:"default(0)"`

	ArticleType *ArticleType `orm:"rel(fk);null;on_delete(set_null)"`
	Users       []*User      `orm:"rel(m2m)"`
}

type ArticleType struct {
	Id       int
	TypeName string     `orm:"unique"`
	Article  []*Article `orm:"reverse(many)"`
}

func init() {

	aName := "default"

	orm.RegisterDataBase(aName, "mysql", "root:zhaoming@tcp(127.0.0.1:3306)/newWeb")

	orm.RegisterModel(new(User), new(Article), new(ArticleType))
	//orm.RegisterModel(new(Article))

	orm.RunSyncdb(aName, false, true)
	//orm.RunSyncdb(aName, true, true)

}
