package controllers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/templates"
)

type HomeRepo struct {
	controller Controller
	store      *store.Store
}

func NewHomeRepo(controller Controller, store *store.Store) *HomeRepo {
	return &HomeRepo{
		controller: controller,
		store:      store,
	}
}

func (r *HomeRepo) Home(c echo.Context) error {
	allow, err := r.store.GetAllowRegistration()
	if err != nil {
		fmt.Println(err)
		allow = false
	}
	data := struct {
		AllowRegistration bool
		URL               string
	}{
		AllowRegistration: allow,
		URL:               "https://" + r.controller.DomainName + "/registration",
	}
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.HomeTemplate)
	if err != nil {
		return err
	}
	return nil
}
