package routes

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PromptPal/PromptPal/ent"
	"github.com/PromptPal/PromptPal/ent/user"
	"github.com/PromptPal/PromptPal/service"
	"github.com/gin-gonic/gin"
)

type authPayload struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Message   string `json:"message"`
}

type authResponse struct {
	User  ent.User `json:"user"`
	Token string   `json:"token"`
}

func authHandler(c *gin.Context) {
	payload := authPayload{}
	if err := c.Bind(&payload); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}
	// do web3 check
	verified, err := web3Service.VerifySignature(payload.Address, payload.Message, payload.Signature)

	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}
	if !verified {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid signature",
		})
		return
	}

	u, err := service.
		EntClient.
		User.
		Query().
		Where(user.Addr(strings.ToLower(payload.Address))).
		Only(c)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	// sign web3 token to client
	token, err := service.SignJWT(u, time.Hour*24*30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, authResponse{
		User:  *u,
		Token: token,
	})
}

func listUsers(c *gin.Context) {
	var query queryPagination
	if err := c.BindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: err.Error(),
		})
		return
	}
	// check signed data
	uid := c.GetInt("uid")

	u, err := service.EntClient.User.Get(c, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	// if < 255, which means it's a normal user
	if u.Level < 255 {
		c.JSON(http.StatusForbidden, errorResponse{
			ErrorCode:    http.StatusForbidden,
			ErrorMessage: "forbidden. only admin can get users",
		})
		return
	}

	us, err := service.
		EntClient.
		User.
		Query().
		Where(user.IDLT(query.Cursor)).
		Limit(query.Limit).
		Order(ent.Desc(user.FieldID)).
		All(c)

	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}

	count, err := service.EntClient.User.Query().Count(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ListResponse[*ent.User]{
		Count: count,
		Data:  us,
	})
}

func getUser(c *gin.Context) {
	uidStr, ok := c.Params.Get("id")
	if !ok {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid id",
		})
		return
	}

	uid, err := strconv.Atoi(uidStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: err.Error(),
		})
		return
	}

	if uid == -1 {
		uid = c.GetInt("uid")
	}

	if uid <= 0 {
		c.JSON(http.StatusBadRequest, errorResponse{
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid id",
		})
		return
	}

	u, err := service.EntClient.User.Get(c, uid)
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{
			ErrorCode:    http.StatusNotFound,
			ErrorMessage: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, u)
}

func createUsers(c *gin.Context) {
	// check signed data
}

func removeUsers(c *gin.Context) {
	// check signed data
}
