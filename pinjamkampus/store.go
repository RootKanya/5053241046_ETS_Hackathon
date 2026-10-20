package main

import (
	"fmt"

	"sort"
	"strings"
	"sync"
	"time"
)

type SumberDaya struct {
	ID    int    `json:"id"`
	Nama  string `json:"nama"`
	Jenis string `json:"jenis"`
	Stok  int    `json:"stok"`
}

type Peminjaman struct {
	ID           int    `json:"id"`
	SumberDayaID int    `json:"sumber_daya_id"`
	Peminjam     string `json:"peminjam"`
	Tanggal      string `json:"tanggal"`
	SlotMulai    int    `json:"slot_mulai"`
	SlotSelesai  int    `json:"slot_selesai"`
	Jumlah       int    `json:"jumlah"`
	Status       string `json:"status"`
}

type KuotaResponse struct {
	Peminjam string `json:"peminjam"`
	Tanggal  string `json:"tanggal"`
	Terpakai int    `json:"terpakai"`
	Maksimal int    `json:"maksimal"`
	Sisa     int    `json:"sisa"`
}

type Store struct {
	mu             sync.Mutex
	sumberDayaMap  map[int]SumberDaya
	peminjamanMap  map[int]Peminjaman
	nextSumberID   int
	nextPeminjamID int
}

func NewStore() *Store {
	return &Store{
		sumberDayaMap:  make(map[int]SumberDaya),
		peminjamanMap:  make(map[int]Peminjaman),
		nextSumberID:   1,
		nextPeminjamID: 1,
	}
}

// --- Sumber Daya ---

func (s *Store) CreateSumberDaya(sd SumberDaya) SumberDaya {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd.ID = s.nextSumberID
	s.nextSumberID++
	s.sumberDayaMap[sd.ID] = sd
	return sd
}

func (s *Store) GetAllSumberDaya() []SumberDaya {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := make([]SumberDaya, 0, len(s.sumberDayaMap))
	for _, v := range s.sumberDayaMap {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res
}

func (s *Store) GetSumberDayaByID(id int) (SumberDaya, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sd, exists := s.sumberDayaMap[id]
	return sd, exists
}

// --- Peminjaman Helper & Business Rules ---

func isValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func (s *Store) GetUserQuota(peminjam string, tanggal string) (terpakai int) {
	// Peminjam dipastikan sudah di-trim di luar
	trimmedUser := strings.TrimSpace(peminjam)
	for _, p := range s.peminjamanMap {
		if p.Status == "aktif" && p.Tanggal == tanggal && p.Peminjam == trimmedUser {
			terpakai++
		}
	}
	return terpakai
}

func (s *Store) CheckConflict(sumberDayaID int, tanggal string, slotMulai, slotSelesai, jumlah int, excludeID int) bool {
	sd, exists := s.sumberDayaMap[sumberDayaID]
	if !exists {
		return true
	}

	for slot := slotMulai; slot <= slotSelesai; slot++ {
		usedUnits := 0
		for _, p := range s.peminjamanMap {
			if p.ID == excludeID {
				continue
			}
			if p.Status == "aktif" && p.SumberDayaID == sumberDayaID && p.Tanggal == tanggal {
				if slot >= p.SlotMulai && slot <= p.SlotSelesai {
					usedUnits += p.Jumlah
				}
			}
		}
		if usedUnits+jumlah > sd.Stok {
			return true 
		}
	}
	return false
}

// --- Peminjaman Operations ---

func (s *Store) CreatePeminjaman(p Peminjaman) (Peminjaman, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	terpakai := s.GetUserQuota(p.Peminjam, p.Tanggal)
	if terpakai >= 2 {
		return Peminjaman{}, 409, fmt.Errorf("kuota peminjaman per hari sudah penuh (maksimal 2)")
	}

	if s.CheckConflict(p.SumberDayaID, p.Tanggal, p.SlotMulai, p.SlotSelesai, p.Jumlah, 0) {
		return Peminjaman{}, 409, fmt.Errorf("kapasitas stok pada slot tersebut tidak mencukupi")
	}

	p.ID = s.nextPeminjamID
	s.nextPeminjamID++
	p.Status = "aktif"
	s.peminjamanMap[p.ID] = p
	return p, 201, nil
}

func (s *Store) GetAllPeminjaman(sumberDayaIDFilter *int, tanggalFilter string, statusFilter string) []Peminjaman {
	s.mu.Lock()
	defer s.mu.Unlock()

	res := make([]Peminjaman, 0)
	for _, p := range s.peminjamanMap {
		if sumberDayaIDFilter != nil && p.SumberDayaID != *sumberDayaIDFilter {
			continue
		}
		if tanggalFilter != "" && p.Tanggal != tanggalFilter {
			continue
		}
		if statusFilter != "" && p.Status != statusFilter {
			continue
		}
		res = append(res, p)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res
}

func (s *Store) GetPeminjamanByID(id int) (Peminjaman, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, exists := s.peminjamanMap[id]
	return p, exists
}

func (s *Store) UpdatePeminjaman(id int, p Peminjaman) (Peminjaman, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldData, exists := s.peminjamanMap[id]
	if !exists {
		return Peminjaman{}, 404, fmt.Errorf("peminjaman tidak ditemukan")
	}

	terpakai := 0
	for _, item := range s.peminjamanMap {
		if item.ID == id {
			continue 
		}
		if item.Status == "aktif" && item.Tanggal == p.Tanggal && item.Peminjam == p.Peminjam {
			terpakai++
		}
	}
	if terpakai >= 2 {
		return Peminjaman{}, 409, fmt.Errorf("kuota peminjaman peminjam target pada tanggal tersebut sudah penuh")
	}

	if s.CheckConflict(p.SumberDayaID, p.Tanggal, p.SlotMulai, p.SlotSelesai, p.Jumlah, id) {
		return Peminjaman{}, 409, fmt.Errorf("kapasitas stok pada slot tersebut tidak mencukupi")
	}

	p.ID = id
	p.Status = oldData.Status
	s.peminjamanMap[id] = p
	return p, 200, nil
}

func (s *Store) DeletePeminjaman(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.peminjamanMap[id]; !exists {
		return false
	}
	delete(s.peminjamanMap, id)
	return true
}

func (s *Store) GetQuotaInfo(peminjam string, tanggal string) (KuotaResponse, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !isValidDate(tanggal) {
		return KuotaResponse{}, 400, fmt.Errorf("tanggal tidak valid")
	}

	terpakai := s.GetUserQuota(peminjam, tanggal)
	sisa := 2 - terpakai
	if sisa < 0 {
		sisa = 0
	}

	return KuotaResponse{
		Peminjam: peminjam,
		Tanggal:  tanggal,
		Terpakai: terpakai,
		Maksimal: 2,
		Sisa:     sisa,
	}, 200, nil
}