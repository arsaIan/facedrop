package controller

import (
	"facedrop/config"
	"facedrop/logger"
	"facedrop/models"
	"facedrop/service"
	"facedrop/utils"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
	storageClient *service.StorageClient
	cfg *config.Config
}

func NewUserController(userService *service.UserService, cfg *config.Config, storageClient *service.StorageClient) *UserController {
	return &UserController{userService: userService, cfg: cfg, storageClient: storageClient}
}


func (c *UserController) Register(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.userService.RegisterUser(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, user)
}
func (c*UserController) Login(ctx *gin.Context) {
	var loginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	user, err := c.userService.Login(loginRequest.Email, loginRequest.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}	
	
	token, err := c.userService.GenerateToken(user.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}	
	ctx.JSON(http.StatusOK, gin.H{"token": token})
	
}
func (c*UserController) Logout(ctx *gin.Context) {
	ctx.SetCookie("token", "", -1, "/", "", false, true)
	ctx.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
func (c *UserController) GetUser(ctx *gin.Context) {
	id := ctx.Param("id")
	idUint, err := utils.StrToUint(id)
	if err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	logger.Info("Getting user by id", logger.Int("id", int(idUint)))
	user, err := c.userService.GetUserByID(idUint)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	_ = ctx.Param("id")
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.userService.UpdateUser(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

func (c *UserController) DeleteUser(ctx *gin.Context) {
	id := ctx.Param("id")
	idUint, err := utils.StrToUint(id)
	if err != nil {
		logger.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.userService.DeleteUser(idUint); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
} 

func (c *UserController) UploadFace(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}		

	file, err := ctx.FormFile("face")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "face file is required"})
		return
	}	
	
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()	

	fileContent, err := io.ReadAll(src)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}	
	fileKey := fmt.Sprintf("users/%d/face/%s", userID, file.Filename)
	faceURL, err := c.storageClient.UploadFile(ctx, fileKey, fileContent, c.cfg.StorageConfig.UserBucket)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload face"})
		return
	}
	
	userFace := models.UserFace{
		UserID: userID,
		FaceURL: faceURL,
	}

	if err := c.userService.CreateUserFace(&userFace); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user face"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "face uploaded successfully"})
	
	
}