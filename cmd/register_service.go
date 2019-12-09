package main

/*
	go run cmd/register_service.go -port=7010 -service-name=checkout -call-service-name=pricing
	go run cmd/register_service.go -port=7020 -service-name=pricing -call-service-name=checkout
*/

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"

	"aimsigh"
)

var port *int
var serviceName *string
var callServiceName *string

func init() {
	port = flag.Int("port", 7010, "port number")
	serviceName = flag.String("service-name", "checkout", "service name")
	callServiceName = flag.String("call-service-name", "pricing", "call service name")
	flag.Parse()
}

func main() {
	fmt.Println("port", *port)
	fmt.Println("service-name", *serviceName)
	fmt.Println("call-service-name", *callServiceName)
	dao := aimsigh.CreateDao()

	go StartServer()
	go keepAndStayAlive(dao)
	for {
		time.Sleep(5 * time.Second)
		host, err := dao.GetServiceAddress("US-PA", *callServiceName)
		if err != nil && !strings.Contains(err.Error(), " servers available") {
			panic(err)
		}
		if host != "" {
			CallService(host)
		}
	}
}

func StartServer() {
	router := echo.New()
	router.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			*serviceName: "ok",
		})
	})
	router.Start(fmt.Sprintf(":%d", *port))
}

func keepAndStayAlive(dao aimsigh.Discoverer) {
	for {
		rec, err := dao.UpdateAliveRecord("US-PA", *serviceName, "127.0.0.1", *port)
		fmt.Println("registration", rec, err)
		time.Sleep(3 * time.Second)
	}
}

func CallService(host string) {
	res, err := http.Get(fmt.Sprintf("http://%s/health", host))
	if err != nil {
		log.Fatal(err)
	}
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Service %s responded with: %s", callServiceName, response)
}
