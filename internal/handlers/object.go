package handlers

import (
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/ohmayank/hako/internal/store"
)

func PutObject(s store.ObjectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.PathValue("key")

		if err := s.Put(r.Context(), bucket, key, r.Body); err != nil {
			if errors.Is(err, store.ErrEmptyBucket) ||
				errors.Is(err, store.ErrEmptyKey) ||
				errors.Is(err, store.ErrWrongPath) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

	}
}

func GetObject(s store.ObjectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.PathValue("key")

		rc, err := s.Get(r.Context(), bucket, key)
		if err != nil {
			if errors.Is(err, store.ErrEmptyBucket) ||
				errors.Is(err, store.ErrEmptyKey) ||
				errors.Is(err, store.ErrWrongPath) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if errors.Is(err, os.ErrNotExist) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rc.Close()

		if _, err := io.Copy(w, rc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func DeleteObject(s store.ObjectStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.PathValue("key")

		if err := s.Delete(r.Context(), bucket, key); err != nil {
			if errors.Is(err, store.ErrEmptyBucket) ||
				errors.Is(err, store.ErrEmptyKey) ||
				errors.Is(err, store.ErrWrongPath) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	}
}
