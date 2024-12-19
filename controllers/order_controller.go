package controllers

import (
	"errors"
	"fmt"
	"kars/database"
	"kars/models"
	"log"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type userInput struct {
	AddressId     uint   `json:"address_id"`
	CouponCode    string `json:"coupon_code"`
	PaymentMethod string `json:"payment_method"`
}

func PlaceOrder(c *fiber.Ctx) error {

	userID := c.Locals("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	var input userInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse address id",
		})
	}

	if input.AddressId == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "address id is required",
		})
	}

	var address models.Address
	if err := database.DB.First(&address, "id = ? AND user_id = ?", input.AddressId, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "address not found",
			})
		}
	}

	var cart models.Cart
	if err := database.DB.Preload("CartItems").First(&cart, "user_id = ?", userID).Error; err != nil {
		log.Println("failed to fetch cart details:", err)
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "cart not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve cart details",
		})
	}

	var coupon models.Coupon
	var couponUsage models.CouponUsage
	if input.CouponCode != "" {
		if err := database.DB.Where("coupon_code = ?", input.CouponCode).First(&coupon).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "coupon not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve coupon",
			})
		}

		if !coupon.IsActive {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "coupon is not active",
			})
		}

		if err := database.DB.First(&couponUsage, "user_id = ? AND coupon_code = ?", userID, input.CouponCode).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				couponUsageLimit := models.CouponUsage{
					CouponCode: input.CouponCode,
					UserID:     cart.UserID,
					Limit:      0,
				}

				if err := database.DB.Create(&couponUsageLimit).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "failed to create coupon usage",
					})
				}
			}
		}

		if coupon.UsageLimit <= couponUsage.Limit {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "you exceeded the coupon usage limit",
			})
		}
	}
	totalPrice := 0.0
	for _, item := range cart.CartItems {
		totalPrice += item.TotalPrice
	}

	var discountAmount float64
	if coupon.DiscountType == "fixed" {
		discountAmount = coupon.DiscountValue
	}
	if coupon.DiscountType == "percentage" {
		discountAmount = totalPrice * coupon.DiscountValue / 100
	}

	if coupon.DiscountType == "percentage" {
		if coupon.MaximumDiscount < discountAmount {
			discountAmount = coupon.MaximumDiscount
		}
	}

	finalPrice := 0.0
	shippingAmount := 0.0
	if totalPrice > 500 {
		shippingAmount = 30
	}
	finalPrice = totalPrice + shippingAmount - discountAmount

	var orderStatus string
	var paymentStatus string

	if input.PaymentMethod == "cash on delivery" {

		if finalPrice > 1000 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "cash on delivery is not allowed for orders with a final price greater than 1000",
			})
		}

		orderStatus = "placed"
		paymentStatus = "pending"
	}

	if input.PaymentMethod == "online payment" {
		orderStatus = "pending"
		paymentStatus = "pending"
	}

	if input.PaymentMethod == "wallet" {
		var wallet models.Wallet
		if err := database.DB.First(&wallet, "user_id = ?", userID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "user wallet not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve user wallet",
			})
		}

		if wallet.TotalAmount < finalPrice {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "not enough balance",
			})
		} else {

			newBalance := wallet.TotalAmount - finalPrice
			fmt.Println(newBalance)
			if err := database.DB.Model(&wallet).Update("total_amount", newBalance).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to update wallet balance",
				})
			}

			walletHistory := models.WalletHistory{
				WalletID: wallet.ID,
				Type:     "debit",
				Amount:   finalPrice,
			}

			if err := database.DB.Create(&walletHistory).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to update wallet history",
				})
			}

			paymentStatus = "paid "
			orderStatus = "placed"
		}
	}

	order := models.Order{
		UserID:         cart.UserID,
		TotalPrice:     totalPrice,
		DiscountAmount: discountAmount,
		ShippingAmount: shippingAmount,
		FinalPrice:     finalPrice,
		OrderAddress: models.OrderAddress{
			Name:         address.Name,
			PhoneNo:      address.PhoneNo,
			AddressLine1: address.AddressLine1,
			AddressLine2: address.AddressLine2,
			City:         address.City,
			State:        address.State,
			PostalCode:   address.PostalCode,
			Country:      address.Country,
			LandMark:     address.LandMark,
		},
		PaymentMethod: input.PaymentMethod,
		OrderStatus:   orderStatus,
		PaymentStatus: paymentStatus,
		CouponCode:    input.CouponCode,
	}

	if err := database.DB.Create(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	var orderItems []models.OrderItem
	for _, item := range cart.CartItems {
		orderItems = append(orderItems, models.OrderItem{
			OrderID:      order.ID,
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ProductPrice: item.ProductPrice,
			Quantity:     item.Quantity,
			TotalPrice:   item.TotalPrice,
		})
	}

	if err := database.DB.Create(&orderItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order items",
		})
	}

	for _, item := range orderItems {
		var product models.Product
		if err := database.DB.First(&product, item.ProductID).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch product",
			})
		}

		if product.Quantity < item.Quantity {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Not enough stock for product " + product.ProductName,
			})
		}

		product.Quantity -= item.Quantity
		if err := database.DB.Model(&product).Update("quantity", product.Quantity).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update product quantity",
			})
		}
	}

	if err := database.DB.Delete(&models.Cart{}, "id = ?", cart.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "order place but failed to clear cart",
		})
	}

	if err := database.DB.First(&couponUsage, "user_id = ? AND coupon_code = ?", userID, input.CouponCode).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve coupon usage",
		})
	}

	couponUsage.Limit += 1
	if err := database.DB.Save(&couponUsage).Error; err != nil {
		log.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update the coupon usage limit",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Order successfully placed",
		"order":       order,
		"order_items": orderItems,
	})
}

