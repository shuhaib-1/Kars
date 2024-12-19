package controllers

import (
	"fmt"
	"kars/database"
	"kars/models"

	"github.com/gofiber/fiber/v2"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

var count uint

func RenderRayzorPay(c *fiber.Ctx) error {
	orderId := c.Query("order_id")
	if orderId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "order id is required"})
	}
	fmt.Print(orderId)
	return c.Render("templates/payment.html", fiber.Map{
		"order_id": orderId, // Replace with any necessary data
	})
}

// Create an order in Razorpay
func CreateOrder(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "order id is required"})
	}

	client := razorpay.NewClient("rzp_test_J7PdmiC5AUYYOX", "osI3mJcQapZ2KQbOi3iemNew")

	var order models.Order
	if err := database.DB.Where("id = ?", orderID).First(&order).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "order not found"})
	}

	if order.PaymentStatus == "paid" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"error": "order already paid"})
	}

	amount := int(order.FinalPrice * 100)
	fmt.Print(amount)
	data := map[string]interface{}{
		"amount":   amount,
		"currency": "INR",
		"receipt":  "order_rcptid_11",
	}

	razorOrder, err := client.Order.Create(data, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "inable to create order"})
	}

	return c.JSON(razorOrder)
}

// Verify Razorpay payment
func VerifyPayment(c *fiber.Ctx) error {
	orderID := c.Params("order_id")
	fmt.Println(orderID)
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "order id is required"})
	}

	var PaymentInfo struct {
		RazorpayPaymentID string `json:"razorpay_payment_id"`
		RazorpayOrderID   string `json:"razorpay_order_id"`
		RazorpaySignature string `json:"razorpay_signature"`
	}
	if err := c.BodyParser(&PaymentInfo); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request data"})
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if tx.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to start transaction"})
	}

	if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Update("payment_status", "paid").Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
	}

	if err := tx.Model(&models.Order{}).Where("id = ?", orderID).Update("order_status", "placed").Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
	}

	var orderItems []models.OrderItem
	if err := tx.Where("id = ?", orderID).Find(&orderItems).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch order items"})
	}
	for _, item := range orderItems {
		if err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).Update("quantity", gorm.Expr("quantity - ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update product stock"})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transaction commit failed"})
	}

	return c.JSON(fiber.Map{"status": "success", "message": "Payment verified and recorded successfully"})
}

func FailedHandling(c *fiber.Ctx) error {
	fmt.Println("poda panny")
	orderID := c.Params("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid order ID"})
	}

	var order models.Order
	if count >= 3 {
		if err := database.DB.Model(&order).Where("id = ?", orderID).Update("payment_status", "failed").Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
		}
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Too many failed attempts"})
	}

	if err := database.DB.Model(&order).Where("id = ?", orderID).Update("payment_status", "failed").Error; err != nil {
		count++
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update order status"})
	}

	count = 0
	return c.JSON(fiber.Map{"message": "Payment status updated to failed successfully"})
}