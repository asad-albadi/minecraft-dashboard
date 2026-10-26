package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestReadOnlyRoutesRequireASession(t *testing.T) {
	app := fiber.New()
	a, _ := accountTestApp(t, "viewer")
	a.Register(app)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/logs", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("GET /api/v1/logs status = %d, want %d", response.StatusCode, http.StatusUnauthorized)
	}
}

func TestSessionLoginAuthorizesReadOnlyRoutes(t *testing.T) {
	app := fiber.New()
	a, _ := accountTestApp(t, "viewer")
	a.Register(app)

	login := httptest.NewRequest(http.MethodPost, "/api/v1/session", bytes.NewBufferString(`{"username":"alex","password":"correct-password"}`))
	login.Header.Set("Content-Type", "application/json")
	login.Header.Set("Origin", "https://dashboard.test")
	loginResponse, err := app.Test(login)
	if err != nil {
		t.Fatal(err)
	}
	defer loginResponse.Body.Close()
	if loginResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("POST /api/v1/session status = %d, want %d", loginResponse.StatusCode, http.StatusNoContent)
	}
	cookies := loginResponse.Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].Value == "" {
		t.Fatalf("login cookie = %#v, want one non-empty HttpOnly cookie", cookies)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/logs", nil)
	request.AddCookie(cookies[0])
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("authenticated GET /api/v1/logs status = %d, want %d", response.StatusCode, http.StatusOK)
	}
}
