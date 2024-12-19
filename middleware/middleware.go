package middleware

import (
	"errors"
	"kars/database"
	"kars/jwtoken"
	"kars/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var redisClient = redis.NewClient(&redis.Options{
	Addr: "localhost:6379",
})

func CheckUserStatus(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is missing",
		})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid authorization format. Expected 'Bearer <token>'",
		})
	}

	tokenString := parts[1]

	isBlacklisted, err := redisClient.Get(c.Context(), tokenString).Result()
	if err == nil && isBlacklisted == "blacklisted" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token has been blacklisted",
		})
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
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
			"error": "Unable to parse token claims",
		})
	}

	UserId := claims["user_id"]
	var user models.User
	if err := database.DB.First(&user, "id = ?", UserId).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve user data",
		})
	}

	if user.IsBlocked {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User blocked from the website",
		})
	}

	if status, ok := claims["status"].(string); !ok || status != "Active" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User is not active",
		})
	}

	c.Locals("user_id", UserId)

	return c.Next()
}

func AdminMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing token",
		})
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid authorization format",
		})
	}
	tokenString := tokenParts[1]

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtoken.JwtSecret, nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	isBlacklisted, err := redisClient.Get(c.Context(), tokenString).Result()
	if err != redis.Nil && isBlacklisted == "blacklisted" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Token is blacklisted",
		})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	adminStatus, ok := claims["admin_status"].(string)
	if !ok || adminStatus != "Active" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin is not active",
		})
	}

	return c.Next()
}
