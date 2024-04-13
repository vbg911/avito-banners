package handlers

import (
	"avito-backend-assignment/internal/storage"
	"database/sql"
	"encoding/json"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type AdminsHandler struct {
	Logger *zap.SugaredLogger
	DB     *sql.DB
	MC     *memcache.Client
}

type requestBody struct {
	TagIds    []int                  `json:"tag_ids"`
	FeatureId int                    `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
}

type createResponse struct {
	BannerID int `json:"banner_id"`
}

type banner struct {
	BannerID  int                    `json:"banner_id"`
	TagIds    []int                  `json:"tag_ids"`
	FeatureId int                    `json:"feature_id"`
	Content   map[string]interface{} `json:"content"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

func (h *AdminsHandler) NewBanner(w http.ResponseWriter, r *http.Request) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Info(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "bad body"}`, http.StatusBadRequest)
	}
	var body requestBody
	err = json.Unmarshal(reqBody, &body)
	if err != nil {
		h.Logger.Info(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "bad body"}`, http.StatusBadRequest)
	}

	bannerId, err := storage.NewBanner(h.DB, body.TagIds, body.FeatureId, body.Content, body.IsActive)
	if err != nil {
		h.Logger.Error(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "cant add new banner in storage"}`, http.StatusInternalServerError)
	}

	response := createResponse{BannerID: bannerId}
	res, err := json.Marshal(response)
	if err != nil {
		h.Logger.Error(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "cant create response"}`, http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(res)
}

func (h *AdminsHandler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.Logger.Info(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "bad id"}`, http.StatusBadRequest)
		return
	}

	rows, err := storage.DeleteBanner(h.DB, id)
	if err != nil {
		h.Logger.Error(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "cant delete banner from storage"}`, http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminsHandler) UpdateBanner(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		h.Logger.Info(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "bad id"}`, http.StatusBadRequest)
		return
	}

	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Info(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "bad body"}`, http.StatusBadRequest)
	}
	var body requestBody
	err = json.Unmarshal(reqBody, &body)
	if err != nil {
		h.Logger.Info(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "bad body"}`, http.StatusBadRequest)
	}

	rows, err := storage.UpdateBanner(h.DB, id, body.TagIds, body.FeatureId, body.Content, body.IsActive)
	if err != nil {
		h.Logger.Error(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "cant update banner in storage"}`, http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *AdminsHandler) Banner(w http.ResponseWriter, r *http.Request) {
	featureId := -1
	fId := r.URL.Query().Get("feature_id")
	if len(fId) != 0 {
		id, err := strconv.Atoi(fId)
		if err != nil {
			h.Logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		featureId = id
	}

	tagId := -1
	tId := r.URL.Query().Get("tag_id")
	if len(tId) != 0 {
		tagID, err := strconv.Atoi(tId)
		if err != nil {
			h.Logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		tagId = tagID
	}

	lim := r.URL.Query().Get("limit")
	limit, err := strconv.Atoi(lim)
	if err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	off := r.URL.Query().Get("offset")
	offset, err := strconv.Atoi(off)
	if err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, err := storage.GetBannersAdmin(h.DB, tagId, featureId, limit, offset)
	if err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var banners []*banner
	for rows.Next() {
		b := new(banner)
		var tags, content []uint8
		if err := rows.Scan(&b.BannerID, &tags, &b.FeatureId, &content, &b.IsActive, &b.CreatedAt, &b.UpdatedAt); err != nil {
			h.Logger.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err := json.Unmarshal(content, &b.Content)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error": "cant parse banner content"}`, http.StatusInternalServerError)
			return
		}

		str := strings.Trim(string(tags), "{}")
		parts := strings.Split(str, ",")
		var numbers []int
		for _, s := range parts {
			n, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error": "cant parse banner"}`, http.StatusInternalServerError)
				return
			}
			numbers = append(numbers, n)
		}
		b.TagIds = numbers
		banners = append(banners, b)
	}

	if err := rows.Err(); err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rows.Close()

	// Преобразование результатов в JSON
	jsonData, err := json.Marshal(banners)
	if err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
