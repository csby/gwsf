package gtype

import (
	"testing"
)

func TestIsEmailFormat(t *testing.T) {
	if IsEmailFormat("") {
		t.Error("empty is invalid email")
	}

	isValid := IsEmailFormat("nameAttest.com")
	if isValid {
		t.Error("not @ character")
	}

	isValid = IsEmailFormat("nameG@test.com")
	if !isValid {
		t.Error("invalid email")
	}
}

func TestIsMobileFormat(t *testing.T) {
	if IsMobileFormat("") {
		t.Error("empty is invalid mobile")
	}

	isValid := IsMobileFormat("11188886666")
	if isValid {
		t.Error("invalid begin with 111")
	}

	isValid = IsMobileFormat("188666888")
	if isValid {
		t.Error("invalid count")
	}

	isValid = IsMobileFormat("1886668888")
	if isValid {
		t.Error("invalid count")
	}

	isValid = IsMobileFormat("188666688886")
	if isValid {
		t.Error("invalid count")
	}

	isValid = IsMobileFormat("18866668888")
	if !isValid {
		t.Error("valid mobile")
	}
}

func TestIsAccountFormat(t *testing.T) {
	if IsAccountFormat("") {
		t.Error("empty is invalid account")
	}

	isValid := IsAccountFormat("18866668888")
	if isValid {
		t.Error("phone number is invalid account")
	}

	isValid = IsAccountFormat("name@test.com")
	if isValid {
		t.Error("email is invalid account")
	}

	isValid = IsAccountFormat("Admin")
	if !isValid {
		t.Error("'Admin' should be valid account")
	}

	isValid = IsAccountFormat("x.m")
	if !isValid {
		t.Error("'x.m' should be valid account")
	}

	isValid = IsAccountFormat("x-m")
	if !isValid {
		t.Error("'x-m' should be valid account")
	}

	isValid = IsAccountFormat("8x_m")
	if !isValid {
		t.Error("'8x_m' should be valid account")
	}

	isValid = IsAccountFormat("64222419720913083Ⅹ")
	if !isValid {
		t.Error("'64222419720913083Ⅹ' should be valid account")
	}

	isValid = IsAccountFormat("64222419720913083x")
	if !isValid {
		t.Error("'64222419720913083x' should be valid account")
	}

	isValid = IsAccountFormat("64222419720913083")
	if !isValid {
		t.Error("'6422241972091308x' should be valid account")
	}

	isValid = IsAccountFormat("1A")
	if !isValid {
		t.Error("'1A' should be valid account")
	}

	isValid = IsAccountFormat("a1")
	if !isValid {
		t.Error("'a1' should be valid account")
	}

	isValid = IsAccountFormat("1")
	if !isValid {
		t.Error("'1' should be valid account")
	}

	isValid = IsAccountFormat("@")
	if isValid {
		t.Error("'@' should be invalid account")
	}

	isValid = IsAccountFormat("n_")
	if !isValid {
		t.Error("'n_' should be valid account")
	}
	isValid = IsAccountFormat("_n")
	if isValid {
		t.Error("'_n' should be invalid account")
	}
	isValid = IsAccountFormat("#")
	if isValid {
		t.Error("'#' should be invalid account")
	}
	isValid = IsAccountFormat("#_")
	if isValid {
		t.Error("'#_' should be invalid account")
	}
}
