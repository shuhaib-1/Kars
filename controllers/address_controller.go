package controllers

import (
	"fmt"
	"kars/database"
	"kars/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type NewAddress struct {
	UserID       uint   `json:"user_id"`
	Name         string `json:"name"`
	PhoneNo      string `json:"phone_no"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	LandMark     string `json:"land_mark"`
	AddressType  string `json:"address_type"`
}

func UserAddAddress(c *fiber.Ctx) error {

	UserId, err := convertToUint(c.Locals("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid user_id: %s", err.Error()),
		})
	}

	var ExistingUser models.User
	if err := database.DB.Where("id = ?", UserId).First(&ExistingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "User not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database error",
		})
	}

	var Address NewAddress
	if err := c.BodyParser(&Address); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse address",
		})
	}

	Address.UserID = UserId

	if err := Address.AddressValidate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	address := models.Address{
		UserID:       Address.UserID,
		Name:         Address.Name,
		PhoneNo:  Address.PhoneNo,
		AddressLine1: Address.AddressLine1,
		AddressLine2: Address.AddressLine2,
		City:         Address.City,
		State:        Address.State,
		PostalCode:   Address.PostalCode,
		Country:      Address.Country,
		LandMark:     Address.LandMark,
		AddressType:  Address.AddressType,
	}

	
	if err := database.DB.Create(&address).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create address",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Successfully added address",
		"address": address,
	})
}

func UserEditAddress(c *fiber.Ctx) error {

	UserId, err := convertToUint(c.Locals("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid user_id: %s", err.Error()),
		})
	}
	
	id := c.Params("address_id")
	if id == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "address id is required",
		})
	}

	var address NewAddress
	if err := c.BodyParser(&address); err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse address",
		})
	}

	if err :=  address.AddressValidateForPatch(); err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	fields := map[string]interface{}{}
	if address.Name != ""{
		fields["name"] = address.Name
	}
	if address.PhoneNo != ""{
		fields["phone_no"]  = address.PhoneNo
	}
	if address.AddressLine1 != ""{
		fields["address_line1"] = address.AddressLine1
	}
	if address.AddressLine2 != ""{
		fields["address_line2"] = address.AddressLine2
	}
	if address.City != ""{
		fields["city"] = address.City
	}
	if address.State != ""{
		fields["state"] = address.State
	}
	if address.PostalCode != ""{
		fields["postal_code"] = address.PostalCode
	}
	if address.Country != ""{
		fields["country"] = address.Country
	}
	if address.LandMark != ""{
		fields["land_mark"] = address.LandMark
	}
	if address.AddressType != ""{
		fields["address_type"] = address.AddressType
	}

	var NewAddress models.Address

	if err := database.DB.Model(&NewAddress).Where("id = ? AND user_id = ?", id, UserId).Updates(&fields).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update address",
		})
	}

	var UpdateAddress models.Address

	if err := database.DB.First(&UpdateAddress, "id = ? And user_id = ?", id, UserId).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated address",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Address updated successfully",
		"updated address": UpdateAddress,
	})
 
}

func UserDeleteAddress(c *fiber.Ctx) error {

	UserId, err := convertToUint(c.Locals("user_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid user_id: %s", err.Error()),
		})
	}

	addressid := c.Params("address_id")
	if addressid == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":"address id is required",
		})
	}

	var address models.Address
	if err := database.DB.Where("id = ? AND user_id = ?", addressid,UserId).First(&address).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "address nof found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch address",
		})
	}

	if err := database.DB.Delete(&address).Error; err != nil{
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete address",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "address deleted successfully",
	})
}

func UserListAddress(c *fiber.Ctx) error {
	userid := c.Locals("user_id")
	if userid == ""{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "user id is required",
		})
	}

	var addresses []models.Address
	if err := database.DB.Where("user_id = ?", userid).Find(&addresses).Error; err != nil{
		if err == gorm.ErrRecordNotFound{
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch user addresses",
		})
	}

	if len(addresses) == 0{
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "no addresses found for this user",
		})
	}

	var response []fiber.Map
	for _,address := range addresses{
		response = append(response, fiber.Map{
			"address_id": address.ID,
			"name": address.Name,
			"phone_no": address.PhoneNo,
			"address_line1": address.AddressLine1,
			"address_line2": address.AddressLine2,
			"city": address.City,
			"state": address.State,
			"postal_code": address.PostalCode,
			"country": address.PostalCode,
			"land_mark": address.LandMark,
			"address_type": address.AddressType,
			"created_at": address.CreatedAt,
		})
	}


	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user addresses retrieved successfully",
		"addresses": response,	
	})
}