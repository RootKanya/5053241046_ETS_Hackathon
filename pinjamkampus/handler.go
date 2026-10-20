package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	store *Store
}

func NewHandler() http.Handler {
	h := &Handler{store: NewStore()}
	mux := http.NewServeMux()

	// Endpoints Sumber Daya
	mux.HandleFunc("POST /sumber-daya", h.handleCreateSumberDaya)
	mux.HandleFunc("GET /sumber-daya", h.handleGetAllSumberDaya)
	mux.HandleFunc("GET /sumber-daya/{id}", h.handleGetSumberDayaByID)

	// Endpoints Peminjaman
	mux.HandleFunc("POST /peminjaman", h.handleCreatePeminjaman)
	mux.HandleFunc("GET /peminjaman", h.handleGetAllPeminjaman)
	mux.HandleFunc("GET /peminjaman/{id}", h.handleGetPeminjamanByID)
	mux.HandleFunc("PUT /peminjaman/{id}", h.handleUpdatePeminjaman)
	mux.HandleFunc("DELETE /peminjaman/{id}", h.handleDeletePeminjaman)

	// Endpoint Fitur Khas Persona B (Cek Quota)
	mux.HandleFunc("GET /peminjam/{nama}/kuota", h.handleGetKuota)
	mux.HandleFunc("GET /ringkasan", h.handleGetRingkasan)
	
	return mux
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// --- Handler Sumber Daya ---

func (h *Handler) handleCreateSumberDaya(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Nama  *string `json:"nama"`
		Jenis *string `json:"jenis"`
		Stok  *int    `json:"stok"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, 400, "Request body JSON tidak valid")
		return
	}

	if body.Nama == nil || body.Jenis == nil || body.Stok == nil {
		writeError(w, 400, "Field nama, jenis, dan stok wajib ada")
		return
	}

	nama := strings.TrimSpace(*body.Nama)
	jenis := strings.TrimSpace(*body.Jenis)
	stok := *body.Stok

	if nama == "" || len([]rune(nama)) > 50 {
		writeError(w, 400, "Nama wajib diisi dan maksimal 50 karakter")
		return
	}
	if jenis != "ruang" && jenis != "alat" {
		writeError(w, 400, "Jenis harus berupa 'ruang' atau 'alat'")
		return
	}
	if stok < 1 {
		writeError(w, 400, "Stok minimal bernilai 1")
		return
	}

	created := h.store.CreateSumberDaya(SumberDaya{
		Nama:  nama,
		Jenis: jenis,
		Stok:  stok,
	})

	writeJSON(w, 201, created)
}

func (h *Handler) handleGetAllSumberDaya(w http.ResponseWriter, r *http.Request) {
	list := h.store.GetAllSumberDaya()
	writeJSON(w, 200, list)
}

func (h *Handler) handleGetSumberDayaByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, 400, "ID harus berupa angka integer")
		return
	}

	sd, exists := h.store.GetSumberDayaByID(id)
	if !exists {
		writeError(w, 404, "Sumber daya tidak ditemukan")
		return
	}

	writeJSON(w, 200, sd)
}

// --- Handler Peminjaman ---

func (h *Handler) parsePeminjamanBody(r *http.Request) (Peminjaman, string, bool) {
	var body struct {
		SumberDayaID *int    `json:"sumber_daya_id"`
		Peminjam     *string `json:"peminjam"`
		Tanggal      *string `json:"tanggal"`
		SlotMulai    *int    `json:"slot_mulai"`
		SlotSelesai  *int    `json:"slot_selesai"`
		Jumlah       *int    `json:"jumlah"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return Peminjaman{}, "Body JSON tidak valid", false
	}

	if body.SumberDayaID == nil || body.Peminjam == nil || body.Tanggal == nil ||
		body.SlotMulai == nil || body.SlotSelesai == nil || body.Jumlah == nil {
		return Peminjaman{}, "Semua 6 field peminjaman wajib diisi", false
	}

	peminjam := strings.TrimSpace(*body.Peminjam)
	tanggal := strings.TrimSpace(*body.Tanggal)

	if peminjam == "" || len([]rune(peminjam)) > 50 {
		return Peminjaman{}, "Peminjam wajib diisi dan maksimal 50 karakter", false
	}
	if !isValidDate(tanggal) {
		return Peminjaman{}, "Tanggal tidak valid (format YYYY-MM-DD)", false
	}
	if *body.SlotMulai < 1 || *body.SlotMulai > 8 || *body.SlotSelesai < 1 || *body.SlotSelesai > 8 {
		return Peminjaman{}, "Slot harus bernilai antara 1 sampai 8", false
	}
	if *body.SlotMulai > *body.SlotSelesai {
		return Peminjaman{}, "slot_mulai harus lebih kecil atau sama dengan slot_selesai", false
	}

	// Persona B Rule 1: Max 3 slots duration
	duration := *body.SlotSelesai - *body.SlotMulai + 1
	if duration > 3 {
		return Peminjaman{}, "Durasi peminjaman maksimal 3 slot", false
	}

	sd, exists := h.store.GetSumberDayaByID(*body.SumberDayaID)
	if !exists {
		return Peminjaman{}, "Sumber daya tidak ditemukan", false
	}

	if *body.Jumlah < 1 || *body.Jumlah > sd.Stok {
		return Peminjaman{}, "Jumlah peminjaman harus antara 1 dan stok sumber daya", false
	}

	return Peminjaman{
		SumberDayaID: *body.SumberDayaID,
		Peminjam:     peminjam,
		Tanggal:      tanggal,
		SlotMulai:    *body.SlotMulai,
		SlotSelesai:  *body.SlotSelesai,
		Jumlah:       *body.Jumlah,
	}, "", true
}

