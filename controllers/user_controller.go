package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"kars/database"
	"kars/jwtoken"
	"kars/models"
	"kars/utils"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
}

var store = session.New()
var ctx = context.Background()

type UserInput struct {
	UserName string `json:"user_name" gorm:"size:255"`
	Email    string `json:"email" gorm:"size:255;not null;unique"`
	Password string `json:"password" gorm:"size:255;not null"`
	PhoneNo  string `json:"phone_no" gorm:"size:10"`
}

var redisClient = redis.NewClient(&redis.Options{
	Addr:     "redis:6379", // Ensure this is correct (default Redis port)
	Password: "",               // Make sure it's empty if no password is set
	DB:       0,
})

func UserSignUp(c *fiber.Ctx) error {

	var Input UserInput

	if err := c.BodyParser(&Input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	if err := Input.UserValidate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var ExistingUser models.User

	if err := database.DB.Where("email = ?", Input.Email).First(&ExistingUser).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "User already exists",
		})
	}

	otp, err := utils.GenerateOTP(Input.Email, 6)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating OTP",
		})
	}

	fmt.Println(Input.Email)
	fmt.Println(otp.Code)

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not connect to Redis",
		})
	}

	userData := map[string]interface{}{
		"otp":       otp.Code,
		"user_name": Input.UserName,
		"password":  Input.Password,
		"email":     Input.Email,
		"phone_no":  Input.PhoneNo,
	}

	err = redisClient.HMSet(ctx, Input.Email, userData).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to store OTP and user data in Redis",
		})
	}

	err = redisClient.Expire(ctx, Input.Email, time.Minute).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to set expiration time for OTP and user data in Redis",
		})
	}

	if err := utils.SendOtp(Input.Email, otp.Code); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error sending OTP email",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OTP sent successfully",
		"otp":     otp.Code,
	})
}

func VerifyOtpAndCreateUser(c *fiber.Ctx) error {
	type VerifyInput struct {
		VerifyOTP string `json:"otp"`
	}

	var input VerifyInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Input",
		})
	}

	email := c.Params("user_email")
	if email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is Required",
		})
	}

	storedData, err := redisClient.HGetAll(ctx, email).Result()
	fmt.Print(storedData)
	if err != nil || storedData["otp"] != input.VerifyOTP {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired OTP",
		})
	}

	redisClient.Del(ctx, email)

	HashedPassword, err := bcrypt.GenerateFromPassword([]byte(storedData["password"]), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	newUser := models.User{
		UserName: storedData["user_name"],
		Email:    storedData["email"],
		Password: string(HashedPassword),
		PhoneNo:  storedData["phone_no"],
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error creating user",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user": fiber.Map{
			"user_id":   newUser.ID,
			"user_name": newUser.UserName,
			"email":     newUser.Email,
			"phone_no":  newUser.PhoneNo,
			"status":    newUser.Status,
		},
	})
}

func ResendOTP(c *fiber.Ctx) error {
	email := c.Params("user_email")
	if email == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Email is Required",
		})
	}

	otp, err := utils.GenerateOTP(email, 6)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating OTP",
		})
	}

	fmt.Println("user:", email)
	fmt.Println("New OTP:", otp.Code)
	fmt.Println("Expiration Time:", otp.ExpiredAt)

	err = redisClient.HSet(ctx, email, "otp", otp.Code).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update OTP in Redis",
		})
	}

	err = redisClient.Expire(ctx, email, time.Minute).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to set expiration time for OTP and user data in Redis",
		})
	}

	if err := utils.SendOtp(email, otp.Code); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to Resend OTP",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OTP Resend successfully",
		"email":   email,
	})
}

var state = "random_state_string"

func InitGoogleSignIn(c *fiber.Ctx) error {

	authURL := utils.GetGoogleAuthURL(state)
	log.Println("in user:", authURL)
	return c.Redirect(authURL)
}

func GoogleSignUpCallback(c *fiber.Ctx) error {

	if c.Query("state") != state {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid state",
		})
	}

	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Code not found",
		})
	}

	token, err := utils.ExchangeCodeForToken(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to exchange token",
		})
	}

	client := utils.GetGoogleOAuthConfig().Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user info",
		})
	}
	defer resp.Body.Close()

	var GoogleUser struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&GoogleUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed decode user details",
		})
	}

	var ExistingUser models.User
	if err := database.DB.Where("email = ?", GoogleUser.Email).First(&ExistingUser).Error; err == nil {
		return c.JSON(fiber.Map{
			"message": "User already exists",
			"user":    ExistingUser,
		})
	}

	newUser := models.User{
		UserName: GoogleUser.Name,
		Email:    GoogleUser.Email,
		Status:   "Active",
	}

	if err := database.DB.Create(&newUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	jwtToken, err := jwtoken.GenerateUserJWT(ExistingUser.ID, ExistingUser.IsBlocked, ExistingUser.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate JWT",
		})
	}

	return c.JSON(fiber.Map{
		"message":   "User sign in successfully",
		"user_id":   newUser.ID,
		"user_name": newUser.UserName,
		"email":     newUser.Email,
		"status":    newUser.Status,
		"jwt_token": jwtToken,
	})
}

