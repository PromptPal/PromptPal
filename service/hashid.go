package service

import (
	"errors"

	"github.com/PromptPal/PromptPal/config"
	"github.com/speps/go-hashids/v2"
)

type HashIDService interface {
	Encode(id int) (string, error)
	Decode(id string) (int, error)
}

type hashIDService struct {
	h *hashids.HashID
}

func NewHashIDService() HashIDService {
	hd := hashids.NewData()
	hd.Salt = config.GetRuntimeConfig().HashidSalt
	hd.MinLength = 12
	h, _ := hashids.NewWithData(hd)

	return &hashIDService{
		h: h,
	}
}

func (h *hashIDService) Encode(id int) (string, error) {
	return h.h.Encode([]int{id})
}

func (h *hashIDService) Decode(id string) (int, error) {
	list, err := h.h.DecodeWithError(id)

	if err != nil {
		return 0, err
	}

	if len(list) == 0 {
		return 0, errors.New("invalid id")
	}

	return list[0], nil
}
