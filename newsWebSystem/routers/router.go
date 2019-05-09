package routers

import (
	"newWeb/controllers"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func init() {

	//路由之前 先进行过滤
	//InsertFilter(arm1,arm2,arm3) arm1要过滤的路径(支持正则） arm2过滤器要过滤的位置 arm3回调
	beego.InsertFilter("/article/*", beego.BeforeExec,
		func(context *context.Context) {

			userName := context.Input.Session("userName")
			if userName == nil {
				//未登录授权 跳转至登录页面
				context.Redirect(302, "/login")
				return
			}
			//登录授权，不拦截

		})

	beego.Router("/", &controllers.MainController{})
	//注册
	beego.Router("/register", &controllers.UserController{}, "get:ShowRegister;post:Register")
	//登录
	beego.Router("/login", &controllers.UserController{}, "get:ShowLogin;post:Login")
	//首页
	beego.Router("/article/index", &controllers.ArticleController{}, "get,post:ShowIndex")
	//添加文章
	beego.Router("/article/addArticle", &controllers.ArticleController{}, "get:AddArticle;post:HandleAddArticle")
	//显示文章详情
	beego.Router("/article/content", &controllers.ArticleController{}, "get:ShowContent")
	//编辑文章
	beego.Router("/article/update", &controllers.ArticleController{}, "get:UpdateContent;post:HandleUpdateContent")
	//删除文章
	beego.Router("/article/delete", &controllers.ArticleController{}, "get:HandleDelete")
	//添加文章分类
	beego.Router("/article/addType", &controllers.ArticleController{}, "get:ShowAddType;post:HandleAddType")
	//删除一个分类
	beego.Router("/article/deleteType", &controllers.ArticleController{}, "get:DeleteType")

}
