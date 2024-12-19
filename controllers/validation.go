package controllers

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (input *UserInput) UserValidate() error {
	if input.UserName == "" {
		return errors.New("username is required")
	}

	if input.Email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(input.Email) {
		return errors.New("invalid email format")
	}

	if input.Password == "" {
		return errors.New("password required")
	}

	if len(input.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	if input.PhoneNo != "" && len(input.PhoneNo) != 10 {
		return errors.New("phone number must be 10 digits")
	}

	return nil
}

func (input *AdminInput) AdminValidate() error {
	if input.Name == "" {
		return errors.New("name is required")
	}

	if input.Email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(input.Email) {
		return errors.New("invalid email format")
	}

	if input.Password == "" {
		return errors.New("invalid email format")
	}

	if len(input.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	return nil
}

func (input *Product) ProductValidate() error {

	if input.ProductName == "" {
		return errors.New("Product name is required")
	}

	if input.Description == "" {
		return errors.New("Product description is required")
	}

	if input.Price < 0 {
		return errors.New("Product price must be greater than 0")
	}

	if input.Quantity <= -1 {
		return errors.New("Product quantity must be greater than 0")
	}

	if input.ImgURLs == "" {
		return errors.New("Product imgurl is required")
	}

	if input.OfferType != "" {
		if input.OfferType != "percentage" && input.OfferType != "fixed" {
			return errors.New("discount type must be either 'percentage' or 'fixed'")
		}
	}
	if input.OfferValue < 0 {
		return errors.New("discount value must be greater than 0")
	}

	if input.OfferType == "percentage" && input.OfferValue > 100 {
		return errors.New("percentage discount cannot exceed 100%")
	}

	if input.OfferType != "" && input.OfferValue <= 0 {
		return errors.New("you have add the both offer type and offer value")
	}

	if input.OfferType == "" && input.OfferValue > 0 {
		return errors.New("you have add the both offer type and offer value")
	}

	return nil
}

func (input *Product) ProductValidateForPatch() error {

	if input.ProductName != "" && len(input.ProductName) == 0 {
		return errors.New("Product name is required")
	}

	if input.Description != "" && len(input.Description) < 10 {
		return errors.New("Product description is atleast 4 words")
	}

	if input.Price < 0 {
		return errors.New("Product price must be greater than or equal to 0")
	}

	if input.Quantity < 0 {
		return errors.New("Product quantity must be zero or greater")
	}

	if input.ImgURLs != "" && len(input.ImgURLs) == 0 {
		return errors.New("Product imgurl is required")
	}

	if input.OfferType != "" || input.OfferValue != 0 { 
		if input.OfferType != "" { 
			if input.OfferType != "percentage" && input.OfferType != "fixed" {
				return errors.New("discount type must be either 'percentage' or 'fixed'")
			}
		}
	
		if input.OfferValue < 0 { 
			return errors.New("discount value must be greater than 0")
		}
	
		if input.OfferType == "percentage" && input.OfferValue > 100 {
			return errors.New("percentage discount cannot exceed 100%")
		}
	
		if (input.OfferType != "" && input.OfferValue <= 0) || (input.OfferType == "" && input.OfferValue > 0) {
			return errors.New("you must provide both offer type and offer value together")
		}
	}
	

	return nil
}

func (a *NewAddress) AddressValidate() error {

	validString := regexp.MustCompile(`^[a-zA-Z0-9\s,.-]*$`)

	if a.Name == "" {
		return errors.New("name is required")
	}
	if !validString.MatchString(a.Name) {
		return errors.New("invalid name: must not contain special characters like *")
	}
	if len(a.PhoneNo) != 10 || a.PhoneNo == "" {
		return errors.New("phone must contain 10 digits")
	}
	if !validString.MatchString(a.PhoneNo) {
		return errors.New("phone number: must not contain special characters like *")
	}
	if a.AddressLine1 == "" {
		return errors.New("addressline 1 is required")
	}
	if !validString.MatchString(a.AddressLine1) {
		return errors.New("invalid address line 1: must not contain special characters like *")
	}
	if a.City == "" {
		return errors.New("city is required")
	}
	if !validString.MatchString(a.City) {
		return errors.New("invalid city: must not contain special characters like *")
	}
	if a.State == "" {
		return errors.New("state is required")
	}
	if !validString.MatchString(a.State) {
		return errors.New("invalid state: must not contain special characters like *")
	}
	if a.PostalCode == "" {
		return errors.New("postal code is required")
	}
	if !validString.MatchString(a.PostalCode) {
		return errors.New("invalid postal code: must not contain special characters like *")
	}
	if a.Country == "" {
		return errors.New("country is required")
	}
	if !validString.MatchString(a.Country) {
		return errors.New("invalid country: must not contain special characters like *")
	}
	if a.AddressType == "" || (a.AddressType != "shipping" && a.AddressType != "billing") {
		return errors.New("invalid address type: must be 'shipping' or 'billing'")
	}

	if !validString.MatchString(a.AddressLine2) {
		return errors.New("invalid address line 2: must not contain special characters like *")
	}
	if a.LandMark != "" && !validString.MatchString(a.LandMark) {
		return errors.New("invalid landmark: must not contain special characters like *")
	}

	if a.AddressLine1 != "" && a.AddressLine1 == a.AddressLine2 {
		return errors.New("address line 1 and address line 2 should not be identical")
	}

	return nil
}

func (a *NewAddress) AddressValidateForPatch() error {
	// Regex for validating strings
	validString := regexp.MustCompile(`^[a-zA-Z0-9\s,.-]*$`)

	// Validate fields only if they are not empty (to support partial updates)
	if a.Name != "" && !validString.MatchString(a.Name) {
		return errors.New("invalid name: must not contain special characters like *")
	}

	if a.PhoneNo != "" {
		if len(a.PhoneNo) != 10 {
			return errors.New("phone must contain exactly 10 digits")
		}
		if !validString.MatchString(a.PhoneNo) {
			return errors.New("invalid phone number: must not contain special characters like *")
		}
	}

	if a.AddressLine1 != "" && !validString.MatchString(a.AddressLine1) {
		return errors.New("invalid address line 1: must not contain special characters like *")
	}

	if a.City != "" && !validString.MatchString(a.City) {
		return errors.New("invalid city: must not contain special characters like *")
	}

	if a.State != "" && !validString.MatchString(a.State) {
		return errors.New("invalid state: must not contain special characters like *")
	}

	if a.PostalCode != "" && !validString.MatchString(a.PostalCode) {
		return errors.New("invalid postal code: must not contain special characters like *")
	}

	if a.Country != "" && !validString.MatchString(a.Country) {
		return errors.New("invalid country: must not contain special characters like *")
	}

	if a.AddressType != "" && (a.AddressType != "shipping" && a.AddressType != "billing") {
		return errors.New("invalid address type: must be 'shipping' or 'billing'")
	}

	// Optional fields
	if a.AddressLine2 != "" && !validString.MatchString(a.AddressLine2) {
		return errors.New("invalid address line 2: must not contain special characters like *")
	}

	if a.LandMark != "" && !validString.MatchString(a.LandMark) {
		return errors.New("invalid landmark: must not contain special characters like *")
	}

	// Ensure address line 1 and address line 2 are not identical if both are provided
	if a.AddressLine1 != "" && a.AddressLine1 == a.AddressLine2 {
		return errors.New("address line 1 and address line 2 should not be identical")
	}

	return nil
}

func (input *UserInput) UserValidateForPatch() error {
	if input.UserName != "" && len(input.UserName) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	if input.Email != "" {
		return errors.New("you can't change your email")
	}

	if input.Password != "" {
		return errors.New("you can't change you password")
	}

	if input.PhoneNo != "" {
		if len(input.PhoneNo) != 10 {
			return errors.New("phone number must be 10 digits")
		}
	}

	return nil
}

func convertToUint(userID interface{}) (uint, error) {
	if userID == nil {
		return 0, fmt.Errorf("user_id is nil")
	}

	switch v := userID.(type) {
	case uint:
		return v, nil
	case int:
		if v < 0 {
			return 0, fmt.Errorf("id cannot be negative")
		}
		return uint(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("id cannot be negative")
		}
		return uint(v), nil
	case string:
		parsedID, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("failed to parse id string to uint: %v", err)
		}
		return uint(parsedID), nil
	default:
		return 0, fmt.Errorf("unsupported type for id: %T", userID)
	}
}

func (input *Coupon) ValidateCoupon() error {
	if input.CouponName == "" {
		return errors.New("coupon name is required")
	}

	if input.CouponName != strings.ToUpper(input.CouponName) {
		return errors.New("coupon name must be in uppercase")
	}

	if input.CouponCode == "" {
		return errors.New("coupon code is required")
	}

	if input.DiscountType != "percentage" && input.DiscountType != "fixed" {
		return errors.New("discount type must be either 'percentage' or 'fixed'")
	}

	if input.DiscountValue <= 0 {
		return errors.New("discount value must be greater than 0")
	}

	if input.DiscountType == "percentage" && input.MaximumDiscount > 0 && input.DiscountValue > 100 {
		return errors.New("percentage discount cannot exceed 100%")
	}
	if input.DiscountType == "percentage" && input.MaximumDiscount > 0 && input.MaximumDiscount < input.DiscountValue {
		return errors.New("maximum discount cannot be less than the discount value")
	}

	if input.MinimumAmount < 0 {
		return errors.New("minimum purchase amount must be greater than or equal to 0")
	}

	if input.UsageLimit <= 0 {
		return errors.New("total usage limit must be greater than 0")
	}

	if input.StartDate.IsZero() {
		return errors.New("start date is required")
	}
	if input.ExpiryTime.IsZero() {
		return errors.New("expiry time is required")
	}
	if input.ExpiryTime.Before(input.StartDate) {
		return errors.New("expiry time must be after the start date")
	}
	if input.ExpiryTime.Before(time.Now()) {
		return errors.New("expiry time cannot be in the past")
	}
	return nil
}

func (input *Coupon) ValidateCouponForPatch() error {
	if input.CouponName != "" {
		if input.CouponName != strings.ToUpper(input.CouponName) {
			return errors.New("coupon name must be in uppercase")
		}
	}

	if input.CouponCode != "" {
		if len(input.CouponCode) == 0 {
			return errors.New("coupon code is required")
		}
	}

	if input.DiscountType != "" {
		if input.DiscountType != "percentage" && input.DiscountType != "fixed" {
			return errors.New("discount type must be either 'percentage' or 'fixed'")
		}
	}

	if input.DiscountValue != 0 {
		if input.DiscountValue <= 0 {
			return errors.New("discount value must be greater than 0")
		}
		if input.DiscountType == "percentage" && input.DiscountValue > 100 {
			return errors.New("percentage discount cannot exceed 100%")
		}
	}

	if input.MaximumDiscount != 0 {
		if input.DiscountType == "percentage" && input.MaximumDiscount < input.DiscountValue {
			return errors.New("maximum discount cannot be less than the discount value")
		}
	}

	if input.MinimumAmount != 0 {
		if input.MinimumAmount < 0 {
			return errors.New("minimum purchase amount must be greater than or equal to 0")
		}
	}

	if input.UsageLimit != 0 {
		if input.UsageLimit <= 0 {
			return errors.New("total usage limit must be greater than 0")
		}
	}

	if !input.StartDate.IsZero() {
		if input.ExpiryTime.Before(input.StartDate) {
			return errors.New("expiry time must be after the start date")
		}
	}

	if !input.ExpiryTime.IsZero() {
		if input.ExpiryTime.Before(time.Now()) {
			return errors.New("expiry time cannot be in the past")
		}
	}

	return nil
}
