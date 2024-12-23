package controllers

import (
	"bytes"
	"fmt"
	"kars/database"
	"kars/models"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jung-kurt/gofpdf/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"github.com/wcharczuk/go-chart/v2"
	"gorm.io/gorm"
)

type OrderCount struct {
	TotalOrder     uint `json:"total_order"`
	TotalPending   uint `json:"total_pending"`
	TotalPlaced    uint `json:"total_placed"`
	TotalShipped   uint `json:"total_shipped"`
	TotalDelivered uint `json:"total_delivered"`
	TotalCancelled uint `json:"total_cancelled"`
	TotalReturned  uint `json:"total_returned"`
}

type AmountInformation struct {
	TotalAmountBeforeDeduction float64 `json:"total_amount_before_deduction" gorm:"column:total_amount_before_deduction"`
	TotalCouponDeduction       float64 `json:"total_coupon_deduction" gorm:"column:total_coupon_deduction"`
	TotalDeliveryCharges       float64 `json:"total_delivery_charge" gorm:"column:total_delivery_charge"`
	TotalAmountAfterDeduction  float64 `json:"total_amount_after_deduction" gorm:"column:total_amount_after_deduction"`
}

func getTotalSales(startDate, endDate time.Time) (float64, error) {
	var totalSales float64

	fmt.Printf("Fetching sales data between %v and %v\n", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	err := database.DB.Model(&models.Order{}).
		Where("created_at BETWEEN ? AND ?", startDate.UTC(), endDate.UTC()).
		Select("COALESCE(SUM(final_price), 0)").
		Scan(&totalSales).Error

	if err != nil {
		fmt.Printf("Error executing query: %v\n", err)
		return 0, err
	}

	fmt.Printf("Total sales fetched: %f\n", totalSales)
	return totalSales, nil
}

func GetSalesReport(c *fiber.Ctx) error {

	reportType := c.Query("report_type")
	if reportType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "report type is required",
		})
	}

	var startDate, endDate time.Time
	var err error
	currentTime := time.Now()

	switch reportType {
	case "daily":
		startDate = time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location())
		endDate = startDate.Add(24 * time.Hour)

	case "weekly":
		weekdayOffset := int(currentTime.Weekday() - time.Monday)
		if weekdayOffset < 0 {
			weekdayOffset += 7
		}
		startDate = currentTime.AddDate(0, 0, -weekdayOffset)
		startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, currentTime.Location())
		endDate = startDate.Add(7 * 24 * time.Hour)

	case "yearly":
		startDate = time.Date(currentTime.Year(), time.January, 1, 0, 0, 0, 0, currentTime.Location())
		endDate = time.Date(currentTime.Year()+1, time.January, 1, 0, 0, 0, 0, currentTime.Location())

	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report type. Use 'daily', 'weekly', 'yearly', or 'custom'.",
		})
	}

	if startDate.After(endDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date range. start_date must be before end_date.",
		})
	}

	totalSales, err := getTotalSales(startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to calculate %s sales", reportType),
		})
	}

	var orderCount OrderCount
	err = database.DB.Model(&models.Order{}).
		Select(`
			COUNT(*) AS total_order,
			COUNT(CASE WHEN order_status = 'pending' THEN 1 END) AS total_pending,
			COUNT(CASE WHEN order_status = 'placed' THEN 1 END) AS total_placed,
			COUNT(CASE WHEN order_status = 'shipped' THEN 1 END) AS total_shipped,
			COUNT(CASE WHEN order_status = 'delivered' THEN 1 END) AS total_delivered,
			COUNT(CASE WHEN order_status = 'cancelled' THEN 1 END) AS total_cancelled,
			COUNT(CASE WHEN order_status = 'returned' THEN 1 END) AS total_returned
		`).
		Where("created_at BETWEEN ? AND ?", startDate.UTC(), endDate.UTC()).
		Scan(&orderCount).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch order counts",
		})
	}

	var amountInfo AmountInformation
	err = database.DB.Model(&models.Order{}).
		Select(`
			COALESCE(SUM(total_price), 0) AS total_amount_before_deduction,
			COALESCE(SUM(discount_amount), 0) AS total_coupon_deduction,
			COALESCE(SUM(shipping_amount), 0) AS total_delivery_charge,
			COALESCE(SUM(final_price), 0) AS total_amount_after_deduction
		`).
		Where("created_at BETWEEN ? AND ?", startDate.UTC(), endDate.UTC()).
		Scan(&amountInfo).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch amount information",
		})
	}

	formattedStartDate := startDate.Format("2006-01-02")
	formattedEndDate := endDate.Format("2006-01-02")

	chartPath := generateChartReport(orderCount)
	barChartPath := generateBarChartReport(amountInfo)

	pdf, err := generatePDFReport(orderCount, amountInfo, totalSales, formattedStartDate, formattedEndDate, chartPath, barChartPath)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "failed to generate pdf",
		})
	}
	filename := fmt.Sprintf("sales_report_%s.pdf", time.Now().Format("20060102150405"))

	if err := pdf.OutputFileAndClose(filename); err != nil {
		log.Println(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate PDF"})
	}

	return c.SendFile(filename)
}

