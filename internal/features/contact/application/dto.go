package application

import "github.com/sraj/addressbook/internal/features/contact/domain"

type AddressRequest struct {
	Label   string `json:"label"   validate:"required,oneof=Home Office Other"`
	Line1   string `json:"line1"   validate:"required,max=200"`
	Line2   string `json:"line2"   validate:"omitempty,max=200"`
	City    string `json:"city"    validate:"required,max=100"`
	State   string `json:"state"   validate:"required,max=100"`
	Zip     string `json:"zip"     validate:"required,max=20"`
	Country string `json:"country" validate:"required,max=100"`
}

func (a AddressRequest) ToDomain() domain.Address {
	return domain.Address{
		Label:   a.Label,
		Line1:   a.Line1,
		Line2:   a.Line2,
		City:    a.City,
		State:   a.State,
		Zip:     a.Zip,
		Country: a.Country,
	}
}

type CreateRequest struct {
	Name      string           `json:"name"      validate:"required,min=1,max=200"`
	Emails    []string         `json:"emails"    validate:"required,min=1,dive,email"`
	Phones    []string         `json:"phones"    validate:"required,min=1,dive,max=30"`
	Addresses []AddressRequest `json:"addresses" validate:"required,min=1,dive"`
	Notes     string           `json:"notes"     validate:"omitempty,max=2000"`
}

type UpdateRequest struct {
	Name      string           `json:"name"      validate:"required,min=1,max=200"`
	Emails    []string         `json:"emails"    validate:"required,min=1,dive,email"`
	Phones    []string         `json:"phones"    validate:"required,min=1,dive,max=30"`
	Addresses []AddressRequest `json:"addresses" validate:"required,min=1,dive"`
	Notes     string           `json:"notes"     validate:"omitempty,max=2000"`
}