func UserLogin(c *fiber.Ctx) error {
	type UserInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input UserInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	var ExistingUser models.User
	result := database.DB.Where("email = ?", input.Email).First(&ExistingUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if ExistingUser.IsBlocked {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You are blocked from the website",
		})
	}

	if ExistingUser.Status == "Active" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User is already logged in",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(ExistingUser.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Incorrect Password",
		})
	}

	ExistingUser.Status = "Active"
	token, err := jwtoken.GenerateUserJWT(ExistingUser.ID, ExistingUser.IsBlocked, ExistingUser.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create JWT token",
		})
	}

	if err := database.DB.Save(&ExistingUser).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update user status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "User login successfully",
		"token":   token,
	})

}

func UserLogout(c *fiber.Ctx) error {

	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing Token",
		})
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid authorization format",
		})
	}
	tokenString := tokenParts[1]

	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtoken.JwtSecret, nil
	})
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Token expiration not found",
		})
	}
	expTime := time.Unix(int64(exp), 0)

	err = redisClient.Set(c.Context(), tokenString, "blacklisted", time.Until(expTime)).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to blacklist token",
		})
	}

	userID := claims["user_id"]

	var user models.User
	Result := database.DB.First(&user, userID)
	if Result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	user.Status = "Inactive"
	database.DB.Save(&user)

	return c.JSON(fiber.Map{
		"message": "User logged out successfullly",
	})
}

func UserProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	var user models.User

	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user data",
		})
	}

	response := fiber.Map{
		"name":       user.UserName,
		"email":      user.Email,
		"phone_no":   user.PhoneNo,
		"created_at": user.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User details retrieved successfully",
		"details": response,
	})
}

func EditProfile(c *fiber.Ctx) error {

	userid, err := convertToUint(c.Locals("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid user_id: %s", err.Error()),
		})
	}

	var input UserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse details",
		})
	}

	if err := input.UserValidateForPatch(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	fields := map[string]interface{}{}
	if input.UserName != "" {
		fields["user_name"] = input.UserName
	}
	if input.PhoneNo != "" {
		fields["phone_no"] = input.PhoneNo
	}

	var updateinput models.User
	if err := database.DB.Model(&updateinput).Where("id = ?", userid).Updates(&fields).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update user details",
		})
	}

	if err := database.DB.First(&updateinput, "id = ?", userid).Error; err != nil {
		log.Println(err, userid)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch user details",
		})
	}

	response := fiber.Map{
		"user_name":  updateinput.UserName,
		"email":      updateinput.Email,
		"phone_no":   updateinput.PhoneNo,
		"created_at": updateinput.CreatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":              "User details updated successfully",
		"updated user details": response,
	})

}

func ForgotPasswordStep1(c *fiber.Ctx) error {

	userid, err := convertToUint(c.Locals("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid user_id: %s", err.Error()),
		})
	}

	fmt.Println(userid)
	var user models.User
	if err := database.DB.First(&user, "id = ?", userid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve user data",
		})
	}

	otp, err := utils.GenerateOTP(user.Email, 6)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error generating OTP",
		})
	}

	session, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error initializing session",
		})
	}
	session.Set("verified", false)
	session.Set("otp", otp.Code)
	session.Set("user_id", userid)
	if err := session.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error saving session",
		})
	}

	fmt.Println(user.Email)
	fmt.Println(otp.Code)

	if err := utils.SendOtp(user.Email, otp.Code); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error sending OTP email",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OTP sent successfully",
	})
}

func ForgotPasswordStep2(c *fiber.Ctx) error {

	userid, err := convertToUint(c.Locals("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid user_id: %s", err.Error()),
		})
	}

	type VerifyInput struct {
		VerifyOTP string `json:"otp"`
	}

	var input VerifyInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse user input",
		})
	}

	session, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error retrieving session",
		})
	}

	storedOTP := session.Get("otp")
	userId := session.Get("user_id")

	if storedOTP == nil || userId == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "OTP expired",
		})
	}

	if userId != userid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user not found",
		})
	}

	if storedOTP.(string) != input.VerifyOTP {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid OTP",
		})
	}

	session.Set("verified", true)
	if err := session.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error saving session",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "OTP successfully verified",
	})

}

func ForgotPasswordStep3(c *fiber.Ctx) error {

	userid, err := convertToUint(c.Locals("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid user_id: %s", err.Error()),
		})
	}

	type UserInput struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	var input UserInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse user input",
		})
	}

	if input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password required",
		})
	}

	if len(input.Password) < 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password must be at least 6 characters long",
		})
	}

	if input.Password != input.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password and confirm password must be same",
		})
	}

	session, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error retrieving session",
		})
	}

	verified := session.Get("verified")
	if verified == nil || !verified.(bool) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "OTP is not verified",
		})
	}

	userId := session.Get("user_id")
	if userId == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User not found in session",
		})
	}

	if userId != userid {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User ID in session does not match URL parameter",
		})
	}

	HashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	var user models.User
	if err := database.DB.Model(&user).Where("id = ?", userid).Update("password", string(HashedPassword)).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error updating password",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Password updated successfully",
	})

}
