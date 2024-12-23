package controllers

import (
	"errors"
	"fmt"
	"kars/database"
	"kars/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Category struct {
	Name string `json:"category_name"`
	OfferType string `json:"offer_type"`
	OfferValue float64 `json:"offer_value"`
}

func AddCategory(c *fiber.Ctx) error {

	var input Category

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Input",
		})
	}

	if input.Name == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	if input.OfferType != "" {
		if input.OfferType != "percentage" && input.OfferType != "fixed" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "discount type must be either 'percentage' or 'fixed'",
			})
		}
	}

	if input.OfferValue < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "offer value must be greater than 0",
		})
	}

	if input.OfferType != "" && input.OfferValue > 0 {
		if input.OfferType == "percentage" && input.OfferValue > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "percentage discount cannot exceed 100%",
			})
		}
	}

	if input.OfferType != "" && input.OfferValue <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "you have add the both offer type and offer value",
		})
	}

	if input.OfferType == "" && input.OfferValue > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "you have add the both offer type and offer value",
		})
	}


	var existingCategory models.Category
	err := database.DB.First(&existingCategory, "LOWER(category_name) = LOWER(?)", input.Name).Error
	if err == nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "category already exists",
		})
	}

	if !errors.Is(err, gorm.ErrRecordNotFound){
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to check existing product",
		})
	}

	newCategory := models.Category{
		CategoryName: input.Name,
		OfferType: input.OfferType,
		OfferValue: input.OfferValue,
	}

	if err := database.DB.Create(&newCategory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create categroy",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Categroy created successfully",
		"category": newCategory,
	})
}

func EditCategory(c *fiber.Ctx) error {

	id := c.Params("category_id")
	if id == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "category id is required",
		})
	}

	var input Category

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Invalid Input",
		})
	}


	var category models.Category
	if err := database.DB.Where("id = ?", id).First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Category not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve category",
		})
	}

	if input.Name != "" && input.Name != category.CategoryName{
		var existingCategory models.Category
		if err := database.DB.Where("LOWER(category_name) = LOWER(?) AND id != ?", input.Name, id).First(&existingCategory).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Category with this name already exists",
			})
		}
		category.CategoryName = input.Name
	}

	if input.OfferType != "" || input.OfferValue != 0 {
		if input.OfferType != "" {
			if input.OfferType != "percentage" && input.OfferType != "fixed" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "offertype must be either 'percentage' or 'fixed'",
				})
			}
			category.OfferType = input.OfferType
		}

		if input.OfferValue < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "offervalue must be greater than or equal to 0",
			})
		}

		if input.OfferType == "percentage" && input.OfferValue > 100 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "percentage discount cannot exceed 100%",
			})
		}

		category.OfferValue = input.OfferValue
	}

	if err := database.DB.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Category failed to save",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Category updated successfully",
		"categroy": category,
	})
}

func DeleteCategory(c *fiber.Ctx) error {

	id := c.Params("category_id")
	if id == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "category id is required",
		})
	}

	var category models.Category

	result := database.DB.First(&category, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Category not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve category",
		})
	}

	if category.IsListed == "listed"{
		category.IsListed = "unlisted"
	}else{
		category.IsListed = "listed"
	}

	fmt.Println(category)

	
	saveResult := database.DB.Save(&category)
	if saveResult.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save category",
		})
	}

	response := fiber.Map{
		"id": category.ID,
		"name": category.CategoryName,
		"is_listed": category.IsListed,
		"created_at": category.CreatedAt,
		"updated_at": category.UpdatedAt,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "category listing updated sucessfully",
		"category": response,
	})
}
