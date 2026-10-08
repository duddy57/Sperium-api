package machines

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

type machineFakeService struct {
	createID    uuid.UUID
	createToken string
	createErr   error
	details     Machine
	detailsErr  error
	list        []Machine
	listErr     error
	updateErr   error
	deleteErr   error
	activateErr error
}

func (f *machineFakeService) Create(_ context.Context, _ CreateMachineReq, _ uuid.UUID) (uuid.UUID, string, error) {
	return f.createID, f.createToken, f.createErr
}
func (f *machineFakeService) Details(_ context.Context, _ uuid.UUID) (Machine, error) {
	return f.details, f.detailsErr
}
func (f *machineFakeService) Delete(_ context.Context, _ uuid.UUID) error { return f.deleteErr }
func (f *machineFakeService) List(_ context.Context, _ uuid.UUID) ([]Machine, error) {
	return f.list, f.listErr
}
func (f *machineFakeService) Update(_ context.Context, _ uuid.UUID, _ UpdateMachineReq) error {
	return f.updateErr
}
func (f *machineFakeService) Activate(_ context.Context, _ uuid.UUID, _ ActivateMachineReq) error {
	return f.activateErr
}

func machineTestHandler(service *machineFakeService) *Handler {
	return NewMachineHandler(service, validator.New())
}

func machineRequest(handler http.HandlerFunc, method, body string, params map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/", strings.NewReader(body))
	route := chi.NewRouteContext()
	for key, value := range params {
		route.URLParams.Add(key, value)
	}
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, route))
	response := httptest.NewRecorder()
	handler(response, req)
	return response
}

func TestHandlerCreateSuccess(t *testing.T) {
	orgID, machineID := uuid.New(), uuid.New()
	response := machineRequest(machineTestHandler(&machineFakeService{
		createID: machineID, createToken: "token",
	}).create, http.MethodPost, `{"name":"server-1"}`, map[string]string{"organization_id": orgID.String()})

	assert.Equal(t, http.StatusCreated, response.Code)
	var body createMachineResp
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	assert.Equal(t, machineID, body.ID)
	assert.Equal(t, "token", body.Token)
}

func TestHandlerCreateInvalidJSONAndUUID(t *testing.T) {
	validOrg := uuid.New().String()
	response := machineRequest(machineTestHandler(&machineFakeService{}).create, http.MethodPost, `{`, map[string]string{"organization_id": validOrg})
	assert.Equal(t, http.StatusBadRequest, response.Code)

	response = machineRequest(machineTestHandler(&machineFakeService{}).create, http.MethodPost, `{"name":"server-1"}`, map[string]string{"organization_id": "bad"})
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestHandlerCreateValidationAndServiceError(t *testing.T) {
	response := machineRequest(machineTestHandler(&machineFakeService{}).create, http.MethodPost, `{}`, map[string]string{"organization_id": uuid.New().String()})
	assert.Equal(t, http.StatusBadRequest, response.Code)

	response = machineRequest(machineTestHandler(&machineFakeService{createErr: errors.New("create failed")}).create, http.MethodPost, `{"name":"server-1"}`, map[string]string{"organization_id": uuid.New().String()})
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestHandlerDetailsListUpdateDeleteAndActivate(t *testing.T) {
	machineID, orgID := uuid.New(), uuid.New()
	handler := machineTestHandler(&machineFakeService{
		details: Machine{ID: machineID, OrganizationID: orgID, Name: "server-1", CreatedAt: time.Now()},
		list:    []Machine{{ID: machineID, OrganizationID: orgID}},
	})

	response := machineRequest(handler.get, http.MethodGet, "", map[string]string{"machine_id": machineID.String()})
	assert.Equal(t, http.StatusOK, response.Code)
	response = machineRequest(handler.list, http.MethodGet, "", map[string]string{"organization_id": orgID.String()})
	assert.Equal(t, http.StatusOK, response.Code)
	response = machineRequest(handler.update, http.MethodPut, `{"name":"server-2"}`, map[string]string{"machine_id": machineID.String()})
	assert.Equal(t, http.StatusNoContent, response.Code)
	response = machineRequest(handler.delete, http.MethodDelete, "", map[string]string{"machine_id": machineID.String()})
	assert.Equal(t, http.StatusNoContent, response.Code)
	response = machineRequest(handler.activate, http.MethodPut, `{"token":"token","hostname":"host","agent_version":"v1"}`, map[string]string{"machine_id": machineID.String()})
	assert.Equal(t, http.StatusCreated, response.Code)
}

func TestHandlerServiceErrorsAndInvalidUUID(t *testing.T) {
	handler := machineTestHandler(&machineFakeService{
		detailsErr: errors.New("details"), listErr: errors.New("list"), updateErr: errors.New("update"),
		deleteErr: errors.New("delete"), activateErr: errors.New("activate"),
	})
	assert.Equal(t, http.StatusBadRequest, machineRequest(handler.get, http.MethodGet, "", map[string]string{"machine_id": "bad"}).Code)
	assert.Equal(t, http.StatusNotFound, machineRequest(handler.get, http.MethodGet, "", map[string]string{"machine_id": uuid.New().String()}).Code)
	assert.Equal(t, http.StatusInternalServerError, machineRequest(handler.list, http.MethodGet, "", map[string]string{"organization_id": uuid.New().String()}).Code)
	assert.Equal(t, http.StatusInternalServerError, machineRequest(handler.update, http.MethodPut, `{"name":"server-2"}`, map[string]string{"machine_id": uuid.New().String()}).Code)
	assert.Equal(t, http.StatusInternalServerError, machineRequest(handler.delete, http.MethodDelete, "", map[string]string{"machine_id": uuid.New().String()}).Code)
	assert.Equal(t, http.StatusInternalServerError, machineRequest(handler.activate, http.MethodPut, `{"token":"token","hostname":"host","agent_version":"v1"}`, map[string]string{"machine_id": uuid.New().String()}).Code)
}
