package controllers

import (
	"kars/database"
	"kars/models"
	"log"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Product struct {
	ProductName string  `json:"product_name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	CategoryID  uint    `json:"category_id"`
	Color       string  `json:"color"`
	ImgURLs     string  `json:"img_urls"`
	Status      string  `json:"status"`
	OfferType   string  `json:"offer_type"`
	OfferValue  float64 `json:"offer_value"`
}

func AddProduct(c *fiber.Ctx) error {

	var input Product
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Input",
		})
	}

	if err := input.ProductValidate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var existingProduct models.Product
	if err := database.DB.Where("product_name = ?", input.ProductName).First(&existingProduct).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "Product already exits",
		})
	} else if err != gorm.ErrRecordNotFound {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check existing product",
		})
	}

	var existingCategory models.Category
	if err := database.DB.First(&existingCategory, input.CategoryID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Category not found",
		})
	}

	NewProduct := models.Product{
		ProductName: input.ProductName,
		Description: input.Description,
		Price:       input.Price,
		Quantity:    input.Quantity,
		CategoryID:  input.CategoryID,
		Color:       input.Color,
		ImgURLs:     input.ImgURLs,
		Status:      input.Status,
		OfferType:   input.OfferType,
		OfferValue:  input.OfferValue,
	}

	if err := database.DB.Create(&NewProduct).Error; err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Product failed to create",
		})
	}

	var categoryName string
	if err := database.DB.Model(&models.Category{}).Select("category_name").Where("id = ?", NewProduct.CategoryID).Scan(&categoryName).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load category name",
		})
	}

	response := fiber.Map{
		"message":     "Product is successfully created",
		"ID":          NewProduct.ID,
		"CategoryId":  NewProduct.CategoryID,
		"ProductName": NewProduct.ProductName,
		"Description": NewProduct.Description,
		"Price":       NewProduct.Price,
		"Quantity":    NewProduct.Quantity,
		"Category":    categoryName,
		"Color":       NewProduct.Color,
		"ImgURLs":     NewProduct.ImgURLs,
		"Status":      NewProduct.Status,
		"OfferType":   NewProduct.OfferType,
		"OfferValue":  NewProduct.OfferValue,
		"IsListed":    NewProduct.IsListed,
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":         "Product is successfully created",
		"Product_details": response,
	})
}

func EditProduct(c *fiber.Ctx) error {

	id := c.Params("product_id")
	if id == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "id is required",
		})
	}

	var input Product
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Input",
		})
	}

	var product models.Product

	if err := database.DB.Where("id = ?", id).First(&product).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Product not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve product",
		})
	}

	if input.ProductName != "" && input.ProductName != product.ProductName {
		var existingProduct models.Product
		err := database.DB.Where("LOWER(product_name) = LOWER(?) AND id != ?", input.ProductName, id).First(&existingProduct).Error
		if err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "A product with the same name already exists",
			})
		} else if err != gorm.ErrRecordNotFound {
			log.Printf("Error checking product name: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check product name",
			})
		}
		product.ProductName = input.ProductName
	}

	var categoryname string
	if input.CategoryID != 0 && input.CategoryID != product.CategoryID {
		var category models.Category
		if err := database.DB.First(&category, input.CategoryID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "Category not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to validate category",
			})
		}
		product.CategoryID = input.CategoryID
		categoryname = category.CategoryName
	} else {
		var category models.Category
		if err := database.DB.Select("category_name").Where("id = ?", product.CategoryID).First(&category).Error; err == nil {
			categoryname = category.CategoryName
		}
	}

	if err := input.ProductValidateForPatch(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if input.ProductName != "" {
		product.ProductName = input.ProductName
	}
	if input.Description != "" {
		product.Description = input.Description
	}
	if input.Price != 0 {
		product.Price = input.Price
	}
	if input.Quantity > 0 {
		product.Quantity = input.Quantity
	}
	if input.Color != "" {
		product.Color = input.Color
	}
	if input.ImgURLs != "" {
		product.ImgURLs = input.ImgURLs
	}
	if input.Status != "" {
		product.Status = input.Status
	}
	if input.OfferType != "" {
		product.OfferType = input.OfferType
	}
	if input.OfferValue > 0 {
		product.OfferValue = input.OfferValue
	}

	if err := database.DB.Save(&product).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update product",
		})
	}

	response := fiber.Map{
		"id":            product.ID,
		"product_name":  product.ProductName,
		"description":   product.Description,
		"price":         product.Price,
		"quantity":      product.Quantity,
		"color":         product.Color,
		"img_urls":      product.ImgURLs,
		"status":        product.Status,
		"offer_type":    product.OfferType,
		"offer_value":   product.OfferValue,
		"category_name": categoryname,
		"category_id":   product.CategoryID,
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "Product updated successfully",
		"product_details": response,
	})
}

func DeleteProduct(c *fiber.Ctx) error {

	id := c.Params("product_id")
	if id == "" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	var product models.Product
	result := database.DB.First(&product, id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Product not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve product",
		})
	}

	if product.IsListed == "listed" {
		product.IsListed = "unlisted"
	} else {
		product.IsListed = "listed"
	}

	saveResult := database.DB.Save(&product)
	if saveResult.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save product",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "Product successfully updated",
		"product_details": product,
	})
}

func UserProductList(c *fiber.Ctx) error {

	var productlist []models.Product

	if err := database.DB.Preload("Category", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_listed = ?", "listed")
	}).Where("is_listed = ?", "listed").Find(&productlist).Error; err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch products details",
		})
	}

	if len(productlist) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "No products found",
		})
	}

	var response []fiber.Map
	for _, product := range productlist {
		finalPrice := product.Price

		if product.Category.OfferValue > product.OfferValue {
			if product.Category.OfferType == "percentage" {
				finalPrice -= finalPrice * product.Category.OfferValue / 100
			} else if product.Category.OfferType == "fixed" {
				finalPrice -= product.Category.OfferValue
			}
		} else {
			if product.OfferType == "percentage" {
				finalPrice -= finalPrice * product.OfferValue / 100
			} else if product.OfferType == "fixed" {
				finalPrice -= product.OfferValue
			}
		}

		if finalPrice < 0 {
			finalPrice = 0
		}

		productMap := fiber.Map{
			"product_id": product.ID,
			"product_name":  product.ProductName,
			"description":   product.Description,
			"price":         product.Price,
			"final_price":   finalPrice,
			"quantity":      product.Quantity,
			"color":         product.Color,
			"status":        product.Status,
			"category_id": product.Category.ID,
			"category_name": product.Category.CategoryName,
			"img_urls":      product.ImgURLs,
		}
		response = append(response, productMap)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":         "successfully fetched all the products details",
		"product_details": response,
	})
}
