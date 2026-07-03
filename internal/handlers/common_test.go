package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Jedeft/demo-micro-gw-admin/internal/supports"
)

func TestUint32IDReq_Valid(t *testing.T) {
	_, missingPKMsg := supports.GetErrMsg(supports.MissingPKError)

	tests := []struct {
		name    string
		req     Uint32IDReq
		wantErr bool
		wantMsg string
	}{
		{name: "zero id", req: Uint32IDReq{ID: 0}, wantErr: true, wantMsg: missingPKMsg},
		{name: "valid id", req: Uint32IDReq{ID: 1}, wantErr: false},
		{name: "large id", req: Uint32IDReq{ID: 99999}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStringIDReq_Valid(t *testing.T) {
	_, missingPKMsg := supports.GetErrMsg(supports.MissingPKError)

	tests := []struct {
		name    string
		req     StringIDReq
		wantErr bool
		wantMsg string
	}{
		{name: "empty id", req: StringIDReq{ID: ""}, wantErr: true, wantMsg: missingPKMsg},
		{name: "valid id", req: StringIDReq{ID: "abc-123"}, wantErr: false},
		{name: "numeric string", req: StringIDReq{ID: "42"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Valid()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
