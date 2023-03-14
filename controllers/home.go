package controllers

import (
	"github.com/labstack/echo/v4"
)

type HomeRepo struct {
	controller Controller
}

func NewHomeRepo(controller Controller) *HomeRepo {
	return &HomeRepo{
		controller: controller,
	}
}

func (r *HomeRepo) Home(c echo.Context) error {
	err := r.controller.Template.RenderTemplate(c.Response().Writer, r.controller.pageParams, nil, "home.tmpl")
	if err != nil {
		return err
	}
	return nil
}
