package controllers

import (
	"encoding/base64"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"newWeb/models"
	"strconv"
)

type UserController struct {
	beego.Controller
}

func (c *UserController) ShowRegister() {

	c.TplName = "register.html"

}

func (c *UserController) Register() {

	name := c.GetString("userName")
	password := c.GetString("password")

	if name == "" || password == "" {

		beego.Info("数据不完整,请重新输入!")
		return
	}

	user := &models.User{}

	user.Name = name
	user.Pwd = password

	o := orm.NewOrm()

	id, err := o.Insert(user)
	if err != nil {
		beego.Error("插入失败，请检查！")
		return
	}
	beego.Info("插入成功，id =", strconv.FormatInt(id, 10), "name=", name, "password=", password)
	//c.Ctx.WriteString("注册成功！")
	//c.TplName = "login.html"

	//跳转页面
	c.Redirect("/login", 302)
}

//显示登录界面
func (c *UserController) ShowLogin() {

	userName := c.Ctx.GetCookie("userName")

	//解密用户名
	bs, _ := base64.StdEncoding.DecodeString(userName)

	if userName == "" {
		//没有从cookie获取到用户名
		c.Data["userName"] = ""
		c.Data["checked"] = ""

	} else {
		//获取到用户名
		c.Data["userName"] = string(bs)
		c.Data["checked"] = "checked"
	}

	c.TplName = "login.html"

}

//登录操作
func (c *UserController) Login() {

	name := c.GetString("userName")
	password := c.GetString("password")

	if name == "" || password == "" {

		beego.Info("数据不完整,请重新输入!")
		return
	}

	user := &models.User{}

	user.Name = name
	user.Pwd = password

	o := orm.NewOrm()

	err := o.Read(user, "Name")
	if err != nil {
		beego.Error("用户名不存在，请重新输入！")
		return
	}

	if user.Pwd != password {
		beego.Error("密码错误，请重新输入！")
		return
	}
	//c.Ctx.WriteString("登录成功！")
	//处理记住用户名
	remember := c.GetString("remember")
	if remember == "on" {
		//加密 解决中文和特殊字符
		enName := base64.StdEncoding.EncodeToString([]byte(name))
		//记住用户名

		c.Ctx.SetCookie("userName", enName, 60)
	} else {
		//不记用户名
		c.Ctx.SetCookie("userName", name, -1)

	}

	//设置session

	c.SetSession("userName", name)

	//跳转页面
	c.Redirect("/article/index", 302)
}
