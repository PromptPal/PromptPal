package service

import (
	"errors"
	"slices"
	"strings"
	"time"
)

var ErrorNoCostFound = errors.New("no cost found for")
var ErrorInvalidModel = errors.New("invalid model")

type ModelCost struct {
	StartFrom              time.Time
	InputTokenCostInCents  float64
	OutputTokenCostInCents float64
}

var costMap map[string][]ModelCost

func init() {
	costMap = map[string][]ModelCost{
		"o1-mini": {
			ModelCost{
				StartFrom:              time.Date(2024, 9, 12, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.0003,
				OutputTokenCostInCents: 0.0012,
			},
		},
		"o1": {
			ModelCost{
				StartFrom:              time.Date(2024, 9, 12, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.0015,
				OutputTokenCostInCents: 0.006,
			},
		},
		"gpt-3.5-turbo": {
			ModelCost{
				StartFrom:              time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.00005,
				OutputTokenCostInCents: 0.00015,
			},
		},
		"gpt-4-turbo": {
			ModelCost{
				StartFrom:              time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.001,
				OutputTokenCostInCents: 0.003,
			},
		},
		"gpt-4o": {
			ModelCost{
				StartFrom:              time.Date(2024, 5, 14, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.0005,
				OutputTokenCostInCents: 0.0015,
			},
		},
		"gpt-4o-mini": {
			ModelCost{
				StartFrom:              time.Date(2024, 7, 19, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.000015,
				OutputTokenCostInCents: 0.00006,
			},
		},
		"o3": {
			ModelCost{
				StartFrom:              time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.000015,
				OutputTokenCostInCents: 0.00006,
			},
		},
		"o3-mini": {
			ModelCost{
				StartFrom:              time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.0000011,
				OutputTokenCostInCents: 0.00000055,
			},
		},
		"gemini-pro": {
			ModelCost{
				StartFrom:              time.Date(2024, time.May, 14, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.0007,
				OutputTokenCostInCents: 0.0021,
			},
		},
		"gemini-1.5-flash": {
			ModelCost{
				StartFrom:              time.Date(2024, time.October, 1, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.0000075,
				OutputTokenCostInCents: 0.000003,
			},
		},
		"gemini-1.5-pro": {
			ModelCost{
				StartFrom:              time.Date(2024, time.October, 1, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.000125,
				OutputTokenCostInCents: 0.0005,
			},
		},
		"deepseek-chat": {
			ModelCost{
				StartFrom:              time.Date(2024, time.October, 1, 0, 0, 0, 0, time.UTC),
				InputTokenCostInCents:  0.0027,
				OutputTokenCostInCents: 0.011,
			},
		},
	}
}

func GetCosts(model string, currentAt time.Time) (*ModelCost, error) {
	modelCostList, ok := costMap[strings.ToLower(model)]
	if !ok {
		return nil, ErrorInvalidModel
	}

	slices.SortFunc(modelCostList, func(a, b ModelCost) int {
		return a.StartFrom.Compare(b.StartFrom)
	})

	for i := len(modelCostList) - 1; i >= 0; i-- {
		if currentAt.After(modelCostList[i].StartFrom) {
			return &modelCostList[i], nil
		}
	}

	return nil, ErrorNoCostFound
}
