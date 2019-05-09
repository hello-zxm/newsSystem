package controllers

import (
	"bytes"
	"encoding/gob"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/gomodule/redigo/redis"
	"math"
	"newWeb/models"
	"path"
	"strconv"
)

type ArticleController struct {
	beego.Controller
}

//展示首页
func (c *ArticleController) ShowIndex() {

	//获取session
	userName := c.GetSession("userName")
	if userName == nil {
		//没有获取到session 没有登录 没有通过授权

		c.Redirect("/login", 302)
		return
	}
	c.Data["userName"] = userName.(string)

	//获取所有文章数据，展示到页面

	o := orm.NewOrm()
	qs := o.QueryTable("Article")
	var articles []models.Article
	//qs.All(&articles)

	//获取选中的类型
	typeName := c.GetString("select")
	//总共数据个数
	var count int64

	beego.Info("typeName=", typeName)

	if typeName == "" || typeName == "全部" {
		//获取总记录数
		count, _ = qs.RelatedSel("ArticleType").Count()
	} else {
		count, _ = qs.RelatedSel("ArticleType").Filter("ArticleType__TypeName", typeName).Count()
	}

	//获取总页数
	pageIndex := 2

	pageCount := int(math.Ceil(float64(count) / float64(pageIndex)))
	//获取首页和末页数据
	//获取页码
	pageNum, err := c.GetInt("pageNum")
	if err != nil {
		pageNum = 1
	}
	beego.Info("当前页数为:", pageNum)

	//获取对应页的数据   获取几条数据     起始位置
	//ORM多表查询的时候默认是惰性查询 关联查询之后，如果关联的字段为空，数据查询不到

	//where ArticleType.typeName = typename   filter相当于where
	if typeName == "" || typeName == "全部" {
		qs.Limit(pageIndex, pageIndex*(pageNum-1)).RelatedSel("ArticleType").All(&articles)
	} else {
		qs.Limit(pageIndex, pageIndex*(pageNum-1)).RelatedSel("ArticleType").Filter("ArticleType__TypeName", typeName).All(&articles)
	}

	//查询文章类型 先从redis查询 如果redis没有 从mysql查询 并写入redis
	var aTypes []models.ArticleType

	conn, err := redis.Dial("tcp", ":6379")

	if err != nil {
		beego.Error("redis连接错误：", err)
		return
	}
	defer conn.Close()

	key := "articleType"
	//从redis获取key的值
	reply, err := conn.Do("get", key)

	//转换结果
	res, err := redis.Bytes(reply, err)

	if len(res) == 0 {
		//redis没有查到数据 从mysql查
		beego.Info("从mysql读取类型")
		//从mysql查询所有文章类型，并展示
		o.QueryTable("ArticleType").All(&aTypes)

		//查到后存入redis
		//序列化
		var buffer bytes.Buffer        //缓冲区
		enc := gob.NewEncoder(&buffer) //解码器
		err = enc.Encode(&aTypes)
		if err != nil {
			beego.Error("类型转码失败：", err)
			return
		}

		_, err = conn.Do("set", key, buffer.Bytes())
		if err != nil {
			beego.Error("redis写入错误:", err)
			return
		}

	} else {
		//redis查到数据	直接从redis读取
		beego.Info("从redis读取类型")

		//反序列化

		dec := gob.NewDecoder(bytes.NewReader(res))
		err = dec.Decode(&aTypes)
		if err != nil {
			beego.Error("类型解码失败:", err)
			return
		}

	}
	c.Data["aTypes"] = aTypes

	////从mysql查询所有文章类型，并展示
	//var aTypes []models.ArticleType
	//o.QueryTable("ArticleType").All(&aTypes)
	//
	//c.Data["aTypes"] = aTypes

	c.Data["articles"] = articles
	c.Data["count"] = count
	c.Data["pageCount"] = pageCount
	c.Data["pageNum"] = pageNum

	c.Data["TypeName"] = typeName

	c.Layout = "layout.html"

	//多级拼接 前端页面 如 css js 都可以 和c.data一样使用
	c.LayoutSections = make(map[string]string)
	c.LayoutSections["indexJS"] = "indexJS.html"

	c.TplName = "index.html"

}

