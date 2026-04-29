package validator

import (
	"regexp"
	"testing"
)

func TestNotBlank(t *testing.T) {
	if NotBlank("   ") {
		t.Fatal("expected whitespace-only string to fail NotBlank")
	}
	if !NotBlank("go") {
		t.Fatal("expected non-empty string to pass NotBlank")
	}
}

func TestMaxCharsAndMinChars(t *testing.T) {
	if !MaxChars("abc", 3) {
		t.Fatal("expected equal-length value to pass MaxChars")
	}
	if MaxChars("abcd", 3) {
		t.Fatal("expected longer value to fail MaxChars")
	}
	if !MinChars("abcd", 4) {
		t.Fatal("expected equal-length value to pass MinChars")
	}
	if MinChars("ab", 3) {
		t.Fatal("expected shorter value to fail MinChars")
	}
}

func TestPermittedInt(t *testing.T) {
	if !PermittedInt(7, 1, 7, 365) {
		t.Fatal("expected value in set to pass")
	}
	if PermittedInt(2, 1, 7, 365) {
		t.Fatal("expected value outside set to fail")
	}
}

func TestMatches(t *testing.T) {
	rx := regexp.MustCompile(`^g.*o$`)
	if !Matches("go", rx) {
		t.Fatal("expected match")
	}
	if Matches("forum", rx) {
		t.Fatal("expected non-match")
	}
}

func TestIncorrectInput(t *testing.T) {
	if !IncorrectInput("go, forum2026") {
		t.Fatal("expected valid tags to pass")
	}
	if IncorrectInput("@bad") {
		t.Fatal("expected invalid tags to fail")
	}
}

func TestValidatorErrorCollection(t *testing.T) {
	v := Validator{}
	if !v.Valid() {
		t.Fatal("new validator should be valid")
	}

	v.CheckField(false, "email", "bad email")
	v.CheckField(false, "email", "second message ignored")
	v.AddNonFieldError("global error")

	if v.Valid() {
		t.Fatal("validator with errors should be invalid")
	}
	if got := v.FieldErrors["email"]; got != "bad email" {
		t.Fatalf("expected first field error, got %q", got)
	}
	if len(v.NonFieldErrors) != 1 || v.NonFieldErrors[0] != "global error" {
		t.Fatalf("unexpected non-field errors: %v", v.NonFieldErrors)
	}
}
