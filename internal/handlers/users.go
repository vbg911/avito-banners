package handlers

import (
	"avito-backend-assignment/internal/storage"
	"database/sql"
	"github.com/bradfitz/gomemcache/memcache"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type UsersHandler struct {
	Logger *zap.SugaredLogger
	DB     *sql.DB
	MC     *memcache.Client
}

func (h *UsersHandler) Banner(w http.ResponseWriter, r *http.Request) {
	tagId, err := strconv.Atoi(r.URL.Query().Get("tag_id"))
	if err != nil {
		http.Error(w, "invalid tag_id", http.StatusBadRequest)
		return
	}
	featureId, err := strconv.Atoi(r.URL.Query().Get("feature_id"))
	if err != nil {
		http.Error(w, "invalid feature_id", http.StatusBadRequest)
		return
	}
	var lastRevision bool
	useLastRevision := r.URL.Query().Get("use_last_revision")
	if useLastRevision != "" {
		lastRevision, err = strconv.ParseBool(r.URL.Query().Get("use_last_revision"))
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error": "invalid use_last_revision"}`, http.StatusBadRequest)
			return
		}
	}

	key := strconv.Itoa(tagId) + ":" + strconv.Itoa(featureId)
	// идем в кеш
	if !lastRevision {
		item, err := h.MC.Get(key)
		if err != nil {
			if err == memcache.ErrCacheMiss {
				h.Logger.Info("Данные с ключом ", key, " не найдены в Memcached")
			}
			h.Logger.Infof("Ошибка при чтении данных из Memcached: %v", err)
		} else {
			h.Logger.Info("ответ из memcached")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(item.Value)
			if err != nil {
				h.Logger.Error(err.Error())
				http.Error(w, `{"error": "cant send response"}`, http.StatusInternalServerError)
				return
			}
			return
		}

	}

	jsonResponse, err := storage.GetBanners(h.DB, tagId, featureId)
	if err != nil {
		h.Logger.Error(err.Error())
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error": "cant get banner from storage"}`, http.StatusInternalServerError)
		return
	}

	if jsonResponse == nil {
		h.Logger.Info("banner not founded")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	//кешируем ответ на 5 минут
	err = h.MC.Set(&memcache.Item{Key: key, Value: jsonResponse, Expiration: 300})
	if err != nil {
		h.Logger.Errorf("Ошибка при записи данных в Memcached: %v", err)
	}
	h.Logger.Info("Данные успешно записаны в Memcached")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, `{"error": "cant send response"}`, http.StatusInternalServerError)
		return
	}
}
