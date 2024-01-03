package controllers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/stv_web/storage"
	"github.com/ystv/stv_web/store"
	"github.com/ystv/stv_web/templates"
)

type RegistrationRepo struct {
	controller Controller
	store      *store.Store
}

func NewRegistrationRepo(controller Controller, store *store.Store) *RegistrationRepo {
	return &RegistrationRepo{
		controller: controller,
		store:      store,
	}
}

func (r *RegistrationRepo) Register(c echo.Context) error {
	stv, err := r.store.Get()
	if err != nil {
		return r.errorHandle(c, err)
	}

	if !stv.GetAllowRegistration() {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	err = r.controller.Template.RenderTemplate(c.Response().Writer, nil, templates.RegistrationTemplate)
	if err != nil {
		return err
	}
	return nil
}

func (r *RegistrationRepo) QR(c echo.Context) error {
	stv, err := r.store.Get()
	if err != nil {
		return r.errorHandle(c, err)
	}

	if !stv.GetAllowRegistration() {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	err = r.controller.Template.RenderTemplate(c.Response().Writer, nil, templates.QRTemplate)
	if err != nil {
		return err
	}
	return nil
}

func (r *RegistrationRepo) AddVoter(c echo.Context) error {
	stv, err := r.store.Get()
	if err != nil {
		return r.errorHandle(c, err)
	}

	if !stv.GetAllowRegistration() {
		return c.Redirect(http.StatusTemporaryRedirect, "/")
	}

	err = c.Request().ParseForm()
	if err != nil {
		return r.errorHandle(c, err)
	}
	email := c.Request().FormValue("email")
	name := c.Request().FormValue("name")
	if len(name) == 0 || len(email) == 0 {
		return r.errorHandle(c, fmt.Errorf("name and email need to be filled"))
	}
	_, err = r.store.AddVoter(&storage.Voter{
		Email: email,
		Name:  name,
	})
	if err != nil {
		return r.errorHandle(c, err)
	}

	err = r.controller.Template.RenderTemplate(c.Response().Writer, nil, templates.RegisteredTemplate)
	if err != nil {
		return r.errorHandle(c, err)
	}

	return nil
}

func (r *RegistrationRepo) errorHandle(c echo.Context, err error) error {
	data := struct {
		Error string
	}{
		Error: err.Error(),
	}
	fmt.Println(data.Error)
	err = r.controller.Template.RenderTemplate(c.Response().Writer, data, templates.RegistrationErrorTemplate)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
