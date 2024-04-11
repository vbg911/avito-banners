package handlers

import (
	"avito-backend-assignment/internal/storage"
	"database/sql"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type UsersHandler struct {
	Logger *zap.SugaredLogger
	DB     *sql.DB
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
			//todo при ошибке возвращать json
			http.Error(w, "invalid use_last_revision", http.StatusBadRequest)
			return
		}
	}

	//todo проверка токена на валидность
	token := r.Header.Get("token")

	h.Logger.Infof("tag_id: %v, feture_id: %v, use_last_revision: %v, token: %v", tagId, featureId, lastRevision, token)

	//todo если можно не последнюю ревизию идем в кеш

	jsonResponse, err := storage.GetBanners(h.DB, tagId, featureId)
	if err != nil {
		h.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if jsonResponse == nil {
		h.Logger.Error("banner not founded")
		w.WriteHeader(http.StatusNotFound)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		h.Logger.Error("W.write err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
