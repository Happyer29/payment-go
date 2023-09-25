package validation

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"payment-go/internal/models"
	"strings"
)

type IShopValidator interface {
	GetRouteForValidation() string
	Validate() bool
	ValidateMetaTag() bool
	ValidateRoute() bool
}
type shopValidator struct {
	opt    ShopValidationOptions
	client *http.Client
	url    *url.URL
}

type ShopValidationOptions struct {
	ShopUrl     string
	ConfirmCode string
}

func NewValidator(opt ShopValidationOptions) (IShopValidator, error) {
	if !models.ValidateHost(opt.ShopUrl) {
		return nil, fmt.Errorf("invalid shop host")
	}

	u, err := url.Parse(opt.ShopUrl)
	if err != nil {
		return nil, err
	}

	return &shopValidator{
		opt:    opt,
		client: &http.Client{},
		url:    u,
	}, nil
}

func (sv *shopValidator) Validate() bool {
	if sv.ValidateMetaTag() {
		return true
	}
	if sv.ValidateRoute() {
		return true
	}
	return false
}

func (sv *shopValidator) ValidateMetaTag() bool {
	res, err := sv.client.Get(sv.opt.ShopUrl)
	if err != nil {
		return false
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}

	return strings.Contains(string(body), sv.opt.ConfirmCode)
}

func (sv *shopValidator) ValidateRoute() bool {
	res, err := sv.client.Get(sv.GetRouteForValidation())
	if err != nil {
		return false
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}

	return string(body) == sv.opt.ConfirmCode
}

func (sv *shopValidator) GetRouteForValidation() string {
	return sv.opt.ShopUrl + "/" + sv.opt.ConfirmCode
}