func CancelOrder(c *fiber.Ctx) error {

	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	var order models.Order
	if err := database.DB.First(&order, "id = ?", orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve order details",
		})
	}

	switch order.OrderStatus {
	case "shipped":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order already shipped and you can't cancel the order",
		})
	case "cancelled":
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order already cancelled",
		})
	}

	var orderItems []models.OrderItem
	if err := database.DB.Find(&orderItems, "order_id = ?", order.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve order items",
		})
	}

	for _, item := range orderItems {
		var product models.Product
		if err := database.DB.First(&product, item.ProductID).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to retrieve product",
			})
		}

		product.Quantity += item.Quantity
		if err := database.DB.Model(&product).Update("quantity", product.Quantity).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update product quantity",
			})
		}
	}

	if order.PaymentStatus == "paid" {
		var wallet models.Wallet

		if err := database.DB.First(&wallet, "user_id = ?", order.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {

				newWallet := models.Wallet{
					UserID:      order.UserID,
					TotalAmount: order.FinalPrice,
				}
				if err := database.DB.Create(&newWallet).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "failed to create wallet",
					})
				}

				walletHistory := models.WalletHistory{
					WalletID: newWallet.ID,
					Type:     "credit",
					Amount:   order.FinalPrice,
				}
				if err := database.DB.Create(&walletHistory).Error; err != nil {
					log.Print(err)
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "failed to update wallet history",
					})
				}
			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to retrieve wallet",
				})
			}
		} else {
			wallet.TotalAmount += order.FinalPrice
			if err := database.DB.Model(&wallet).Update("total_amount", wallet.TotalAmount).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to update wallet",
				})
			}

			walletHistory := models.WalletHistory{
				WalletID: wallet.ID,
				Type:     "credit",
				Amount:   order.FinalPrice,
			}
			if err := database.DB.Create(&walletHistory).Error; err != nil {
				log.Print(err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to update wallet history",
				})
			}
		}
	}

	updates := map[string]interface{}{
		"order_status":   "cancelled",
		"payment_status": "returned",
	}

	if err := database.DB.Model(&order).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update order status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "order successfully cancelled",
		"oreder":  order,
	})
}

