package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}

type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

var errTrailingData = errors.New("unexpected data after JSON object")

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{clock: clock, tasks: make(map[string]Task)}
}

func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, ErrInvalidTitle
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	t := Task{
		ID:        strconv.FormatUint(r.seq, 10),
		Title:     title,
		UpdatedAt: r.clock.Now(),
	}
	r.tasks[t.ID] = t
	return t, nil
}

func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.tasks[id]
	return t, ok
}

func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		if t.Done == done {
			out = append(out, t)
		}
	}
	return out
}

func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	if t.Done != done {
		t.Done = done
		t.UpdatedAt = r.clock.Now()
		r.tasks[id] = t
	}
	return t, nil
}

type httpHandler struct {
	repo TaskRepo
	mux  *http.ServeMux
}

func NewHTTPHandler(repo TaskRepo) http.Handler {
	h := &httpHandler{repo: repo, mux: http.NewServeMux()}
	h.mux.HandleFunc("POST /tasks", h.handleCreate)
	h.mux.HandleFunc("GET /tasks", h.handleList)
	h.mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		h.handleGet(w, r, r.PathValue("id"))
	})
	h.mux.HandleFunc("PATCH /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		h.handlePatch(w, r, r.PathValue("id"))
	})
	return h
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title *string `json:"title"`
	}
	if err := decodeStrictJSON(r.Body, &req); err != nil || req.Title == nil {
		http.Error(w, "body must be {\"title\": string}", http.StatusBadRequest)
		return
	}

	t, err := h.repo.Create(*req.Title)
	switch {
	case errors.Is(err, ErrInvalidTitle):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case err != nil:
		http.Error(w, "could not create task", http.StatusInternalServerError)
	default:
		writeJSON(w, http.StatusCreated, t)
	}
}

func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	t, ok := h.repo.Get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	done := false
	if vals, given := r.URL.Query()["done"]; given {
		if len(vals) != 1 || (vals[0] != "true" && vals[0] != "false") {
			http.Error(w, "done must be true or false", http.StatusBadRequest)
			return
		}
		done = vals[0] == "true"
	}

	tasks := slices.Clone(h.repo.List(done))
	if tasks == nil {
		tasks = []Task{}
	}
	slices.SortFunc(tasks, func(a, b Task) int {
		if c := b.UpdatedAt.Compare(a.UpdatedAt); c != 0 {
			return c
		}
		return strings.Compare(a.ID, b.ID)
	})
	writeJSON(w, http.StatusOK, tasks)
}

func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var req struct {
		Done *bool `json:"done"`
	}
	if err := decodeStrictJSON(r.Body, &req); err != nil || req.Done == nil {
		http.Error(w, "body must be {\"done\": bool}", http.StatusBadRequest)
		return
	}

	t, err := h.repo.SetDone(id, *req.Done)
	switch {
	case errors.Is(err, ErrNotFound):
		http.NotFound(w, r)
	case err != nil:
		http.Error(w, "could not update task", http.StatusInternalServerError)
	default:
		writeJSON(w, http.StatusOK, t)
	}
}

func decodeStrictJSON(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return err
	}
	if _, err := dec.Token(); err != io.EOF {
		return errTrailingData
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
