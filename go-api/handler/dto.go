package handler

import (
	"github.com/andrucar25/interseguro-coding-challenge/go-api/external/statistics"
	"github.com/andrucar25/interseguro-coding-challenge/go-api/qr"
)

type qrRequest struct {
	Matrix qr.Matrix `json:"matrix"`
}

type qrResponse struct {
	Q          qr.Matrix         `json:"q"`
	R          qr.Matrix         `json:"r"`
	Statistics statistics.Result `json:"statistics"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
