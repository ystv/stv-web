package cookie

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

func SessionMiddleware(store sessions.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("_session_store", store)
			return next(c)
		}
	}
}

func GetSession(c echo.Context, name string) (*sessions.Session, error) {
	store, ok := c.Get("_session_store").(sessions.Store)
	if !ok {
		return nil, echo.NewHTTPError(500, "Cannot get session store")
	}
	return store.Get(c.Request(), name)
}
