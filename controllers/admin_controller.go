package controllers

import (
	"fmt"
	"kars/database"
	"kars/jwtoken"
	"kars/models"
	"log"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AdminInput struct {
	Name     string `json:"admin_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func AdminSignUp(c *fiber.Ctx) error {

	var input AdminInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Input",
		})
	}

	if err := input.AdminValidate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var existingAdmin models.Admin

	if err := database.DB.Where("email = ?", input.Email).First(&existingAdmin).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Admin already exists",
		})
	} else if err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error checking existing admin",
		})
	}

	HashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	admin := models.Admin{
		AdminName: input.Name,
		Email:     input.Email,
		Password:  string(HashedPassword),
	}

	if err := database.DB.Create(&admin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create admin",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Admin created successfully",
		"name":    admin.AdminName,
		"email":   admin.Email,
	})

}

func AdminLogin(c *fiber.Ctx) error {
	type AdminInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var input AdminInput

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Input",
		})
	}

	var existingAdmin models.Admin

	result := database.DB.Where("email = ?", input.Email).First(&existingAdmin)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "admin not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "database error",
		})
	}

	if existingAdmin.Status == "Active" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "admin is already login",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingAdmin.Password), []byte(input.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "incorrect Password",
		})
	}

	existingAdmin.Status = "Active"

	token, err := jwtoken.GenerateAdminJWT(existingAdmin.ID, existingAdmin.Status)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create JWT token",
		})
	}

	if err := database.DB.Save(&existingAdmin).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update admin status",
		})
	}

	return c.JSON(fiber.Map{
		"message": "admin login successfully",
		"token":   token,
	})

}

func UserList(c *fiber.Ctx) error {

	var Users []models.User

	if err := database.DB.Find(&Users).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users details",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully fetched all the use details",
		"Users":   Users,
	})
}
func BlockUser(c *fiber.Ctx) error {
	UserId := c.Params("user_id")
	if UserId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	var user models.User
	result := database.DB.First(&user, UserId)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	if user.IsBlocked {
		user.IsBlocked = false
		user.Status = "Inactive"
		if err := database.DB.Save(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to block user",
			})
		}
	} else {
		user.IsBlocked = true
		user.Status = "Inactive"
		if err := database.DB.Save(&user).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to unblock user",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "User updated successfully",
		"user":    user,
	})
}

func OrderList(c *fiber.Ctx) error {

	var order []models.Order
	if err := database.DB.Preload("OrderItems").Find(&order).Error; err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve orders",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "orders successfully retrieved",
		"orders":  order,
	})
}

func ChangeStatusShipped(c *fiber.Ctx) error {

	orderid := c.Params("order_id")
	if orderid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	var order models.Order
	if err := database.DB.First(&order, "id = ?", orderid).Error; err != nil {
		log.Println(err)
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve order",
		})
	}

	switch order.OrderStatus {
	case "cancelled":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order already cancelled",
		})
	case "shipped":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order already cancelled",
		})
	}

	if err := database.DB.Model(&order).Update("order_status", "shipped").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update order status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully updated order status",
		"order":   order,
	})
}

func TopSellingProducts(c *fiber.Ctx) error {

	var topProducts []struct {
		ProductID    uint    `json:"product_id"`
		ProductName  string  `json:"product_name"`
		ProductPrice float64 `json:"product_price"`
		TotalSold    int     `json:"total_sold"`
	}

	var topCategories []struct {
        CategoryID   uint   `json:"category_id"`
        CategoryName string `json:"category_name"`
        TotalSold    int    `json:"total_sold"`
    }

	if err := database.DB.
		Table("order_items").
		Select("products.id as product_id, products.product_name as product_name,products.price as product_price, SUM(order_items.quantity) as total_sold").
		Joins("JOIN products ON products.id = order_items.product_id").
		Group("products.id, products.product_name").
		Order("total_sold DESC").
		Limit(10).
		Scan(&topProducts).Error; err != nil {
		fmt.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve top selling products",
		})
	}

	if err := database.DB.
        Table("order_items").
        Select("categories.id as category_id, categories.category_name as category_name, SUM(order_items.quantity) as total_sold").
        Joins("JOIN products ON products.id = order_items.product_id").
        Joins("JOIN categories ON categories.id = products.category_id").
        Group("categories.id, categories.category_name").
        Order("total_sold DESC").
        Limit(10).
        Scan(&topCategories).Error; err != nil {
			fmt.Print(err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "failed to retrieve top selling categories",
        })
    }

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Top selling items",
		"products": topProducts,
		"categories": topCategories,
	})
}
