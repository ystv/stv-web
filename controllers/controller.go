package controllers

import (
	"github.com/labstack/echo/v4"
	"github.com/ystv/stv_web/structs"
	"github.com/ystv/stv_web/templates"
	"github.com/ystv/stv_web/utils"
	"net/http"
	"time"
)

// ControllerInterface is the interface to which controllers adhere.
type ControllerInterface interface {
	Get()     //method = GET processing
	Post()    //method = POST processing
	Delete()  //method = DELETE processing
	Put()     //method = PUT handling
	Head()    //method = HEAD processing
	Patch()   //method = PATCH treatment
	Options() //method = OPTIONS processing
	Connect() //method = CONNECT processing
	Trace()   //method = TRACE processing
}

// Controller is the base type of controllers in the 2016site architecture.
type Controller struct {
	pageParams structs.PageParams
	Template   *templates.Templater
	DomainName string
}

func GetController(domainName string) Controller {
	return Controller{
		pageParams: struct {
			GetYear   func() int
			Interface interface{}
		}{
			GetYear: func() int {
				year, _, _ := time.Now().Date()
				return year
			},
		},
		DomainName: domainName,
	}
}

// Get handles a HTTP GET request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Get(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Post handles a HTTP POST request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Post(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Delete handles a HTTP DELETE request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Delete(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Put handles a HTTP PUT request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Put(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Head handles a HTTP HEAD request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Head(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Patch handles a HTTP PATCH request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Patch(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Options handles a HTTP OPTIONS request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Options(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Connect handles a HTTP CONNECT request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Connect(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}

// Trace handles a HTTP TRACE request.
//
// Unless overridden, controllers refuse this method.
func (c *Controller) Trace(eC echo.Context) error {
	return eC.JSON(http.StatusMethodNotAllowed, utils.Error{Error: "Method Not Found"})
}
