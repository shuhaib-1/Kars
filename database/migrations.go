package database

import (
	"kars/models"
	"log"
)
func MigrateModels() error {
	//user model migration
	if err := DB.AutoMigrate(&models.User{}); err != nil {
		log.Println("Failed to migrate User model:", err)
		return err
	}

    //admin model migration
	if err := DB.AutoMigrate(&models.Admin{}); err != nil {
		log.Println("Failed to migrate admin model:", err)
		return err
	}

	if err := DB.AutoMigrate(&models.Category{}); err != nil {
		log.Println("Failed to migrate Category model:", err)
	} else {
		log.Println("Category model migration was successful")
	}

	if err := DB.AutoMigrate(&models.Product{}); err != nil {
		log.Println("Failed to migrate product model:", err)
	} else {
		log.Println("Product model migration was successful")
	}

	if err := DB.AutoMigrate(&models.Address{}); err != nil{
		log.Println("Failed to migrate address model:", err)
	}else{
		log.Println("Address model migration was successfull")
	}

	if err := DB.AutoMigrate(&models.Cart{}); err != nil{
		log.Println("Failed to migrate cart model:", err)
	}else{
		log.Println("Cart model migration was successfull")
	}

	if err := DB.AutoMigrate(&models.CartItem{}); err != nil{
		log.Println("Failed to migrate cart item model:", err)
	}else{
		log.Println("Cart item model migration was successfull")
	}

	if err := DB.AutoMigrate(&models.Order{}); err != nil{
		log.Println("Failed to migrate order model:", err)
	}else{
		log.Println("order model migration was successfull")
	}

	if err := DB.AutoMigrate(&models.OrderItem{}); err != nil{
		log.Println("Failed to migrate order item model:", err)
	}else{
		log.Println("order item model migration was successfull")
	}

	if err := DB.AutoMigrate(&models.Wishlist{}); err != nil{
		log.Println("Failed to migrate wishlist model:", err)
	}else{
		log.Println("wishlist model migration was successfull")
	}

	if err := DB.AutoMigrate(&models.Coupon{}); err != nil{
		log.Println("Failed to migrate coupon model:", err)
	}else{
		log.Println("coupon model migration was successfull")
	}

	if err := DB.AutoMigrate(&models.CouponUsage{}); err != nil{
		log.Println("Failed to migrate coupon usage model:", err)
	}else{
		log.Println("coupon usage model migration was successfull")
	}

	if err := DB.AutoMigrate(&models.Wallet{}); err != nil{
		log.Println("Failed to migrate wallet model:", err)
	}else{
		log.Println("coupon wallet migration was successfull")
	}

	if err := DB.AutoMigrate(&models.WalletHistory{}); err != nil{
		log.Println("Failed to migrate wallet History model:", err)
	}else{
		log.Println("coupon wallet History migration was successfull")
	}

	return nil
}