func ReturnOrder(c *fiber.Ctx) error {

	userID := c.Locals("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	var order models.Order
	if err := database.DB.First(&order, "id = ?", orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve order",
		})
	}

	if order.OrderStatus == "cancelled" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order already cancelled",
		})
	}

	if order.OrderStatus == "shipped" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "you can't return the order",
		})
	}

	if order.OrderStatus == "placed" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order not shipped, you can cancel the order instead",
		})
	}

	if order.OrderStatus == "delivered" {
		var wallet models.Wallet
		if err := database.DB.First(&wallet, "user_id = ?", order.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				newWallet := models.Wallet{
					UserID:      order.UserID,
					TotalAmount: order.FinalPrice,
				}

				if err := database.DB.Create(&newWallet).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "failed to create wallet",
					})
				}

				walletHistory := models.WalletHistory{
					WalletID: wallet.ID,
					Type:     "credit",
					Amount:   order.FinalPrice,
				}

				if err := database.DB.Create(&walletHistory).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "failed to update wallet history",
					})
				}

			} else {
				wallet.TotalAmount += order.FinalPrice
				if err := database.DB.Save(&wallet).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "failed to udpate wallet",
					})
				}

				walletHistory := models.WalletHistory{
					WalletID: wallet.ID,
					Type:     "credit",
					Amount:   order.FinalPrice,
				}

				if err := database.DB.Create(&walletHistory).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "failed to update wallet history",
					})
				}

			}
		}

		order.OrderStatus = "returned"
		order.PaymentStatus = "returned"
		if err := database.DB.Save(&order).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed update order",
			})
		}
	}

	if err := database.DB.First(&order, "id = ?", orderID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve updated order",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":       "order successfully updated",
		"updated_order": order,
	})
}

func ListOrdersForUser(c *fiber.Ctx) error {

	userid := c.Locals("user_id")
	if userid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	var orders []models.Order
	if err := database.DB.Preload("OrderItems").Order("created_at DESC").Find(&orders, "user_id", userid).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrive orders",
		})
	}

	if len(orders) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no orders found",
		})
	}

	var response []fiber.Map

	for _, order := range orders {
		orderResponse := fiber.Map{
			"order_id":        order.ID,
			"total_price":     order.TotalPrice,
			"final_price":     order.FinalPrice,
			"address":         order.OrderAddress,
			"discount_amount": order.DiscountAmount,
			"payment_method":  order.PaymentMethod,
			"payment_status":  order.PaymentStatus,
			"order_status":    order.OrderStatus,
			"created_at":      order.CreatedAt,
			"items":           []fiber.Map{},
		}

		for _, item := range order.OrderItems {
			itemResponse := fiber.Map{
				"product_id":    item.ProductID,
				"product_name":  item.ProductName,
				"product_price": item.ProductPrice,
				"quantiy":       item.Quantity,
				"total_price":   item.TotalPrice,
			}

			orderResponse["items"] = append(orderResponse["items"].([]fiber.Map), itemResponse)
		}

		response = append(response, orderResponse)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Orders retrieved successfully.",
		"orders":  response,
	})
}

func GetWallet(c *fiber.Ctx) error {

	userID := c.Locals("user_id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	var wallet models.Wallet
	if err := database.DB.First(&wallet, "user_id = ?", userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "wallet not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve wallet",
		})
	}

	var walletHistory []models.WalletHistory
	if err := database.DB.Find(&walletHistory, "wallet_id = ?", wallet.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "wallet history not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve wallet history",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully get wallet",
		"wallet":  wallet,
		"history": walletHistory,
	})
}

