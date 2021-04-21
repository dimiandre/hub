package types

import (
	"fmt"
	hubtypes "github.com/sentinel-official/hub/types"
	"testing"
)

func TestProvider_Validate(t *testing.T) {
	correctAddress := hubtypes.ProvAddress{}

	tests := []struct {
		name string
		p    Provider
		want error
	}{
		{"address zero", Provider{"", "", "", "", ""}, errInvalidProvName},
		{"address empty", Provider{correctAddress.String(), "", "", "", ""}, errInvalidProvName},
		{"name length zero", Provider{correctAddress.String(), "", "", "", ""}, errInvalidProvName},
		{"name length greater than 64", Provider{correctAddress.String(), GT64, "", "", ""}, errInvalidProvName},
		{"identity length greater than 64", Provider{correctAddress.String(), "name", GT64, "", ""}, errInvalidProvIdentity},
		{"website length greater than 64", Provider{correctAddress.String(), "name", "", GT64, ""}, errInvalidProvWebsite},
		{"description length greater than 256", Provider{correctAddress.String(), "name", "", "", GT256}, errInvalidProvDescription},
		{"valid", Provider{correctAddress.String(), "name", "", "", ""}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Validate(); err != nil && err.Error() != tt.want.Error() {
				t.Errorf("Validate() error = %v", err)
			}
		})
	}
}

var (
	errInvalidProvAddress     = fmt.Errorf("address should not be nil or empty")
	errInvalidProvName        = fmt.Errorf("name length should be between 1 and 64")
	errInvalidProvIdentity    = fmt.Errorf("identity length should be between 0 and 64")
	errInvalidProvWebsite     = fmt.Errorf("website length should be between 0 and 64")
	errInvalidProvDescription = fmt.Errorf("description length should be between 0 and 256")
)
