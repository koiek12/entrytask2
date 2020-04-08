package jwt

import (
	"testing"
	"time"
)

func TestAuthenticate(t *testing.T) {
	issuer := NewTokenIssuer("valid", time.Hour)
	token := issuer.GenerateToken("id")
	id, err := issuer.AuthenticateToken(token)
	t.Logf(id)
	if id != "id" || err != nil {
		t.Fail()
	}
}

func TestAuthenticateWithWrongKey(t *testing.T) {
	issuer := NewTokenIssuer("valid", time.Hour)
	issuer2 := NewTokenIssuer("invalid", time.Hour)

	token := issuer.GenerateToken("id")
	id, err := issuer2.AuthenticateToken(token)
	if id == "" && err != nil {
		t.Logf(err.Error())
	} else {
		t.Fail()
	}
}
func TestExpiredKey(t *testing.T) {
	issuer := NewTokenIssuer("valid", time.Millisecond)
	token := issuer.GenerateToken("id")
	time.Sleep(time.Second)

	id, err := issuer.AuthenticateToken(token)
	if id == "" && err != nil {
		t.Logf(err.Error())
	} else {
		t.Fail()
	}
}
