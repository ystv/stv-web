package controllers

import "github.com/labstack/echo/v4"

type AdminRepo struct {
	controller Controller
}

func NewAdminRepo(controller Controller) *AdminRepo {
	return &AdminRepo{
		controller: controller,
	}
}

func (r *AdminRepo) Admin(c echo.Context) error {
	return nil
}
