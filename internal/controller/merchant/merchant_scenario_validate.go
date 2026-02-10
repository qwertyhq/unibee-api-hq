package merchant

import (
	"context"

	_scenario "unibee/api/merchant/scenario"
	"unibee/internal/logic/scenario"
)

func (c *ControllerScenario) Validate(ctx context.Context, req *_scenario.ValidateReq) (res *_scenario.ValidateRes, err error) {
	dsl, err := scenario.ParseDSL(req.ScenarioJson)
	if err != nil {
		return &_scenario.ValidateRes{
			Valid:  false,
			Errors: []string{err.Error()},
		}, nil
	}

	errors := scenario.ValidateDSL(dsl)
	return &_scenario.ValidateRes{
		Valid:  len(errors) == 0,
		Errors: errors,
	}, nil
}