func generatePDFReport(orderCount OrderCount, amountInfo AmountInformation, totalSales float64, startDate, endDate string, chartPath string, barChartPath string) (gofpdf.Pdf, error) {
	pdf := gofpdf.New("P", "mm", "", "")
	pdf.AddPageFormat("P", gofpdf.SizeType{Wd: 210, Ht: 450}) // Standard A4 size

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(0, 10, "Sales Report", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Report Dates
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(40, 10, "Start Date:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 10, startDate, "", 1, "L", false, 0, "")

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(40, 10, "End Date:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 10, endDate, "", 1, "L", false, 0, "")
	// pdf.Ln(5)

	// Total Sales
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(40, 10, "Total Sales:", "", 0, "L", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 10, fmt.Sprintf("%.2f", totalSales), "", 1, "L", false, 0, "")
	pdf.Ln(10)

	// Order Count Section
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "Order Counts", "B", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Table Header
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(90, 10, "Order Status", "1", 0, "C", false, 0, "")
	pdf.CellFormat(90, 10, "Count", "1", 1, "C", false, 0, "")

	// Table Rows
	pdf.SetFont("Arial", "", 12)
	orderStatuses := []struct {
		Status string
		Count  uint
	}{
		{"Total Orders", orderCount.TotalOrder},
		{"Pending Orders", orderCount.TotalPending},
		{"Placed Orders", orderCount.TotalPlaced},
		{"Shipped Orders", orderCount.TotalShipped},
		{"Delivered Orders", orderCount.TotalDelivered},
		{"Cancelled Orders", orderCount.TotalCancelled},
		{"Returned Orders", orderCount.TotalReturned},
	}

	for _, order := range orderStatuses {
		pdf.CellFormat(90, 10, order.Status, "1", 0, "L", false, 0, "")
		pdf.CellFormat(90, 10, fmt.Sprintf("%d", order.Count), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(10)

	// Amount Details Section
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "Amount Information", "B", 1, "L", false, 0, "")
	pdf.Ln(5)

	// Table Header
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(100, 10, "Description", "1", 0, "C", false, 0, "")
	pdf.CellFormat(80, 10, "Amount", "1", 1, "C", false, 0, "")

	// Table Rows
	amountDetails := []struct {
		Description string
		Amount      float64
	}{
		{"Total Amount Before Deduction", amountInfo.TotalAmountBeforeDeduction},
		{"Total Coupon Deduction", amountInfo.TotalCouponDeduction},
		{"Total Delivery Charges", amountInfo.TotalDeliveryCharges},
		{"Total Amount After Deduction", amountInfo.TotalAmountAfterDeduction},
	}

	for _, amount := range amountDetails {
		pdf.CellFormat(100, 10, amount.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(80, 10, fmt.Sprintf("%.2f", amount.Amount), "1", 1, "R", false, 0, "")
	}
	pdf.Ln(10)

	// Add Chart Images
	if chartPath != "" {
		pdf.Image(chartPath, 55, pdf.GetY(), 100, 0, false, "", 0, "")
		pdf.Ln(110)
	} else {
		pdf.CellFormat(0, 10, "Chart not available", "", 1, "C", false, 0, "")
	}

	if barChartPath != "" {
		if _, err := os.Stat(barChartPath); err == nil {
			pdf.Image(barChartPath, 55, pdf.GetY(), 100, 0, false, "", 0, "")
			pdf.Ln(110)
		} else {
			pdf.CellFormat(0, 10, "Bar chart file not found", "", 1, "C", false, 0, "")
		}
	} else {
		pdf.CellFormat(0, 10, "Bar chart not available", "", 1, "C", false, 0, "")
	}

	return pdf, nil
}

func generateChartReport(orderCount OrderCount) string {

	pieValues := []chart.Value{
		{Value: float64(orderCount.TotalPending), Label: "Pending"},
		{Value: float64(orderCount.TotalPlaced), Label: "Placed"},
		{Value: float64(orderCount.TotalShipped), Label: "Shipped"},
		{Value: float64(orderCount.TotalDelivered), Label: "Delivered"},
		{Value: float64(orderCount.TotalCancelled), Label: "Cancelled"},
		{Value: float64(orderCount.TotalReturned), Label: "Returned"},
	}

	total := orderCount.TotalOrder

	filteredPieValues := []chart.Value{}
	for _, v := range pieValues {
		if v.Value > 0 {
			filteredPieValues = append(filteredPieValues, chart.Value{
				Label: v.Label,
				Value: (v.Value / float64(total)) * 100,
			})
		}
	}

	if len(filteredPieValues) == 0 {
		fmt.Println("All values are zero, no chart will be rendered")
		return ""
	}

	pie := chart.PieChart{
		Width:  600,
		Height: 600,
		Values: filteredPieValues,
	}

	fileName := "order_chart.png"
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error creating chart file:", err)
		return ""
	}
	defer f.Close()

	err = pie.Render(chart.PNG, f)
	if err != nil {
		fmt.Println("Error rendering chart:", err)
		return ""
	}

	return fileName
}

func generateBarChartReport(amountInfo AmountInformation) string {
	fmt.Println("Debug Values:")
	fmt.Printf("Total Before Deduction: %.2f\n", amountInfo.TotalAmountBeforeDeduction)
	fmt.Printf("Coupon Deduction: %.2f\n", amountInfo.TotalCouponDeduction)
	fmt.Printf("Delivery Charges: %.2f\n", amountInfo.TotalDeliveryCharges)
	fmt.Printf("After Deduction: %.2f\n", amountInfo.TotalAmountAfterDeduction)

	values := []chart.Value{
		{Label: fmt.Sprintf("Total Before\nDeduction\n₹%.2f", amountInfo.TotalAmountBeforeDeduction), Value: amountInfo.TotalAmountBeforeDeduction, Style: chart.Style{FillColor: drawing.ColorFromHex("2E86C1")}},
		{Label: fmt.Sprintf("Coupon\nDeduction\n₹%.2f", amountInfo.TotalCouponDeduction), Value: amountInfo.TotalCouponDeduction, Style: chart.Style{FillColor: drawing.ColorFromHex("E74C3C")}},
		{Label: fmt.Sprintf("Delivery\nCharges\n₹%.2f", amountInfo.TotalDeliveryCharges), Value: amountInfo.TotalDeliveryCharges, Style: chart.Style{FillColor: drawing.ColorFromHex("27AE60")}},
		{Label: fmt.Sprintf("After\nDeduction\n₹%.2f", amountInfo.TotalAmountAfterDeduction), Value: amountInfo.TotalAmountAfterDeduction, Style: chart.Style{FillColor: drawing.ColorFromHex("F39C12")}},
	}

	barChart := chart.BarChart{
		Title: "Amount Breakdown",
		Background: chart.Style{
			Padding: chart.Box{
				Top:    50,
				Left:   50,
				Right:  50,
				Bottom: 50,
			},
			FillColor: drawing.ColorWhite,
		},
		TitleStyle: chart.Style{
			FontSize:  44,
			FontColor: drawing.ColorBlack,
		},
		Width:    2400, // Increased width for better label display
		Height:   1800, // Increased height for better spacing
		BarWidth: 250,  // Increased bar width for better clarity
		Bars:     values,
		YAxis: chart.YAxis{
			Style: chart.Style{
				FontSize:  44,
				FontColor: drawing.ColorBlack,
			},
			ValueFormatter: func(v interface{}) string {
				return fmt.Sprintf("₹%.2f", v.(float64))
			},
			Ticks: []chart.Tick{
				{Value: 1000, Label: "₹1,000"},
				{Value: 3000, Label: "₹3,000"},
				{Value: 6000, Label: "₹6,000"},
				{Value: 9000, Label: "₹9,000"},
				{Value: 12000, Label: "₹12,000"},
				{Value: 15000, Label: "₹15,000"},
				{Value: 18000, Label: "₹18,000"},
				{Value: 21000, Label: "₹21,000"},
				{Value: 24000, Label: "₹24,000"},
				{Value: 27000, Label: "₹27,000"},
				{Value: 30000, Label: "₹30,000"},
				{Value: 33000, Label: "₹33,000"},
				{Value: 36000, Label: "₹36,000"},
				{Value: 39000, Label: "₹39,000"},
			},
			Range: &chart.ContinuousRange{
				Min: 0,
				Max: amountInfo.TotalAmountBeforeDeduction * 1.1,
			},
		},
		XAxis: chart.Style{
			FontSize:  36, // Increased font size for better readability
			FontColor: drawing.ColorBlack,
		},
		Canvas: chart.Style{
			FillColor: drawing.ColorWhite,
		},
	}

	outputPath := "bar_chart.png"

	// Remove the existing file
	os.Remove(outputPath)
	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return ""
	}
	defer f.Close()

	// Render chart
	err = barChart.Render(chart.PNG, f)
	if err != nil {
		fmt.Println("Error rendering chart:", err)
		return ""
	}

	return outputPath
}


func InvoiceDownload(c *fiber.Ctx) error {
	orderID := c.Query("order_id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "order id is required",
		})
	}

	var order models.Order
	if err := database.DB.Preload("OrderItems").First(&order, "id = ?", orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "order not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve order",
		})
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 20) // Large font for heading
	pdf.CellFormat(190, 12, "ORDER BILL", "0", 1, "C", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12) // Bold font for section titles
	pdf.Cell(40, 10, "From:")
	pdf.SetFont("Arial", "", 12) // Regular font for details
	pdf.MultiCell(150, 7, "Kars\nNear Thrissur Round\nThrissur, kerala, 680702\nPhone: +91 8921236125", "0", "L", false)
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "To:")
	pdf.SetFont("Arial", "", 12)
	pdf.MultiCell(150, 7, fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		order.OrderAddress.Name,         // Customer Name
		order.OrderAddress.AddressLine1, // Street Address
		order.OrderAddress.City,         // City
		order.OrderAddress.PostalCode,
		order.OrderAddress.PhoneNo), "0", "L", false)
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Order ID:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(80, 10, orderID)
	pdf.Ln(8)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Order Date:")
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(80, 10, order.CreatedAt.Format("2006-01-02 15:04:05"))
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)

	pdf.CellFormat(20, 10, "Item ID", "1", 0, "C", false, 0, "")
	pdf.CellFormat(50, 10, "Item Name", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 10, "Quantity", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Price", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 10, "Total", "1", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 12)
	for _, item := range order.OrderItems {
		if item.IsCancelled == "cancelled" {
			continue
		}

		pdf.CellFormat(20, 10, fmt.Sprintf("%d", item.ProductID), "1", 0, "C", false, 0, "")
		pdf.CellFormat(50, 10, item.ProductName, "1", 0, "L", false, 0, "")
		pdf.CellFormat(20, 10, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("$%.2f", item.ProductPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(30, 10, fmt.Sprintf("$%.2f", float64(item.Quantity)*item.ProductPrice), "1", 1, "R", false, 0, "")
	}
	pdf.SetFont("Arial", "B", 12)
	pdf.Ln(8)
	pdf.Cell(140, 10, "Total Amount:")
	pdf.Cell(40, 10, fmt.Sprintf("$%.2f", order.TotalPrice))

	pdf.SetFont("Arial", "B", 12)
	pdf.Ln(8)
	pdf.Cell(140, 10, "Delivery Charge:")
	pdf.Cell(40, 10, fmt.Sprintf("$%.2f", order.ShippingAmount))

	if order.DiscountAmount != 0 {
		pdf.SetFont("Arial", "B", 12)
		pdf.Ln(8)
		pdf.Cell(140, 10, "Discount Amount:")
		pdf.Cell(40, 10, fmt.Sprintf("$%.2f", order.DiscountAmount))
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Ln(8)
	pdf.Cell(140, 10, "Final Amount:")
	pdf.Cell(40, 10, fmt.Sprintf("$%.2f", order.FinalPrice))

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to generate PDF",
		})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=invoice.pdf")

	return c.Send(buf.Bytes())
}
