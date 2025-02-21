package controllers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/ystv/stv_web/templates"
	"github.com/ystv/stv_web/utils"
)

// ControllerInterface is the interface to which controllers adhere.
type ControllerInterface interface {
	Get(echo.Context) error     // method = GET processing
	Post(echo.Context) error    // method = POST processing
	Delete(echo.Context) error  // method = DELETE processing
	Put(echo.Context) error     // method = PUT handling
	Head(echo.Context) error    // method = HEAD processing
	Patch(echo.Context) error   // method = PATCH treatment
	Options(echo.Context) error // method = OPTIONS processing
	Connect(echo.Context) error // method = CONNECT processing
	Trace(echo.Context) error   // method = TRACE processing
	encrypt([]byte) ([]byte, error)
	decrypt([]byte) ([]byte, error)
}

var _ ControllerInterface = &Controller{}

// Controller is the base type of controllers.
type Controller struct {
	Template     *templates.Templater
	DomainName   string
	_cipherBlock cipher.Block
}

func GetController(domainName string, encryptionKey []byte) (Controller, error) {
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return Controller{}, err
	}

	return Controller{
		DomainName:   domainName,
		_cipherBlock: block,
	}, nil
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

func (c *Controller) encrypt(text []byte) ([]byte, error) {
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(c._cipherBlock, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func (c *Controller) decrypt(text []byte) ([]byte, error) {
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(c._cipherBlock, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, err
	}
	return data, nil
}
