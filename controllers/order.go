package controllers

import (
	"dbo-api/config"
	"dbo-api/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func activeOrders(db *gorm.DB) *gorm.DB {
	return db.Where("orders.deleted_at IS NULL")
}

// GetOrders with Paginate & Search (?page=1&limit=10&search=ORD-001&status=PAID)
func GetOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	offset := (page - 1) * limit

	var orders []models.Order
	var total int64

	query := config.DB.Model(&models.Order{}).
		Scopes(activeOrders).
		Preload("Customer").
		Preload("Items")

	if search != "" {
		query = query.Where("order_no ILIKE ?", "%"+search+"%")
	}
	if status != "" {
		query = query.Where("status = ?", strings.ToUpper(status))
	}

	query.Count(&total)
	query.Limit(limit).Offset(offset).Order("order_id DESC").Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
		"meta": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func GetOrderDetail(c *gin.Context) {
	var order models.Order
	err := config.DB.Scopes(activeOrders).
		Preload("Customer").
		Preload("Items").
		Where("order_id = ?", c.Param("id")).
		First(&order).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

func CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserEmail := c.GetString("email")
	if currentUserEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Verify active customer exists
	var customerCount int64
	config.DB.Model(&models.Customer{}).
		Where("customer_id = ? AND deleted_at IS NULL", req.CustomerID).
		Count(&customerCount)
	if customerCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Active customer not found"})
		return
	}

	// Build items & calculate totals
	var order models.Order
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Atomically pull next sequence number from PostgreSQL
		var seq int64
		if err := tx.Raw("SELECT nextval('order_no_seq')").Scan(&seq).Error; err != nil {
			return err
		}
		orderNo := fmt.Sprintf("ORD-%s-%05d", time.Now().Format("20060102"), seq)

		// 2. Calculate line item totals and grand total
		var grandTotal float64
		items := make([]models.OrderItem, 0, len(req.Items))
		for _, itemReq := range req.Items {
			itemTotal := float64(itemReq.Qty) * itemReq.Price
			grandTotal += itemTotal
			items = append(items, models.OrderItem{
				Name:      itemReq.Name,
				Qty:       itemReq.Qty,
				Price:     itemReq.Price,
				Total:     itemTotal,
				CreatedBy: currentUserEmail,
				UpdatedBy: currentUserEmail,
			})
		}

		status := req.Status
		if status == "" {
			status = "PENDING"
		}

		order = models.Order{
			CustomerID: req.CustomerID,
			OrderNo:    orderNo,
			Total:      grandTotal,
			Status:     status,
			Items:      items,
			CreatedBy:  currentUserEmail,
			UpdatedBy:  currentUserEmail,
		}

		return tx.Create(&order).Error
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order "})
		return
	}

	// Reload with Customer relation for complete response
	config.DB.Preload("Customer").Preload("Items").First(&order, order.OrderID)
	c.JSON(http.StatusCreated, gin.H{"data": order})
}

func UpdateOrder(c *gin.Context) {
	var order models.Order
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	if err := config.DB.Scopes(activeOrders).Where("orders.order_id = ?", orderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	var req models.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserEmail := c.GetString("email")
	if currentUserEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Verify active customer exists
	var customerCount int64
	config.DB.Model(&models.Customer{}).
		Where("customer_id = ? AND deleted_at IS NULL", req.CustomerID).
		Count(&customerCount)
	if customerCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Active customer not found"})
		return
	}

	err = config.DB.Transaction(func(tx *gorm.DB) error {
		// Replace existing order items cleanly inside the transaction
		if err := tx.Where("order_id = ?", order.OrderID).Delete(&models.OrderItem{}).Error; err != nil {
			return err
		}

		var grandTotal float64
		newItems := make([]models.OrderItem, 0, len(req.Items))
		for _, itemReq := range req.Items {
			itemTotal := float64(itemReq.Qty) * itemReq.Price
			grandTotal += itemTotal
			newItems = append(newItems, models.OrderItem{
				OrderID:   order.OrderID,
				Name:      itemReq.Name,
				Qty:       itemReq.Qty,
				Price:     itemReq.Price,
				Total:     itemTotal,
				CreatedBy: currentUserEmail,
				UpdatedBy: currentUserEmail,
			})
		}

		if err := tx.Create(&newItems).Error; err != nil {
			return err
		}

		order.CustomerID = req.CustomerID
		order.Status = req.Status
		order.Total = grandTotal
		order.UpdatedBy = currentUserEmail

		return tx.Save(&order).Error
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to update order"})
		return
	}

	config.DB.Preload("Customer").Preload("Items").First(&order, order.OrderID)
	c.JSON(http.StatusOK, gin.H{"data": order})
}

// Soft Delete Order (sets deleted_at and deleted_by)
func DeleteOrder(c *gin.Context) {
	currentUserEmail := c.GetString("email")
	if currentUserEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	now := time.Now()
	result := config.DB.Model(&models.Order{}).
		Scopes(activeOrders).
		Where("order_id = ?", c.Param("id")).
		Updates(map[string]interface{}{
			"deleted_at": now,
			"deleted_by": currentUserEmail,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}
