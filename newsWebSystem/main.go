package main

import (
	_ "newWeb/models"
	_ "newWeb/routers"
	"github.com/astaxie/beego"
)

func main() {
	beego.AddFuncMap("LastPage", LastPage)
	beego.AddFuncMap("NextPage", NextPage)
	beego.AddFuncMap("AddNum", AddNum)
	beego.Run()
}

func LastPage(page int) int {
	if page <= 1 {

		page = 1

	} else {
		page -= 1
	}
	return page
}

func NextPage(page, count int) int {
	if page >= count {
		return count
	} else {
		return page + 1
	}
}

func AddNum(a int) int {
	a++
	return a;
}
