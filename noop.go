package bdd

import (
	"context"

	"github.com/cucumber/godog"
)

func noopTestSuiteHook() {}

func noopBeforeScenarioHook(ctx context.Context, sn *godog.Scenario) (context.Context, error) {
	return ctx, nil
}

func noopAfterScenarioHook(ctx context.Context, sn *godog.Scenario, err error) (context.Context, error) {
	return ctx, nil
}

func noopTestBeforeStepHook(ctx context.Context, sn *godog.Step) (context.Context, error) {
	return ctx, nil
}

func noopTestAfterStepHook(ctx context.Context, sn *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
	return ctx, nil
}
