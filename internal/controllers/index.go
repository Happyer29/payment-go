package controllers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"sync"
)

type IIndexController interface {
	Home(ctx *fiber.Ctx) error
	Test(ctx *fiber.Ctx) error
	Payment(ctx *fiber.Ctx) error
}
type indexController struct {
}

var indexControllerIns IIndexController
var indexControllerOnce = sync.Once{}

func IndexController() IIndexController {
	indexControllerOnce.Do(func() {
		indexControllerIns = &indexController{}
	})
	return indexControllerIns
}

func (con *indexController) Home(ctx *fiber.Ctx) error {
	return ctx.SendString("hello")
}

func (con *indexController) Test(ctx *fiber.Ctx) error {
	client := http.Client{}
	res, err := client.Get("http://laravel/")
	if err != nil {
		fmt.Println(err)
		return ctx.SendString("error")
	}

	var data []byte
	_, err = res.Body.Read(data)
	if err != nil {
		fmt.Println(err)
		return ctx.SendString("error")
	}

	fmt.Println(string(data))

	return ctx.SendString("hello from controller")
}

func (con *indexController) Payment(ctx *fiber.Ctx) error {
	orderNumber := ctx.Params("order_number")
	html := `<!DOCTYPE html><html>
	<head><title>Processing...</title></head>
	<body>
	<script type="text/javascript">
        const ev = new EventSource('/api/order/wait-link?order_number=%s')
        ev.onmessage = (event) => {
            try{
				res = JSON.parse(event.data)
				if(res.link) location.href = res.link
				else ev.close()
			}catch(e){
				console.log(e)
				ev.close()
			}
        }
		ev.onclose = () => {
			//history.back()
			alert('go back')
		}
        console.log(ev)
	</script>
	</body>
	</html>`
	ctx.Response().Header.Set("Content-Type", "text/html")
	return ctx.SendString(fmt.Sprintf(html, orderNumber))
}
