package app

import (
	"fmt"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/gofiber/fiber/v2/middleware/cors"
	"log"
	"net/http"
	"payment-go/internal/config"
	"payment-go/internal/controllers"
	"payment-go/internal/controllers/analytics"
	"payment-go/internal/controllers/crud"
	"payment-go/internal/controllers/webhook"
	"payment-go/internal/database"
	"payment-go/internal/models"
	"payment-go/internal/providers"
	"payment-go/internal/services"
	"payment-go/internal/utils/proxy"
	"strconv"
	"sync"
	"time"
)

type IApp interface {
	Prepare() error
	Launch() error
}
type app struct {
	config    *config.Config
	fiber     *fiber.App
	taskQueue proxy.ITaskQueue
}

var appOnce = sync.Once{}
var appIns IApp

func GetApp() IApp {
	appOnce.Do(func() {
		appIns = &app{
			fiber:     fiber.New(),
			taskQueue: proxy.NewTaskQueue(10, 100*time.Millisecond),
		}
	})
	return appIns
}

func (a *app) Prepare() error {
	a.config = config.GetConfig()

	_, err := database.NewConnection(a.config)
	if err != nil {
		return err
	}

	err = database.MakeMigrations(models.GetModels())
	if err != nil {
		return err
	}

	a.fiber.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
	}))

	a.makeRoutes()

	a.init()

	return nil
}

func (a *app) prepareTaskQueue() {
	providers.TaskQueueProvider()
}

func (a *app) makeRoutes() {
	a.fiber.Get("/", controllers.IndexController().Home)

	// Public API routes
	api := a.fiber.Group("/api")
	func() {

		// create order and get wait-link
		api.Post("/order/create", controllers.OrderController().Create)

		// get payment info
		api.Get("/order/payment-info", controllers.OrderController().GetPaymentInfo)

		// check order status
		api.Get("/order/:order_number/check-status", controllers.OrderController().CheckStatus)

		// wait-link SSE
		api.Get(
			"/order/wait-link",
			adaptor.HTTPHandler(
				http.HandlerFunc(controllers.OrderController().CreateLinkSSE),
			),
		)
		// wait-link page
		api.Get("/order/:order_number/process_to_payment", controllers.IndexController().Payment)

		// withdraw
		api.Post("/withdraw/create", controllers.WithdrawController().Create)
	}()

	// API Webhooks (public)
	wh := api.Group("/webhook") // /api/webhook/*
	func() {

		// bank webhook via sms
		wh.Post("/bank/payment-info", webhook.BankWebhookController().PaymentInfo)

	}()

	// CRUD routes
	func() {
		group := a.fiber.Group("/crud")
		cruds := map[string]crud.ICrudController{
			"order":        crud.OrderCrudController(),
			"payment_link": crud.PaymentLinkCrudController(),
			"card":         crud.CardCrudController(),
			"shop":         crud.ShopCrudController(),
			"bank_message": crud.BankMessageCrudController(),
			"withdraw":     crud.WithdrawCrudController(),
		}
		for prefix, crud := range cruds {
			rules := crud.GetActions()
			if rules.Create {
				group.Post(fmt.Sprintf("/%s/create", prefix), crud.Create)
			}
			if rules.Read {
				group.Get(fmt.Sprintf("/%s/read", prefix), crud.Read)
			}
			if rules.Update {
				group.Post(fmt.Sprintf("/%s/update", prefix), crud.Update)
			}
			if rules.Delete {
				group.Post(fmt.Sprintf("/%s/delete", prefix), crud.Delete)
			}
			if rules.List {
				group.Get(fmt.Sprintf("/%s/list", prefix), crud.List)
			}
		}

		// card
		group.Get("/card/check-activity", crud.CardCrudController().CheckActivity)
		group.Post("/card/change-balance", crud.CardCrudController().ChangeBalance)

		// shop
		group.Post("/shop/regenerate-private-key", crud.ShopCrudController().RegeneratePrivateKey)
		group.Post("/shop/validate-host", crud.ShopCrudController().ValidateShopHost)

		// bank message
		group.Post("/bank_message/approve", crud.BankMessageCrudController().Approve)
		group.Post("/bank_message/decline", crud.BankMessageCrudController().Decline)

		// withdraw
		group.Post("/withdraw/approve", crud.WithdrawCrudController().Approve)
		group.Post("/withdraw/decline", crud.WithdrawCrudController().Decline)
		group.Post("/withdraw/process", crud.WithdrawCrudController().Process)

		// filters
		group.Post("/order/find", crud.OrderCrudController().Find)
		group.Post("/shop/find", crud.ShopCrudController().Find)
		group.Post("/payment_link/find", crud.PaymentLinkCrudController().Find)
		group.Post("/bank_message/find", crud.BankMessageCrudController().Find)
		group.Post("/withdraw/find", crud.WithdrawCrudController().Find)
	}()

	// Analytics routes
	func() {
		group := a.fiber.Group("/analytics")

		group.Post("/card", analytics.CardAnalyticsController().Totals)
	}()

	// create order and get wait-link
	a.fiber.Post("/order/create", controllers.OrderController().Create)
	// check order status
	a.fiber.Get("/order/:order_id/check-status", controllers.OrderController().CheckStatus)
}

func (a *app) init() {
	// init services
	services.BankApiService()
	services.CardService()
	services.OrderService()
	errors := services.PaymentLinkService().LoadPendingLinks()
	if errors != nil {
		for _, err := range errors {
			log.Println(err)
		}
	}
}

func (a *app) Launch() error {
	return a.fiber.Listen(":" + strconv.Itoa(int(a.config.AppPort)))
}
