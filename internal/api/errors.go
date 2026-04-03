package api

import (
	"errors"
	"net/http"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func writeLookupError(writer http.ResponseWriter, err error, notFoundMessage, internalMessage string) {
	if errors.Is(err, domain.ErrNotFound) {
		http.Error(writer, notFoundMessage, http.StatusNotFound)
		return
	}
	http.Error(writer, internalMessage, http.StatusInternalServerError)
}
