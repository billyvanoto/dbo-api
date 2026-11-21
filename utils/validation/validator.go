package validation

import (
	"dbo-api/config"
	"dbo-api/models"
	"log"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var phoneRegex = regexp.MustCompile(`^(\+62|62|0)8[1-9][0-9]{6,10}$`)

func RegisterCustomValidations() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		log.Fatal("Failed to assert Gin validator engine")
	}
	if err := v.RegisterValidation("phone_id", validatePhoneID); err != nil {
		log.Fatal("Failed to register phone_id:", err)
	}
	if err := v.RegisterValidation("unique_customer_email", validateUniqueCustomerEmail); err != nil {
		log.Fatal("Failed to register unique_customer_email:", err)
	}
}

func validatePhoneID(fl validator.FieldLevel) bool {
	return phoneRegex.MatchString(fl.Field().String())
}

func validateUniqueCustomerEmail(fl validator.FieldLevel) bool {
	if config.DB == nil {
		log.Println("WARNING: config.DB is nil inside validateUniqueEmail")
		return false
	}

	email := strings.TrimSpace(strings.ToLower(fl.Field().String()))
	if email == "" {
		return true
	}

	var count int64
	query := config.DB.Model(&models.Customer{}).
		Where("LOWER(email) = ? AND deleted_at IS NULL AND deleted_by IS NULL", email)

	parent := fl.Parent()
	if idField := parent.FieldByName("CustomerID"); idField.IsValid() && idField.Uint() > 0 {
		query = query.Where("customer_id <> ?", idField.Uint())
	}

	if err := query.Count(&count).Error; err != nil {
		log.Println("Error checking unique email:", err)
		return false
	}

	return count == 0
}
