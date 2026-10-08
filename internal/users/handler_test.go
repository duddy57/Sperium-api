package users

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeService struct {
	createID  uuid.UUID
	createErr error
	loginErr  error
	logoutErr error
	details   Users
	detailErr error
	deleteErr error
	updateErr error
}

func (f *fakeService) Create(_ context.Context, _ CreateUserReq) (uuid.UUID, error) {
	return f.createID, f.createErr
}
func (f *fakeService) Login(_ context.Context, _ LoginUserReq) error   { return f.loginErr }
func (f *fakeService) Logout(_ context.Context) error                  { return f.logoutErr }
func (f *fakeService) Details(_ context.Context) (Users, error)        { return f.details, f.detailErr }
func (f *fakeService) Delete(_ context.Context) error                  { return f.deleteErr }
func (f *fakeService) Update(_ context.Context, _ UpdateUserReq) error { return f.updateErr }

func newTestHandler(service *fakeService) *Handler {
	return NewUserHandler(service, validator.New())
}

func request(handler http.HandlerFunc, method, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	response := httptest.NewRecorder()
	handler(response, req)
	return response
}

func assertErrorResponse(t *testing.T, response *httptest.ResponseRecorder, status int) {
	t.Helper()
	assert.Equal(t, status, response.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "try again later", body["error"])
}

func TestHandlerCreateSuccess(t *testing.T) {
	id := uuid.New()
	handler := newTestHandler(&fakeService{createID: id})
	response := request(handler.create, http.MethodPost, `{"name":"Ada","email":"ada@example.com","password":"password"}`)

	assert.Equal(t, http.StatusCreated, response.Code)
	var body createUserResp
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, "Welcome to sperium!", body.Message)
	assert.Equal(t, id, body.ID)
}

func TestHandlerCreateInvalidJSON(t *testing.T) {
	response := request(newTestHandler(&fakeService{}).create, http.MethodPost, `{invalid`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid json body"`)
}

func TestHandlerCreateInvalidValidation(t *testing.T) {
	response := request(newTestHandler(&fakeService{}).create, http.MethodPost, `{}`)

	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid json body"`)
}

func TestHandlerCreateServiceError(t *testing.T) {
	response := request(newTestHandler(&fakeService{createErr: errors.New("create failed")}).create, http.MethodPost, `{"name":"Ada","email":"ada@example.com","password":"password"}`)

	assertErrorResponse(t, response, http.StatusInternalServerError)
}

func TestHandlerLoginSuccess(t *testing.T) {
	handler := newTestHandler(&fakeService{})
	response := request(handler.login, http.MethodPost, `{"email":"ada@example.com","password":"password"}`)

	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Contains(t, response.Body.String(), `"message":"Welcome to sperium!"`)
}

func TestHandlerLoginInvalidJSON(t *testing.T) {
	response := request(newTestHandler(&fakeService{}).login, http.MethodPost, `{invalid`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid json body"`)
}

func TestHandlerLoginInvalidValidation(t *testing.T) {
	response := request(newTestHandler(&fakeService{}).login, http.MethodPost, `{}`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid json body"`)
}

func TestHandlerLoginServiceError(t *testing.T) {
	response := request(newTestHandler(&fakeService{loginErr: errors.New("login failed")}).login, http.MethodPost, `{"email":"ada@example.com","password":"password"}`)

	assertErrorResponse(t, response, http.StatusInternalServerError)
}

func TestHandlerLogoutSuccess(t *testing.T) {
	response := request(newTestHandler(&fakeService{}).logout, http.MethodGet, "")

	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Contains(t, response.Body.String(), `"message":"Bye bye!"`)
}

func TestHandlerLogoutServiceError(t *testing.T) {
	response := request(newTestHandler(&fakeService{logoutErr: errors.New("logout failed")}).logout, http.MethodGet, "")

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid json body"`)
}

func TestHandlerDetailsSuccess(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	handler := newTestHandler(&fakeService{details: Users{ID: id, Name: "Ada", Email: "ada@example.com", CreatedAt: now, UpdatedAt: now}})
	response := request(handler.details, http.MethodGet, "")

	assert.Equal(t, http.StatusOK, response.Code)
	var body Users
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, id, body.ID)
	assert.Equal(t, "Ada", body.Name)
	assert.Equal(t, "ada@example.com", body.Email)
}

func TestHandlerDetailsServiceError(t *testing.T) {
	response := request(newTestHandler(&fakeService{detailErr: errors.New("details failed")}).details, http.MethodGet, "")

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid json body"`)
}

func TestHandlerDeleteSuccess(t *testing.T) {
	response := request(newTestHandler(&fakeService{}).delete, http.MethodDelete, "")

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), `"message":"Bye bye!"`)
}

func TestHandlerDeleteServiceError(t *testing.T) {
	response := request(newTestHandler(&fakeService{deleteErr: errors.New("delete failed")}).delete, http.MethodDelete, "")

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid json body"`)
}

func TestHandlerUpdateSuccess(t *testing.T) {
	response := request(newTestHandler(&fakeService{}).update, http.MethodPut, `{"name":"Ada"}`)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Contains(t, response.Body.String(), `"message":"Update success!"`)
}

func TestHandlerUpdateInvalidJSON(t *testing.T) {
	response := request(newTestHandler(&fakeService{}).update, http.MethodPut, `{invalid`)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid json body"`)
}

func TestHandlerUpdateInvalidValidation(t *testing.T) {
	response := request(newTestHandler(&fakeService{}).update, http.MethodPut, `{}`)

	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid json body"`)
}

func TestHandlerUpdateServiceError(t *testing.T) {
	response := request(newTestHandler(&fakeService{updateErr: errors.New("update failed")}).update, http.MethodPut, `{"name":"Ada"}`)

	assertErrorResponse(t, response, http.StatusInternalServerError)
}

var _ userService = (*fakeService)(nil)