//显示文章内容
func (c *ArticleController) ShowContent() {

	//获取
	id, err := c.GetInt("id")

	//校验
	if err != nil {

		beego.Error("获取id失败：" + err.Error())
		c.Redirect("/article/index", 302)
		return
	}

	//处理
	o := orm.NewOrm()

	a := &models.Article{}

	a.Id = id

	err = o.Read(a)
	if err != nil {
		beego.Error("获取文章失败：" + err.Error())
		return
	}

	//多对多查询

	//查询一 不能去重 参数1 要查询的对象指针 参数2 要查询的多对多字段名
	//o.LoadRelated(a, "Users")
	//
	//beego.Info("dadf:", a)

	//查询二 可以去重

	//高级查询   首先要指定表  多对多查询二   获取用户名   为了使用高级查询
	var users []models.User

	//先用queryTable 指定要查询的表a1  filter用来过滤 arm1是要查询的表a1的多表字段名 arm2是对应的表名 arm3是筛选字段 逗号后跟随arm3对应
	//的值   Distinct用来去重
	_, err = o.QueryTable("User").Filter("Articles__Article__Id", id).Distinct().All(&users)
	if err != nil {
		beego.Info("0000000000000", err)
	}
	c.Data["users"] = users

	beego.Info("ddadf:", users)

	//成功 阅读量+1 readcount 加1
	a.ReadCount += 1
	o.Update(a)

	//多对多添加

	//获取用户对象
	userName := c.GetSession("userName")

	user := &models.User{}
	user.Name = userName.(string)

	o.Read(user, "Name") //不是主键 要加属性名 再忘了别吃饭了！

	//获取多对多对象  参数1 操作对象 参数2 字段名
	m2m := o.QueryM2M(a, "Users")

	m2m.Add(user)

	c.Layout = "layout.html"

	//返回
	c.TplName = "content.html"
	c.Data["a"] = a

}

//处理编辑页面
func (c *ArticleController) HandleUpdateContent() {

	//获取数据
	//隐藏域传值
	id, err := c.GetInt("id")
	articleName := c.GetString("articleName")
	content := c.GetString("content")
	uploadname := HandlePicFile(c, "uploadname", "update.html")

	//校验
	if err != nil {
		beego.Error("获取id错误：", err.Error())
		return
	}
	beego.Info("id是：", id)

	if articleName == "" || content == "" || uploadname == "" {

		beego.Error("输入信息为空")
		c.Data["errmsg"] = "输入信息为空"
		c.Redirect("/article/update?id="+strconv.Itoa(id), 302)
		return

	}

	//处理
	o := orm.NewOrm()

	a := &models.Article{}

	a.Id = id

	o.Read(a)

	a.Title = articleName
	a.Content = content
	a.Image = uploadname
	o.Update(a)

	//返回

	//c.TplName = "update.html"
	c.Redirect("/article/index", 302)
}

//编辑文章
func (c *ArticleController) UpdateContent() {

	//获取数据

	id, err := c.GetInt("id")
	//校验
	if err != nil {
		beego.Error("获取文章id失败:", err.Error())
		c.Redirect("/article/index", 302)
		return
	}
	beego.Info("id是：", id)

	//处理
	o := orm.NewOrm()

	a := &models.Article{}

	a.Id = id

	err = o.Read(a)

	if err != nil {
		beego.Error("读取文章错误：", err.Error())
		return
	}

	//返回
	c.Data["a"] = a
	c.Layout = "layout.html"
	c.TplName = "update.html"
}

//删除文章
func (c *ArticleController) HandleDelete() {

	id, err := c.GetInt("id")
	if err != nil {
		beego.Error("获取id错误：", err.Error())
		c.Redirect("/article/index", 302)
		return
	}
	o := orm.NewOrm()

	a := &models.Article{}
	a.Id = id

	o.Delete(a)

	c.Redirect("/article/index", 302)
}

//显示添加文章
func (c *ArticleController) AddArticle() {

	//typeName:=c.GetString("select")

	//查询文章分类
	o := orm.NewOrm()

	var aTypes []models.ArticleType

	qs := o.QueryTable(new(models.ArticleType))

	qs.All(&aTypes)

	c.Data["aTypes"] = aTypes
	c.Layout = "layout.html"

	c.TplName = "add.html"

}

