package organizations

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

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeOrganizationService struct {
	createdID    uuid.UUID
	createdSlug  string
	createErr    error
	organization Organization
	detailsErr   error
	defaultErr   error
	deleteErr    error
	updateErr    error
}

func (f *fakeOrganizationService) Create(context.Context, CreateOrganizationReq) (uuid.UUID, string, error) {
	return f.createdID, f.createdSlug, f.createErr
}
func (f *fakeOrganizationService) Details(context.Context, uuid.UUID) (Organization, error) {
	return f.organization, f.detailsErr
}
func (f *fakeOrganizationService) Delete(context.Context, uuid.UUID) error { return f.deleteErr }
func (f *fakeOrganizationService) GetDefault(context.Context) (Organization, error) {
	return f.organization, f.defaultErr
}
func (f *fakeOrganizationService) Update(context.Context, uuid.UUID, UpdateOrganizationReq) error {
	return f.updateErr
}

func organizationRequest(h http.HandlerFunc, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	response := httptest.NewRecorder()
	h(response, req)
	return response
}

func organizationIDRequest(h http.HandlerFunc, method, id, body string) *httptest.ResponseRecorder {
	route := chi.NewRouteContext()
	route.URLParams.Add("id", id)
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, route))
	response := httptest.NewRecorder()
	h(response, req)
	return response
}
func organizationHandler(service *fakeOrganizationService) *Handler {
	return NewOrganizationHandler(service, validator.New())
}

func TestHandlerCreate(t *testing.T) {
	id := uuid.New()
	response := organizationRequest(organizationHandler(&fakeOrganizationService{createdID: id, createdSlug: "acme"}).create, http.MethodPost, "/", `{"name":"Acme","domain":"acme.test"}`)
	assert.Equal(t, http.StatusCreated, response.Code)
	var body createOrganizationResp
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, id, body.ID)
	assert.Equal(t, "acme", body.Slug)
}

func TestHandlerCreateRejectsInvalidJSONAndValidation(t *testing.T) {
	response := organizationRequest(organizationHandler(&fakeOrganizationService{}).create, http.MethodPost, "/", `{bad`)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	response = organizationRequest(organizationHandler(&fakeOrganizationService{}).create, http.MethodPost, "/", `{}`)
	assert.Equal(t, http.StatusUnprocessableEntity, response.Code)
}

func TestHandlerCreateServiceError(t *testing.T) {
	response := organizationRequest(organizationHandler(&fakeOrganizationService{createErr: errors.New("failed")}).create, http.MethodPost, "/", `{"name":"Acme","domain":"acme.test"}`)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"try again later"`)
}

func TestHandlerGetSuccessAndErrors(t *testing.T) {
	id := uuid.New()
	response := organizationIDRequest(organizationHandler(&fakeOrganizationService{organization: Organization{ID: id, Name: "Acme", CreatedAt: time.Now()}}).get, http.MethodGet, id.String(), "")
	assert.Equal(t, http.StatusOK, response.Code)
	response = organizationIDRequest(organizationHandler(&fakeOrganizationService{detailsErr: errors.New("failed")}).get, http.MethodGet, id.String(), "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestHandlerGetInvalidID(t *testing.T) {
	r := chi.NewRouteContext()
	r.URLParams.Add("id", "not-a-uuid")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, r))
	response := httptest.NewRecorder()
	organizationHandler(&fakeOrganizationService{}).get(response, req)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body.String(), `"error":"invalid organization id"`)
}

func TestHandlerUpdateAndDeleteInvalidID(t *testing.T) {
	h := organizationHandler(&fakeOrganizationService{})
	response := organizationRequest(func(w http.ResponseWriter, r *http.Request) {
		route := chi.NewRouteContext()
		route.URLParams.Add("id", "not-a-uuid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, route))
		h.update(w, r)
	}, http.MethodPut, "/", `{"name":"Acme","domain":"acme.test"}`)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	response = organizationRequest(func(w http.ResponseWriter, r *http.Request) {
		route := chi.NewRouteContext()
		route.URLParams.Add("id", "not-a-uuid")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, route))
		h.delete(w, r)
	}, http.MethodDelete, "/", "")
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestHandlerUpdateDeleteAndDefault(t *testing.T) {
	id := uuid.New()
	h := organizationHandler(&fakeOrganizationService{organization: Organization{ID: id}, updateErr: nil})
	response := organizationIDRequest(h.update, http.MethodPut, id.String(), `{"name":"Acme","domain":"acme.test"}`)
	assert.Equal(t, http.StatusCreated, response.Code)
	response = organizationIDRequest(organizationHandler(&fakeOrganizationService{updateErr: errors.New("failed")}).update, http.MethodPut, id.String(), `{"name":"Acme","domain":"acme.test"}`)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	response = organizationIDRequest(organizationHandler(&fakeOrganizationService{}).delete, http.MethodDelete, id.String(), "")
	assert.Equal(t, http.StatusCreated, response.Code)
	response = organizationIDRequest(organizationHandler(&fakeOrganizationService{deleteErr: errors.New("failed")}).delete, http.MethodDelete, id.String(), "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	response = organizationRequest(organizationHandler(&fakeOrganizationService{organization: Organization{ID: id}}).getDefault, http.MethodGet, "/", "")
	assert.Equal(t, http.StatusOK, response.Code)
	response = organizationRequest(organizationHandler(&fakeOrganizationService{defaultErr: errors.New("failed")}).getDefault, http.MethodGet, "/", "")
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestHandlerUpdateRejectsInvalidJSONAndValidation(t *testing.T) {
	h := organizationHandler(&fakeOrganizationService{})
	assert.Equal(t, http.StatusBadRequest, organizationIDRequest(h.update, http.MethodPut, idForTest(), `{bad`).Code)
	assert.Equal(t, http.StatusUnprocessableEntity, organizationIDRequest(h.update, http.MethodPut, idForTest(), `{}`).Code)
}

func idForTest() string { return uuid.New().String() }

var _ organizationService = (*fakeOrganizationService)(nil)
