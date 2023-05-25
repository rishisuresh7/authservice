package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"authservice/factory"
	"authservice/models"
	"authservice/response"
)

func GetAddresses(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address := f.Address()
		res, err := address.GetAddresses(r.Context(), r.Header.Get("userId"))
		if err != nil {
			l.Errorf("GetAddresses: unable to get addresses: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: res}.Send(w)
	}
}

func GetAddress(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		addressId, ok := vars["addressId"]
		if !ok {
			l.Errorf("GetAddress: unable to read 'addressId' from path")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		address := f.Address()
		res, err := address.GetAddress(r.Context(), r.Header.Get("userId"), addressId)
		if err != nil {
			l.Errorf("GetAddress: unable to get address: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: res}.Send(w)
	}
}

func CreateAddress(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var addr models.Address
		err := json.NewDecoder(r.Body).Decode(&addr)
		if err != nil {
			l.Errorf("CreateAddress: unable to decode payload: %s", err)
			response.Error{Error: "invalid request payload"}.ClientError(w)
			return
		}

		err = addr.Validate()
		if err != nil {
			l.Errorf("CreateAddress: invalid address payload: %s", err)
			response.Error{Error: "invalid request payload"}.ClientError(w)
			return
		}

		address := f.Address()
		res, err := address.CreateAddress(r.Context(), fmt.Sprintf("%s", r.Header.Get("userId")), &addr)
		if err != nil {
			l.Errorf("CreateAddress: unable to create address: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: res}.Send(w)
	}
}

func UpdateAddress(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		addressId, ok := vars["addressId"]
		if !ok {
			l.Errorf("UpdateAddress: unable to read 'addressId' from path")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		var addr models.Address
		err := json.NewDecoder(r.Body).Decode(&addr)
		if err != nil {
			l.Errorf("UpdateAddress: unable to decode payload: %s", err)
			response.Error{Error: "invalid request payload"}.ClientError(w)
			return
		}

		err = addr.Validate()
		if err != nil {
			l.Errorf("UpdateAddress: invalid address payload: %s", err)
			response.Error{Error: "invalid request payload"}.ClientError(w)
			return
		}

		address := f.Address()
		res, err := address.UpdateAddress(r.Context(), fmt.Sprintf("%s", r.Header.Get("userId")), addressId, &addr)
		if err != nil {
			l.Errorf("UpdateAddress: unable to create address: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: res}.Send(w)
	}
}

func DeleteAddress(f factory.Factory, l *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		addressId, ok := vars["addressId"]
		if !ok {
			l.Errorf("DeleteAddress: unable to read 'addressId' from path")
			response.Error{Error: "invalid request"}.ClientError(w)
			return
		}

		address := f.Address()
		err := address.DeleteAddress(r.Context(), fmt.Sprintf("%s", r.Header.Get("userId")), addressId)
		if err != nil {
			l.Errorf("DeleteAddress: unable to delete address: %s", err)
			response.Error{Error: "unexpected error happened"}.ServerError(w)
			return
		}

		response.Success{Success: "deleted address successfully"}.Send(w)
	}
}