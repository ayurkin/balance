package http

import (
	e "balance/internal/domain/errors"
	"balance/internal/domain/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func (s *Server) balanceHandlers() http.Handler {
	h := chi.NewMux()
	h.Route("/", func(r chi.Router) {
		h.Post("/income", s.addIncome)
		h.Post("/expense", s.addExpense)
		h.Post("/transfer", s.doTransfer)
		h.Get("/balance", s.getBalance)
	})
	return h
}

func (s *Server) addIncome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}
	incomeParams := &models.BalanceWithDesc{}
	err = json.Unmarshal(body, incomeParams)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}
	incomeParams.Time = time.Now()

	err = s.balance.AddIncome(r.Context(), *incomeParams)

	if err != nil {
		if errors.Is(err, e.DatabaseError) {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"success\"}"))

}

func (s *Server) addExpense(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}
	incomeParams := &models.BalanceWithDesc{}
	err = json.Unmarshal(body, incomeParams)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}
	incomeParams.Time = time.Now()

	err = s.balance.AddExpense(r.Context(), *incomeParams)

	if err != nil {
		if errors.Is(err, e.DatabaseError) {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"success\"}"))
}

func (s *Server) doTransfer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}
	transferParams := &models.Transaction{}
	err = json.Unmarshal(body, transferParams)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}
	transferParams.Time = time.Now()

	err = s.balance.DoTransfer(r.Context(), *transferParams)

	if err != nil {
		if errors.Is(err, e.DatabaseError) {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\": \"success\"}"))
}

func (s *Server) getBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	idRaw := r.URL.Query().Get("user_id")
	if idRaw == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"errorText\": \"missing required user_id parameter\"}"))
		return
	}
	var id int64
	var err error
	id, err = strconv.ParseInt(idRaw, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("{\"errorText\": \"incorrect user_id parameter\"}"))
		return
	}

	balance, err := s.balance.GetBalance(r.Context(), id)

	if err != nil {
		if errors.Is(err, e.DatabaseError) {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write([]byte(fmt.Sprintf("{\"errorText\": \"%s\"}", err)))
		return
	}

	response, err := json.Marshal(balance)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"errorText\": \"server error\"}"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
