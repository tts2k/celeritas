package celeritas

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

type Valildation struct {
	Data   url.Values
	Errors map[string]string
}

func (c *Celeritas) Validator(data url.Values) *Valildation {
	return &Valildation{
		Errors: make(map[string]string),
		Data:   data,
	}
}

func (v *Valildation) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Valildation) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Valildation) Has(field string, r *http.Request) bool {
	x := r.Form.Get(field)
	return x != ""
}

func (v *Valildation) Required(r *http.Request, fields ...string) {
	for _, field := range fields {
		value := r.Form.Get(field)
		if strings.TrimSpace(value) == "" {
			v.AddError(field, "This field cannot be blank")
		}
	}
}

func (v *Valildation) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func (v *Valildation) IsEmail(field, value string) {
	if !govalidator.IsEmail(value) {
		v.AddError(field, "Invalid email address")
	}
}

func (v *Valildation) IsInt (field, value string) {
	_, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, "This field must be an integer")
	}
}

func (v *Valildation) IsFloat(field, value string) {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		v.AddError(field, "This field must be an floating point number")
	}
}

func (v *Valildation) IsDateISO(field, value string) {
	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		v.AddError(field, "This field must be date in the form of YYYY-MM-DD")
	}
}

func (v *Valildation) NoSpaces(field, value string) {
	if govalidator.HasWhitespace(value) {
		v.AddError(field, "Spaces are not permitted")
	}
}