func (h *Handler) handleCreatePeminjaman(w http.ResponseWriter, r *http.Request) {
	p, errMsg, ok := h.parsePeminjamanBody(r)
	if !ok {
		writeError(w, 400, errMsg)
		return
	}

	created, statusCode, err := h.store.CreatePeminjaman(p)
	if err != nil {
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, 201, created)
}

func (h *Handler) handleGetAllPeminjaman(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	var sumberDayaIDPtr *int
	if sdStr := query.Get("sumber_daya_id"); sdStr != "" {
		sdID, err := strconv.Atoi(sdStr)
		if err != nil {
			writeError(w, 400, "Filter sumber_daya_id harus berupa angka integer")
			return
		}
		sumberDayaIDPtr = &sdID
	}

	tanggalFilter := query.Get("tanggal")
	statusFilter := query.Get("status")

	list := h.store.GetAllPeminjaman(sumberDayaIDPtr, tanggalFilter, statusFilter)
	writeJSON(w, 200, list)
}

func (h *Handler) handleGetPeminjamanByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, 400, "ID harus berupa angka integer")
		return
	}

	p, exists := h.store.GetPeminjamanByID(id)
	if !exists {
		writeError(w, 404, "Peminjaman tidak ditemukan")
		return
	}

	writeJSON(w, 200, p)
}

func (h *Handler) handleUpdatePeminjaman(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, 400, "ID harus berupa angka integer")
		return
	}

	if _, exists := h.store.GetPeminjamanByID(id); !exists {
		writeError(w, 404, "Peminjaman tidak ditemukan")
		return
	}

	p, errMsg, ok := h.parsePeminjamanBody(r)
	if !ok {
		writeError(w, 400, errMsg)
		return
	}

	updated, statusCode, err := h.store.UpdatePeminjaman(id, p)
	if err != nil {
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, 200, updated)
}

func (h *Handler) handleDeletePeminjaman(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, 400, "ID harus berupa angka integer")
		return
	}

	success := h.store.DeletePeminjaman(id)
	if !success {
		writeError(w, 404, "Peminjaman tidak ditemukan")
		return
	}

	w.WriteHeader(204)
}

// --- Handler Persona B (Kuota) ---

func (h *Handler) handleGetKuota(w http.ResponseWriter, r *http.Request) {
	namaParam := r.PathValue("nama")
	peminjam := strings.TrimSpace(namaParam)

	tanggal := r.URL.Query().Get("tanggal")
	if tanggal == "" {
		writeError(w, 400, "Query parameter tanggal wajib diisi")
		return
	}

	res, statusCode, err := h.store.GetQuotaInfo(peminjam, tanggal)
	if err != nil {
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, 200, res)
}

// --- Handler Change Request A ---

func (h *Handler) handleGetRingkasan(w http.ResponseWriter, r *http.Request) {
	tanggal := r.URL.Query().Get("tanggal")
	if tanggal == "" {
		writeError(w, 400, "Query parameter tanggal wajib diisi")
		return
	}

	res, statusCode, err := h.store.GetDashboardSummary(tanggal)
	if err != nil {
		writeError(w, statusCode, err.Error())
		return
	}

	writeJSON(w, 200, res)
}