//处理添加文章
func (c *ArticleController) HandleAddArticle() {

	//获取数据
	articleName := c.GetString("articleName")
	content := c.GetString("content")
	typeName := c.GetString("select")
	uploadname := HandlePicFile(c, "uploadname", "add.html")

	if articleName == "" || content == "" || typeName == "" || uploadname == "" {

		beego.Info("数据不能为空，请输入完整数据")
		c.Data["errmsg"] = "数据不能为空，请输入完整数据"
		c.TplName = "add.html"
		return
	}

	//f, h, err := c.GetFile("uploadname")
	//if err != nil {
	//
	//	beego.Info("图片上传失败")
	//	c.Data["errmsg"] = "图片上传失败"
	//	c.TplName = "add.html"
	//
	//	return
	//}
	//defer f.Close()
	//
	//uploadname := h.Filename
	//exp := path.Ext(uploadname)
	//if exp != ".jpg" && exp != ".png" && exp != ".ico" {
	//
	//	beego.Info("图片格式不正确")
	//	c.Data["errmsg"] = "图片格式只能为jpg png"
	//	c.TplName = "add.html"
	//
	//	return
	//}
	//
	//if h.Size > 1024*1024 {
	//	beego.Info("图片过大")
	//	c.Data["errmsg"] = "图片过大"
	//	c.TplName = "add.html"
	//
	//	return
	//}
	//
	////重名判断
	//
	////c.SaveToFile("uploadname", "./src/beeProject/newWeb/static/img/"+uploadname)
	////beego中 ./为当前的beego工作目录
	//c.SaveToFile("uploadname", "./static/img/"+uploadname)

	o := orm.NewOrm()

	//查找articleType 表中的类别
	aType := &models.ArticleType{}
	aType.TypeName = typeName
	o.Read(aType, "TypeName") //切记 查询时 如果不是主键 一定要加字段名

	article := &models.Article{}

	article.Title = articleName
	article.Content = content
	article.Image = uploadname
	article.ArticleType = aType

	_, err := o.Insert(article)
	if err != nil {
		return
	}

	c.Redirect("/article/index", 302)

}

//显示添加分类
func (c *ArticleController) ShowAddType() {

	//查询所有分类 并显示
	o := orm.NewOrm()

	var aTypes []models.ArticleType

	qs := o.QueryTable(new(models.ArticleType))
	qs.All(&aTypes)

	c.Data["aTypes"] = aTypes
	c.Layout = "layout.html"

	c.TplName = "addType.html"

}

//处理添加分类
func (c *ArticleController) HandleAddType() {

	typeName := c.GetString("typeName")

	if typeName == "" {

		beego.Error("输入类型为空，请重新输入")
		c.Redirect("/article/addType", 302)
		return
	}

	//	处理数据
	o := orm.NewOrm()

	aType := &models.ArticleType{}
	aType.TypeName = typeName

	//判断是否有该分类
	err := o.Read(aType, "TypeName")
	if err == nil {
		//无错误 说明该类型已经存在
		beego.Error("该分类已经存在，请重新输入分类")
		c.Redirect("/article/addType", 302)
		return
	}

	//有错误 说明该类型不存在 可以正常添加
	o.Insert(aType)
	c.Redirect("/article/addType", 302)

}

//删除分类
func (c *ArticleController) DeleteType() {

	id, err := c.GetInt("id")
	if err != nil {
		beego.Error("获取id错误")
		return
	}
	beego.Info("aaaaaa:", id)
	o := orm.NewOrm()

	a := &models.ArticleType{}
	a.Id = id

	o.Delete(a)
	c.Redirect("/article/addType", 302)

}

//处理图片 通用
func HandlePicFile(c *ArticleController, keyName string, htmlName string) string {

	f, h, err := c.GetFile(keyName)
	if err != nil {

		beego.Info("图片上传失败")
		c.Data["errmsg"] = "图片上传失败"
		c.TplName = htmlName

		return ""
	}
	defer f.Close()

	uploadname := h.Filename
	exp := path.Ext(uploadname)
	if exp != ".jpg" && exp != ".png" && exp != ".ico" {

		beego.Info("图片格式不正确")
		c.Data["errmsg"] = "图片格式只能为jpg png ico"
		c.TplName = htmlName

		return ""
	}

	if h.Size > 1024*1024 {
		beego.Info("图片过大")
		c.Data["errmsg"] = "图片过大"
		c.TplName = htmlName

		return ""
	}

	//重名判断

	//c.SaveToFile("uploadname", "./src/beeProject/newWeb/static/img/"+uploadname)
	//beego中 ./为当前的beego工作目录
	c.SaveToFile(keyName, "./static/img/"+uploadname)

	return "/static/img/" + uploadname

}
