package routes

import (
	"crawlab/constants"
	"crawlab/model"
	"crawlab/services"
	"crawlab/utils"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

type UserListRequestData struct {
	PageNum  int `form:"page_num"`
	PageSize int `form:"page_size"`
}

type UserRequestData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func GetUser(c *gin.Context) {
	id := c.Param("id")

	user, err := model.GetUser(bson.ObjectIdHex(id))
	if err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
		Data:    user,
	})
}

func GetUserList(c *gin.Context) {
	// 绑定数据
	data := UserListRequestData{}
	if err := c.ShouldBindQuery(&data); err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}
	if data.PageNum == 0 {
		data.PageNum = 1
	}
	if data.PageSize == 0 {
		data.PageNum = 10
	}

	// 获取用户列表
	users, err := model.GetUserList(nil, (data.PageNum-1)*data.PageSize, data.PageSize, "-create_ts")
	if err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	// 获取总用户数
	total, err := model.GetUserListTotal(nil)
	if err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	// 去除密码
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, ListResponse{
		Status:  "ok",
		Message: "success",
		Data:    users,
		Total:   total,
	})
}

func PutUser(c *gin.Context) {
	// 绑定请求数据
	var reqData UserRequestData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}

	// 添加用户
	user := model.User{
		Username: strings.ToLower(reqData.Username),
		Password: utils.EncryptPassword(reqData.Password),
		Role:     constants.RoleNormal,
	}
	if err := user.Add(); err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
	})
}

func PostUser(c *gin.Context) {
	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {
		HandleErrorF(http.StatusBadRequest, c, "invalid id")
	}

	var item model.User
	if err := c.ShouldBindJSON(&item); err != nil {
		HandleError(http.StatusBadRequest, c, err)
		return
	}

	if err := model.UpdateUser(bson.ObjectIdHex(id), item); err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
	})
}

func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {
		HandleErrorF(http.StatusBadRequest, c, "invalid id")
		return
	}

	// 从数据库中删除该爬虫
	if err := model.RemoveUser(bson.ObjectIdHex(id)); err != nil {
		HandleError(http.StatusInternalServerError, c, err)
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
	})
}

func Login(c *gin.Context) {
	// 绑定请求数据
	var reqData UserRequestData
	if err := c.ShouldBindJSON(&reqData); err != nil {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}

	// 获取用户
	user, err := model.GetUserByUsername(strings.ToLower(reqData.Username))
	if err != nil {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}

	// 校验密码
	encPassword := utils.EncryptPassword(reqData.Password)
	if user.Password != encPassword {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}

	// 获取token
	tokenStr, err := services.GetToken(user.Username)
	if err != nil {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
		Data:    tokenStr,
	})
}

func GetMe(c *gin.Context) {
	// 获取token string
	tokenStr := c.GetHeader("Authorization")

	// 校验token
	user, err := services.CheckToken(tokenStr)
	if err != nil {
		HandleError(http.StatusUnauthorized, c, errors.New("not authorized"))
		return
	}
	user.Password = ""

	c.JSON(http.StatusOK, Response{
		Status:  "ok",
		Message: "success",
		Data:    user,
	})
}
