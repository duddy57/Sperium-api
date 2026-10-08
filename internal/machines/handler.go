package machines

import (
	"encoding/json"
	"net/http"
	"uuid"

	"github.com/duddy57/sperium/internal/plataform/json_utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service  machineService
	validate *validator.Validate
}

func NewMachineHandler(service machineService, validate *validator.Validate) *Handler {
	return &Handler{service: service, validate: validate}
}

func NewOrganizationHandler(service machineService, validate *validator.Validate) *Handler {
	return NewMachineHandler(service, validate)
}

var _ machineService = (*Service)(nil)

func (h *Handler) RegisterRoutes(r chi.Router, auth func(http.Handler) http.Handler) {
	r.Route("/{organization_id}", func(r chi.Router) {
		r.Use(auth)
		r.Post("/create", h.create)
		r.Get("/details/{machine_id}", h.get)
		r.Put("/update/{machine_id}", h.update)
		r.Delete("/delete/{machine_id}", h.delete)
		r.Get("/list", h.list)
		r.Put("/activate/{machine_id}", h.activate)
	})
}

func parseUUID(w http.ResponseWriter, value string) (uuid.UUID, bool) {
	id, err := uuid.Parse(value)
	if err != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid id")
		return uuid.Nil(), false
	}
	return id, true
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := parseUUID(w, chi.URLParam(r, "organization_id"))
	if !ok {
		return
	}
	var req CreateMachineReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || h.validate.Struct(req) != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	id, token, err := h.service.Create(r.Context(), req, organizationID)
	if err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}
	json_utils.WriteJSON(w, http.StatusCreated, createMachineResp{Message: "Machine created successfully!", ID: id, Token: token})
}
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, chi.URLParam(r, "machine_id"))
	if !ok {
		return
	}
	machine, err := h.service.Details(r.Context(), id)
	if err != nil {
		json_utils.WriteError(w, http.StatusNotFound, "machine not found")
		return
	}
	json_utils.WriteJSON(w, http.StatusOK, getMachineResp{
		Message: "Machine retrieved successfully!",
		Machine: machine})
}
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := parseUUID(w, chi.URLParam(r, "organization_id"))
	if !ok {
		return
	}
	machines, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}
	json_utils.WriteJSON(w, http.StatusOK, listMachineResp{
		Message: "Machine listed successfully!",
		Machine: machines,
	})
}
func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, chi.URLParam(r, "machine_id"))
	if !ok {
		return
	}
	var req UpdateMachineReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || h.validate.Struct(req) != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := h.service.Update(r.Context(), id, req); err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, chi.URLParam(r, "machine_id"))
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) activate(w http.ResponseWriter, r *http.Request) {
	machineID, ok := parseUUID(w, chi.URLParam(r, "machine_id"))
	if !ok {
		return
	}
	var req ActivateMachineReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || h.validate.Struct(req) != nil {
		json_utils.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := h.service.Activate(r.Context(), machineID, req); err != nil {
		json_utils.WriteError(w, http.StatusInternalServerError, "try again later")
		return
	}
	json_utils.WriteJSON(w, http.StatusCreated, genericResp{Message: "Machine activate successfully!"})
}
