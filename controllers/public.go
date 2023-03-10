package controllers

import (
	"github.com/labstack/echo/v4"
)

type PublicRepo struct {
	controller Controller
}

func NewPublicRepo(controller Controller) *PublicRepo {
	return &PublicRepo{
		controller: controller,
	}
}

func (r *PublicRepo) Public(c echo.Context) error {
	return c.Inline("public/"+c.Param("file"), c.Param("file"))
}

func (r *PublicRepo) PublicFontArial(c echo.Context) error {
	return c.Inline("public/webfonts/Arial/"+c.Param("file"), c.Param("file"))
}

func (r *PublicRepo) PublicFontAllerta(c echo.Context) error {
	return c.Inline("public/webfonts/Allerta/"+c.Param("file"), c.Param("file"))
}
