package main

import (
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Candidate struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Age       int      `json:"age"`
	Interests []string `json:"interests"`
}

type SwipeRequest struct {
	CandidateID string `json:"candidate_id"`
	Action      string `json:"action"` // "like" or "pass"
}
type SwipeResponse struct {
	Matched bool   `json:"matched"`
	Message string `json:"message"`
}

var sampleInterests = []string{
	"climbing", "hiking", "bouldering", "music", "reading",
	"cooking", "running", "yoga", "gaming", "travel",
}

func randName() string {
	first := []string{"Alex", "Sam", "Charlie", "Noa", "Léa", "Inès", "Eli", "Robin", "Nora", "Maya"}
	last := []string{"Martin", "Bernard", "Petit", "Robert", "Richard", "Durand", "Dubois"}
	return first[rand.IntN(len(first))] + " " + last[rand.IntN(len(last))]
}

func randInterests() []string {
	n := 1 + rand.IntN(4)
	m := map[string]struct{}{}
	var out []string
	for len(out) < n {
		i := sampleInterests[rand.IntN(len(sampleInterests))]
		if _, ok := m[i]; !ok {
			m[i] = struct{}{}
			out = append(out, i)
		}
	}
	return out
}

func handleCandidates(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	results := make([]Candidate, 0, limit)
	for i := 0; i < limit; i++ {
		results = append(results, Candidate{
			ID:        strconv.FormatInt(time.Now().UnixNano()+int64(i), 36),
			Name:      randName(),
			Age:       20 + rand.IntN(20),
			Interests: randInterests(),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func handleSwipe(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req SwipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if req.Action != "like" && req.Action != "pass" {
		http.Error(w, "action must be like or pass", http.StatusBadRequest)
		return
	}
	matched := req.Action == "like" && rand.Float64() < 0.18
	msg := "ok"
	if matched {
		msg = "it's a match!"
	}
	writeJSON(w, http.StatusOK, SwipeResponse{Matched: matched, Message: msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	mux := http.NewServeMux()
	mux.HandleFunc("/candidates", withAuth(handleCandidates))
	mux.HandleFunc("/swipe", withAuth(handleSwipe))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })

	addr := ":8080"
	log.Printf("Mock server listening on %s (endpoints: /candidates, /swipe)\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "" && !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "invalid auth header", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
