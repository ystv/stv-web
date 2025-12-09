package controllers

import (
	"errors"
	"log"

	"github.com/labstack/echo/v4"

	"github.com/ystv/stv-web/templates"
)

type ErrorRepo struct {
	controller Controller
}

func NewErrorRepo(controller Controller) *ErrorRepo {
	return &ErrorRepo{
		controller: controller,
	}
}

func (v *ErrorRepo) Error404(c echo.Context) error {
	return v.controller.Template.RenderTemplate(c.Response().Writer, nil, templates.NotFound404Template)
}

func (v *ErrorRepo) CustomHTTPErrorHandler(err error, c echo.Context) {
	log.Println(err)
	var he *echo.HTTPError
	var status int
	if errors.As(err, &he) {
		status = he.Code
	} else {
		status = 500
	}
	c.Response().WriteHeader(status)
	data := struct {
		Code  int
		Error interface{}
	}{
		Code:  status,
		Error: he.Message,
	}
	err1 := v.controller.Template.RenderTemplate(c.Response().Writer, data, templates.ErrorTemplate)
	if err1 != nil {
		log.Printf("failed to render error page: %+v", err1)
	}
}
