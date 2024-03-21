package handlers

import (
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	kafkaapp "popug_auth/kafka"
	"popug_auth/model"
)

func SiginupPage(c *gin.Context) {

	tmpl, err := template.ParseFiles("static/login.html")
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}

	tmpl.Execute(c.Writer, nil)
}

func UserListPage(c *gin.Context, db *gorm.DB) {
	var users []model.User

	result := db.Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении списка пользователей"})
		return
	}

	tmpl, err := template.ParseFiles("static/userlist.html")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при загрузке шаблона"})
		return
	}

	tmpl.Execute(c.Writer, gin.H{
		"Users": users,
		"Roles": model.KnownRoles,
	})
}
func RegisterPage(c *gin.Context) {
	tmpl, err := template.ParseFiles("static/register.html")
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}

	data := gin.H{
		"Roles": model.KnownRoles,
	}

	tmpl.Execute(c.Writer, data)

}

func CreateUserHandler(c *gin.Context, db *gorm.DB, kafka *kafkaapp.CudUser) {
	var form model.LoginForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pwdSalt, pwdHash, err := model.CalculatePasswordHash(form.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании хеша пароля"})
		return
	}

	user := model.User{
		PublicID:     uuid.New().String(),
		UserName:     form.Username,
		PasswordSalt: pwdSalt,
		PasswordHash: pwdHash,
		Role:         form.Role,
	}

	if result := db.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании пользователя"})
		return
	}

	go kafka.AddMsg(user)
	c.Redirect(http.StatusSeeOther, "/siginup")
}

func ChangeUserRoleHandler(c *gin.Context, db *gorm.DB, kafka *kafkaapp.CudUser) {
	// Получаем имя пользователя и новую роль из запроса
	var changeRoleRequest struct {
		Username string `json:"userName"`
		NewRole  string `json:"newRole"`
	}
	if err := c.ShouldBindJSON(&changeRoleRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// Проверяем, существует ли пользователь
	var user model.User
	if err := db.Where("user_name = ?", changeRoleRequest.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Проверяем, допустима ли новая роль
	isValidRole := false
	for _, role := range model.KnownRoles {
		if changeRoleRequest.NewRole == role {
			isValidRole = true
			break
		}
	}
	if !isValidRole {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимая роль"})
		return
	}

	user.Role = changeRoleRequest.NewRole
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении роли пользователя"})
		return
	}

	go kafka.AddMsg(user)

	c.JSON(http.StatusOK, gin.H{"message": "Роль пользователя успешно обновлена"})
}
