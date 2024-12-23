package routes

import (
	"kars/controllers"
	"kars/middleware"

	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App) {
	//User Side Routes
	app.Post("/api/user/signup", controllers.UserSignUp)
	app.Post("/api/user/verify_otp/:user_email", controllers.VerifyOtpAndCreateUser)
	app.Post("/api/user/otp_resend/:user_email", controllers.ResendOTP)
	app.Get("/api/user/profile", middleware.CheckUserStatus, controllers.UserProfile)
	app.Patch("/api/user/profile", middleware.CheckUserStatus, controllers.EditProfile)
	app.Post("/api/user/password", middleware.CheckUserStatus, controllers.ForgotPasswordStep1)
	app.Post("/api/user/password/verify", middleware.CheckUserStatus, controllers.ForgotPasswordStep2)
	app.Patch("/api/user/password", middleware.CheckUserStatus, controllers.ForgotPasswordStep3)
	app.Get("/api/user/google_signin", controllers.InitGoogleSignIn)
	app.Get("/api/user/google_signup/callback", controllers.GoogleSignUpCallback)
	app.Post("/api/user/login", controllers.UserLogin)
	app.Post("/api/user/logout", controllers.UserLogout)
	app.Get("/api/user/products", middleware.CheckUserStatus, controllers.UserProductList)

	//Admin Side Routes
	app.Post("/api/admin/signup", controllers.AdminSignUp)
	app.Post("/api/admin/login", controllers.AdminLogin)
	app.Get("/api/admin/userslist", controllers.UserList)
	app.Post("/api/admin/user/block/:user_id", middleware.AdminMiddleware, controllers.BlockUser)
	app.Get("/api/admin/orderslist", controllers.OrderList)
	app.Patch("/api/admin/order/cancel/:order_id", controllers.CancelOrder)
	app.Patch("/api/admin/order/:order_id", controllers.ChangeStatusShipped)

	//Category Routes
	app.Post("/api/category", controllers.AddCategory)
	app.Patch("/api/category/:category_id", controllers.EditCategory)
	app.Post("/api/category/:category_id", controllers.DeleteCategory)

	//Product Routes
	app.Post("/api/product", middleware.AdminMiddleware, controllers.AddProduct)
	app.Patch("/api/product/:product_id", middleware.AdminMiddleware, controllers.EditProduct)
	app.Post("/api/product/:product_id", middleware.AdminMiddleware, controllers.DeleteProduct)

	//Address Routes
	app.Post("/api/user/address", middleware.CheckUserStatus, controllers.UserAddAddress)
	app.Patch("/api/user/address/:address_id", middleware.CheckUserStatus, controllers.UserEditAddress)
	app.Delete("/api/user/address/:address_id", middleware.CheckUserStatus, controllers.UserDeleteAddress)
	app.Get("/api/user/address", middleware.CheckUserStatus,controllers.UserListAddress)

	//Cart Routes
	app.Post("api/user/cart/:product_id", middleware.CheckUserStatus, controllers.AddToCart)
	app.Delete("/api/user/cart/:product_id", middleware.CheckUserStatus, controllers.RemoveFromCart)
	app.Get("/api/user/cart", middleware.CheckUserStatus, controllers.ListCartProducts)

	//Order Routes
	app.Post("/api/user/order", middleware.CheckUserStatus, controllers.PlaceOrder)
	app.Patch("/api/user/:order_id", middleware.CheckUserStatus, controllers.CancelOrder)
	app.Get("/api/user/order", middleware.CheckUserStatus, controllers.ListOrdersForUser)
	app.Post("/api/user/order/:order_id", middleware.CheckUserStatus, controllers.ReturnOrder)
	app.Post("/api/user/cancel/product/:order_id/:product_id", middleware.CheckUserStatus,controllers.CancelOneProduct)

	//WishList Routes
	app.Post("/api/user/wishlist/:product_id", middleware.CheckUserStatus, controllers.AddWishList)
	app.Delete("/api/user/wishlist/:product_id", middleware.CheckUserStatus, controllers.RemoveFromWishList)
	app.Get("/api/user/wishlist", middleware.CheckUserStatus, controllers.ListWishList)

	//Coupon Routes
	app.Post("/api/admin/coupon",  controllers.AddCoupon)
	app.Patch("/api/admin/coupon/:coupon_id", middleware.AdminMiddleware, controllers.EditCoupon)
	app.Delete("/api/admin/coupon/:coupon_id", middleware.AdminMiddleware, controllers.DeleteCoupon)
	app.Post("/api/user/coupon/:order_id", middleware.AdminMiddleware, controllers.CancelCoupon)

	//Payment Routes
	app.Get("/api/user/render-razorpay", controllers.RenderRayzorPay)
	app.Get("/api/user/repayment", controllers.RenderRayzorPay)
	app.Post("/api/user/create-order/:order_id", controllers.CreateOrder)
	app.Post("api/user/verify-payment/:order_id", controllers.VerifyPayment)
	app.Post("api/user/failed-handling/:order_id", controllers.FailedHandling)

	//Sales Route
	app.Get("/api/admin/sales", controllers.GetSalesReport)
	app.Get("/api/user/wallet", middleware.CheckUserStatus, controllers.GetWallet)
	app.Get("/api/user/invoice", controllers.InvoiceDownload)
	app.Get("/api/admin/top/products", controllers.TopSellingProducts)
}