func CancelOneProduct(c *fiber.Ctx) error {

	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	productID := c.Params("product_id")
	if productID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "product id is required",
		})
	}

	var order models.Order
	if err := database.DB.First(&order, "id = ?", orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve order",
		})
	}

	if order.OrderStatus == "cancelled" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order already cancelled",
		})
	}

	if order.OrderStatus == "shipped" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order shipped you can't cancel the order",
		})
	}

	var orderItems models.OrderItem
	if err := database.DB.First(&orderItems, "order_id = ? AND product_id = ?", orderID, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "order item not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve order item",
		})
	}

	var product models.Product
	if err := database.DB.First(&product, "id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "product not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve product",
		})
	}

	if orderItems.IsCancelled == "cancelled" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order item already cancelled",
		})
	}

	if orderItems.IsCancelled == "ordered" {
		if err := database.DB.Model(&orderItems).Update("is_cancelled", "cancelled").Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed cancel order item",
			})
		}

		var returnPrice, newDiscountAmount, finalPrice float64
		amount := order.TotalPrice - orderItems.ProductPrice
		var couponCode string

		if order.CouponCode != "" {

			var coupon models.Coupon
			if err := database.DB.First(&coupon, "coupon_code = ?", order.CouponCode).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
						"error": "coupon not found",
					})
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to retireve coupon",
				})
			}

			if coupon.MinimumAmount > amount {
				if coupon.DiscountType == "percentage" || coupon.DiscountType == "fixed" {
					returnPrice = order.FinalPrice - amount
					finalPrice = amount
					newDiscountAmount = 0
					couponCode = ""

					var couponUsage models.CouponUsage
					if err := database.DB.First(&couponUsage, "user_id", order.UserID).Error; err != nil {
						if err == gorm.ErrRecordNotFound {
							return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
								"error": "coupon usage limit not found",
							})
						}
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"error": "failed to retrieve coupon usage",
						})
					}

					couponUsage.Limit--
					if err := database.DB.Model(&couponUsage).Update("limit", couponUsage.Limit).Error; err != nil {
						return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
							"error": "failed to update coupon usage",
						})
					}
				}
			} else {
				if coupon.DiscountType == "percentage" {
					newDiscountAmount = amount * coupon.DiscountValue / 100
					finalPrice = amount - newDiscountAmount
					returnPrice = order.FinalPrice - finalPrice
					couponCode = order.CouponCode
				} else if coupon.CouponCode == "fixed" {
					finalPrice = order.FinalPrice - product.Price
					returnPrice = product.Price
					newDiscountAmount = coupon.DiscountValue
					couponCode = order.CouponCode
				}
			}

		} else if order.CouponCode == "" {
			returnPrice = product.Price
			finalPrice = order.TotalPrice - product.Price
			newDiscountAmount = 0
			couponCode = ""
		}

		updates := map[string]interface{}{
			"final_price":     finalPrice,
			"discount_amount": newDiscountAmount,
			"total_price":     amount,
			"coupon_code":     couponCode,
		}

		fmt.Println(updates)

		if err := database.DB.Model(&order).Updates(updates).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update order",
			})
		}

		var wallet models.Wallet

		if err := database.DB.Where("user_id = ?", order.UserID).First(&wallet).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newWallet := models.Wallet{
					UserID:      order.UserID,
					TotalAmount: returnPrice,
				}

				if err := database.DB.Create(&newWallet).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to create wallet",
					})
				}

				walletHistory := models.WalletHistory{
					WalletID: wallet.ID,
					Type:     "credit",
					Amount:   returnPrice,
				}

				if err := database.DB.Create(&walletHistory).Error; err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "failed to update wallet history",
					})
				}

			} else {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to query wallet",
				})
			}
		} else {
			wallet.TotalAmount += returnPrice
			if err := database.DB.Save(&wallet).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to update wallet",
				})
			}

			walletHistory := models.WalletHistory{
				WalletID: wallet.ID,
				Type:     "credit",
				Amount:   returnPrice,
			}

			if err := database.DB.Create(&walletHistory).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "failed to update wallet history",
				})
			}
		}
	}

	quantity := orderItems.Quantity + product.Quantity

	if err := database.DB.Model(&product).Update("quantity", quantity).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update product quantity",
		})
	}

	if err := database.DB.First(&order, "id = ?", orderID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve updated order",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully cancel the product",
		"order":   order,
	})
}
