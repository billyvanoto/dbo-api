package controllers

import (
	"dbo-api/config"
	"dbo-api/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Scope to exclude soft-deleted records
func activeCustomers(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL AND deleted_by IS NULL")
}

// Get with Paginate & Search (?page=1&limit=10&search=budi)
func GetCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	offset := (page - 1) * limit

	var customers []models.Customer
	var total int64

	query := config.DB.Model(&models.Customer{}).Scopes(activeCustomers)
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?", like, like, like)
	}

	query.Count(&total)
	query.Limit(limit).Offset(offset).Order("customer_id DESC").Find(&customers)

	c.JSON(http.StatusOK, gin.H{
		"data": customers,
		"meta": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func GetCustomerDetail(c *gin.Context) {
	var customer models.Customer
	if err := config.DB.Scopes(activeCustomers).Preload("Orders").First(&customer, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": customer})
}

func CreateCustomer(c *gin.Context) {
	var req models.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserEmail := c.GetString("email")
	if currentUserEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	customer := models.Customer{
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Address:   req.Address,
		CreatedBy: currentUserEmail,
		UpdatedBy: currentUserEmail,
	}

	if err := config.DB.Create(&customer).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": customer})
}

func UpdateCustomer(c *gin.Context) {
	var customer models.Customer
	if err := config.DB.Scopes(activeCustomers).First(&customer, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	var req models.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the email belongs to a DIFFERENT active customer
	normalizedEmail := strings.TrimSpace(strings.ToLower(req.Email))
	var existingCount int64
	config.DB.Model(&models.Customer{}).
		Scopes(activeCustomers).
		Where("LOWER(email) = ? AND customer_id <> ?", normalizedEmail, customer.CustomerID).
		Count(&existingCount)

	if existingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already registered"})
		return
	}

	currentUserEmail := c.GetString("email")
	if currentUserEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	customer.Name = req.Name
	customer.Email = req.Email
	customer.Phone = req.Phone
	customer.Address = req.Address
	customer.UpdatedBy = currentUserEmail

	if err := config.DB.Save(&customer).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": customer})
}

// Soft Delete Customer
func DeleteCustomer(c *gin.Context) {
	currentUserEmail := c.GetString("email")
	if currentUserEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	now := time.Now()
	result := config.DB.Model(&models.Customer{}).
		Scopes(activeCustomers).
		Where("customer_id = ?", c.Param("id")).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"deleted_by": currentUserEmail,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

// GetCustomerOptions returns lightweight active customers for dropdowns (?search=budi&limit=20)
func GetCustomerOptions(c *gin.Context) {
	search := strings.TrimSpace(c.Query("search"))
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20 // Hard cap to prevent unbounded queries
	}

	var options []models.CustomerOptionResponse
	query := config.DB.Model(&models.Customer{}).
		Scopes(activeCustomers).
		Select("customer_id, name, email")

	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}

	if err := query.Order("name ASC").Limit(limit).Scan(&options).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer options"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": options})
}
