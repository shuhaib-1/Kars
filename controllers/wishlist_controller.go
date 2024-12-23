package controllers

import (
	"kars/database"
	"kars/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddWishList(c *fiber.Ctx) error {

	userID := c.Locals("user_id")
	if userID == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	productID := c.Params("product_id")
	if productID == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product is required",
		})
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve user",
		})
	}

	var product models.Product
	if err := database.DB.First(&product, "id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not founc",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve product",
		})
	}

	var wishList models.Wishlist
	if err := database.DB.First(&wishList, "user_id = ? AND product_id = ?", userID, productID).Error; err == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product already exist in the wishlist",
		})
	}

	wishList = models.Wishlist{
		UserID:             user.ID,
		ProductID:          product.ID,
		ProductName:        product.ProductName,
		ProductDescription: product.Description,
		ProductPrice:       product.Price,
	}

	if err := database.DB.Create(&wishList).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to add the product to the wishlist",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "product successfully added to the wishlist",
	})
}

func RemoveFromWishList(c *fiber.Ctx) error {

	userID := c.Locals("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	productID := c.Params("product_id")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	var wishList models.Wishlist
	if err := database.DB.First(&wishList, "user_id = ? AND product_id = ?", userID, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found in the wishlist",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve wishlist",
		})
	}

	if err := database.DB.Delete(&wishList).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to remove product from the wishlist",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "product successfully removed from the wishlist",
	})
}
func ListWishList(c *fiber.Ctx) error {

	userID := c.Locals("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id required",
		})
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve user",
		})
	}

	var wishList []models.Wishlist
	if err := database.DB.Find(&wishList, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "not found any products",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve wishlist",
		})
	}

	var response []fiber.Map
	for _, item := range wishList {
		response = append(response, fiber.Map{
			"user_id":            item.UserID,
			"product_id":         item.ProductID,
			"product_decription": item.ProductDescription,
			"product_name":       item.ProductName,
			"product_price":      item.ProductPrice,
		})
	}

	if len(response) == 0{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no wishlist items found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "successfully retrieved wishlist",
		"wishlist": response,
	})
}
