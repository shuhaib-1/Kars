package controllers

import (
	"kars/database"
	"kars/models"
	"strconv"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func AddToCart(c *fiber.Ctx) error{
	 
	userid := c.Params("user_id")
	if userid == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	productid := c.Params("product_id")
	if productid == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	uid, err := strconv.ParseUint(userid, 10,32)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid user id format",
		})
	}
	
	pid, err := strconv.ParseUint(productid, 10,32)
	if err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid product id format",
		})
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", userid).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve user",
		})
	}


	var product models.Product
	if err := database.DB.Where("id = ?", pid).First(&product).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch product",
		})
	}

	if product.Quantity <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product is out of stock",
		})
	}

	var category models.Category
	if err := database.DB.Where("id = ?", product.CategoryID).First(&category).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "category not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve category",
		})
	}

	var productPrice float64
	if category.OfferValue < product.OfferValue {
		if product.OfferType == "percentage" {
			discountPrice := product.Price * product.OfferValue / 100
			productPrice = product.Price - discountPrice
		} else if product.OfferType == "fixed" {
			productPrice = product.Price - product.OfferValue
		}
	} else {
		if category.OfferType == "percentage" {
			discountPrice := product.Price * category.OfferValue / 100
			productPrice = product.Price - discountPrice
		} else if product.Category.OfferType == "fixed" {
			productPrice = product.Price - category.OfferValue
		}
	}

	if productPrice <= 0 {
		productPrice = product.Price
	}

	var cart models.Cart
    if err = database.DB.Where("user_id = ?", uid).First(&cart).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            cart = models.Cart{
                UserID:    uint(uid),
                TotalItems: 0,
            }
            if err := database.DB.Create(&cart).Error; err != nil {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": "failed to create cart",
              })
            }
        } else {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
               "error": "failed to retrieve cart",
            })
        }
    }

	var cartitem models.CartItem

	if err := database.DB.Where("cart_id = ? AND product_id = ?",cart.ID, pid).First(&cartitem).Error; err != nil{
		
		cartitem = models.CartItem{
			CartID: cart.ID,
			ProductID: uint(pid),
			ProductName: product.ProductName,
			ProductPrice: productPrice,
			Quantity: 1,
			TotalPrice: productPrice,
		}
		if err := database.DB.Create(&cartitem).Error; err != nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to add the product to cart",
			})
		}
		cart.TotalItems++
	} else{

		if cartitem.Quantity >= product.Quantity {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "cannot add more of this product; maximum stock reached",
			})
		}

		cartitem.Quantity++
		cartitem.TotalPrice = float64(cartitem.Quantity) * productPrice
		if err := database.DB.Save(&cartitem).Error; err != nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to save the cart",
			})
		}
	}

	if err := database.DB.Model(&cart).Update("TotalItems", cart.TotalItems).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update cart",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "product added to the cart successfully",
		"cart": cart,
		"cart_items": cartitem,
	})

}
func RemoveFromCart(c *fiber.Ctx) error {
	
	userid := c.Locals("user_id")
	if userid == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	productid := c.Params("product_id")
	if userid == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	var cart models.Cart
	if err := database.DB.Where("user_id", userid).First(&cart).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "failed to find the user in the cart",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve cart",
		})
	}

	var cartitem models.CartItem
	if err := database.DB.Where("cart_id = ? AND product_id = ?", cart.ID, productid).First(&cartitem).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found in the cart",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve cart item",
		})
	}

	if err := database.DB.Delete(&cartitem).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to remove product from cart",
		})
	}

	cart.TotalItems--
	if err := database.DB.Model(&cart).Update("TotalItems", cart.TotalItems).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update cart",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Product successfully removed from the cart",
		"cart": cart,
		"cart_items": cartitem,
	})
}

func ListCartProducts(c *fiber.Ctx) error {

	userid := c.Locals("user_id")
	if userid == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	var cart models.Cart
	if err := database.DB.First(&cart,"user_id", userid).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "cart not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve cart details",
		})
	}

	var cartItems []models.CartItem
	if err := database.DB.Find(&cartItems, "cart_id", cart.ID).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve cart items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Carts products",
		"cart": cartItems,
	})
}