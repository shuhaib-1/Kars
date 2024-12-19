package controllers

import (
	"encoding/json"
	"fmt"
	"kars/database"
	"kars/models"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const customDateFormat = "2006-01-02"

func (c *Coupon) UnmarshalJSON(data []byte) error {
	type Alias Coupon
	aux := &struct {
		StartDate  string `json:"start_date"`
		ExpiryTime string `json:"expiry_time"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.StartDate != "" {
		parsedStartDate, err := time.Parse(customDateFormat, aux.StartDate)
		if err != nil {
			return fmt.Errorf("invalid start_date format: %v", err)
		}
		c.StartDate = parsedStartDate
	}

	if aux.ExpiryTime != "" {
		parsedExpiryTime, err := time.Parse(customDateFormat, aux.ExpiryTime)
		if err != nil {
			return fmt.Errorf("invalid expiry_time format: %v", err)
		}
		c.ExpiryTime = parsedExpiryTime
	}

	return nil
}

type Coupon struct {
	CouponName      string    `json:"coupon_name"`
	CouponCode      string    `json:"coupon_code"`
	DiscountType    string    `json:"discount_type"`
	DiscountValue   float64   `json:"discount_value"`
	MaximumDiscount float64   `json:"maximum_discount"`
	MinimumAmount   float64   `json:"minimum_amount"`
	UsageLimit      int       `json:"usage_limit"`
	StartDate       time.Time `json:"start_date"`
	ExpiryTime      time.Time `json:"expiry_time"`
	IsActive        bool      `json:"is_active"`
}

func AddCoupon(c *fiber.Ctx) error {

	var coupon Coupon
	if err := c.BodyParser(&coupon); err != nil {
		log.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse coupon details",
		})
	}

	if err := coupon.ValidateCoupon(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	var existingCoupon models.Coupon

	if err := database.DB.Where("coupon_code = ?", coupon.CouponCode).First(&existingCoupon).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "coupon code already exists",
		})
	}

	if err := database.DB.Where("coupon_name = ?", coupon.CouponName).First(&existingCoupon).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "coupon name already exists",
		})
	}

	newCoupon := models.Coupon{
		CouponName:      coupon.CouponName,
		CouponCode:      coupon.CouponCode,
		DiscountType:    coupon.DiscountType,
		DiscountValue:   coupon.DiscountValue,
		MaximumDiscount: coupon.MaximumDiscount,
		MinimumAmount:   coupon.MinimumAmount,
		UsageLimit:      coupon.UsageLimit,
		StartDate:       coupon.StartDate,
		ExpiryTime:      coupon.ExpiryTime,
		IsActive:        coupon.IsActive,
	}

	if err := database.DB.Create(&newCoupon).Error; err != nil {
		log.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create coupon",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully created coupon",
		"coupon":  newCoupon,
	})
}

func EditCoupon(c *fiber.Ctx) error {

	couponID := c.Params("coupon_id")
	if couponID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "coupon id is required",
		})
	}

	var coupon models.Coupon
	if err := database.DB.First(&coupon, "id = ?", couponID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "coupon not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve coupon",
		})
	}

	var updateData Coupon
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to parse",
		})
	}

	if err := updateData.ValidateCouponForPatch(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	fields := map[string]interface{}{}

	if updateData.CouponName != "" {
		fields["coupon_name"] = updateData.CouponName
	}
	if updateData.CouponCode != "" {
		fields["coupon_code"] = updateData.CouponCode
	}
	if updateData.DiscountType != "" {
		fields["discount_type"] = updateData.DiscountType
	}
	if updateData.DiscountValue > 0 {
		fields["discount_value"] = updateData.DiscountValue
	}
	if updateData.MaximumDiscount > 0 {
		fields["maximum_discount"] = updateData.MaximumDiscount
	}
	if updateData.MinimumAmount > 0 {
		fields["minimum_amount"] = updateData.MinimumAmount
	}
	if updateData.UsageLimit > 0 {
		fields["usage_limit"] = updateData.UsageLimit
	}
	if !updateData.StartDate.IsZero() {
		fields["start_date"] = updateData.StartDate
	}
	if !updateData.ExpiryTime.IsZero() {
		fields["expiry_time"] = updateData.ExpiryTime
	}
	if !updateData.IsActive{
		fields["is_active"] = updateData.IsActive
	}
	if updateData.IsActive { 
		fields["is_active"] = updateData.IsActive
	}

	if len(fields) > 0{
		if err := database.DB.Model(&coupon).Where("id = ?", couponID).Updates(&fields).Error; err != nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to update coupon",
			})
		}
	}

	var updatedCoupon models.Coupon
	if err := database.DB.First(&updatedCoupon, "id = ?", couponID).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve updated coupon",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "coupon successfully updated",
		"updated_coupon": updatedCoupon,
	})
}

func DeleteCoupon(c *fiber.Ctx) error {

	couponID := c.Params("coupon_id")
	if couponID == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "coupon id is required",
		})
	}

	var coupon models.Coupon
	if err := database.DB.First(&coupon, "id = ?", couponID).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "coupon not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve coupon",
		})
	}

	if err := database.DB.Delete(&coupon).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete coupon",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "coupon successfully deleted",
	})
}

func CancelCoupon(c *fiber.Ctx) error {

	orderID := c.Params("order_id")
	if orderID == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	var order models.Order
	if err := database.DB.First(&order, "id = ?", orderID).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve order",
		})
	}

	order.FinalPrice += order.DiscountAmount
	order.DiscountAmount = 0
	order.CouponCode = ""

	if err := database.DB.Save(&order).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update coupon",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "coupon successfully cancelled",
		"updated_order": order,
	})
}
