package handler

import (
	"fmt"
	"net/http"
	"strings"

	"e-commerce.com/internal/models"
	"e-commerce.com/internal/service"
	"e-commerce.com/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func FirstName(fullname string) string {
	// Trim leading/trailing spaces first
	trimmed := strings.TrimSpace(fullname)
	// Split by whitespace (space, tab, etc)
	parts := strings.Fields(trimmed)
	if len(parts) > 0 {
		return parts[0] // first word
	}
	return "" // empty string if no parts found
}

func (h *UserHandler) LoginUserHandler(c *gin.Context) {
	var userLogin models.UserLogin

	if err := c.ShouldBindJSON(&userLogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request payload",
			"success": false,
		})
		return
	}

	user, err := h.userService.Login(c.Request.Context(), &userLogin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	token, err := utils.GenerateJWT(utils.JwtDataType{
		UserId:   user.ID,
		FullName: user.Username,
		Email:    user.Email,
	})

	if token == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to generate token",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"message": "login successfully",
	})

}

func (h *UserHandler) DeleteUserSessionHandler(c *gin.Context) {
	token := utils.ExtractToken(c, "user_token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "no token found",
			"success": false,
		})
		return
	}
	fmt.Println("this is token in delete user session handler ; ", token)
	user, err := utils.ParseJwt(token)
	fmt.Println("this is user in delete user session handler ; ", user)
	fmt.Println("this is err in delete user session handler ; ", err)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user token",
			"success": false,
		})
		return
	}

	key := fmt.Sprintf("userId:usersession:%s", user.UserId)

	if err := h.userService.DeleteUserSession(c.Request.Context(), key); err != nil {
		fmt.Println("this is err in delete user session handler ; ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete session",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "successfully deleted user session",
		"success": true,
	})

}

func (h *UserHandler) GetUserFromToken(c *gin.Context) {
	token := utils.ExtractToken(c, "user_token")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "no token found",
			"success": false,
		})
		return
	}
	user, err := utils.ParseJwt(token)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid user token",
			"success": false,
		})
		return
	}

	userData, err := h.userService.GetUserSession(c.Request.Context(), fmt.Sprintf("userId:usersession:%s", user.UserId))

	if err != nil {
		userDb, err := h.userService.GetUserByEmail(user.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid user token",
				"success": false,
			})
			return
		}
		if userDb == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid user",
				"success": false,
			})
			return
		}
		fmt.Println("this is user data from db : ", userDb)
		safeUser := models.SafeUser{
			ID:         userDb.ID,
			Username:   userDb.Username,
			Email:      userDb.Email,
			Role:       userDb.Role,
			IsVerified: userDb.IsVerified,
			CreatedAt:  userDb.CreatedAt,
			UpdatedAt:  userDb.UpdatedAt,
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "successfully verified",
			"user":    safeUser,
			"success": true,
		})
		return

	}
	safeUser := models.SafeUser{
		ID:         userData.ID,
		Username:   userData.Username,
		Email:      userData.Email,
		Role:       userData.Role,
		IsVerified: userData.IsVerified,
		CreatedAt:  userData.CreatedAt,
		UpdatedAt:  userData.UpdatedAt,
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "successfully verified",
		"user":    safeUser,
		"success": true,
	})
}

func (h *UserHandler) VerifyUserTokenHandler(c *gin.Context) {
	userToken := c.Query("token")
	if userToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid  token",
			"success": false,
		})
		return
	}

	userData, err := h.userService.VerifyUserToken(fmt.Sprintf("verifyToken:%s", userToken), c.Request.Context())

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	jwtData := utils.JwtDataType{
		UserId:   userData.ID,
		FullName: userData.Username,
		Email:    userData.Email,
	}

	token, err := utils.GenerateJWT(jwtData)

	if token == "" || err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "user created, login due to internal problem",
			"success": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"message": "verified successfully",
	})
}

func (h *UserHandler) RegisterUserHandler(c *gin.Context) {
	var user models.User

	// Bind JSON payload to User struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request payload",
			"success": false,
		})
		return
	}

	// Validate required fields
	if user.Email == "" || user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Email and password are required",
			"success": false,
		})
		return
	}

	message, err := h.userService.Register(c.Request.Context(), &user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "email already registered" {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{
			"error":   err.Error(),
			"success": false,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": message,
		"success": true,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
		},
	})

}
