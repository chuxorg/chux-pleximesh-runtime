package qaharness

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/chuxorg/chux-agent-mesh/internal/messagebus"
)

func TestHarnessProducesSerializableReport(t *testing.T) {
	router := messagebus.NewRouter()
	harness := New(router, WithLoadEvents(2000))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	report, err := harness.Run(ctx)
	if err != nil {
		t.Fatalf("run harness: %v", err)
	}
	if !report.Passed {
		t.Fatalf("expected harness to pass scenarios, report: %#v", report)
	}
	if len(report.Scenarios) != 2 {
		t.Fatalf("expected two scenarios, got %d", len(report.Scenarios))
	}
	invalid := scenarioByName(t, report, scenarioInvalidEnvelope)
	if !invalid.Passed {
		t.Fatalf("invalid envelope scenario failed: %+v", invalid)
	}
	if invalid.Metrics["dispatch_attempts"] != 1 {
		t.Fatalf("invalid scenario metrics not recorded: %+v", invalid.Metrics)
	}
	if invalid.Details["validation_error"] == "" {
		t.Fatalf("missing validation evidence: %+v", invalid.Details)
	}

	load := scenarioByName(t, report, scenarioLoadFanOut)
	if !load.Passed {
		t.Fatalf("load scenario failed: %+v", load)
	}
	if load.Metrics["events_dispatched"] != 2000 {
		t.Fatalf("expected 2000 dispatched, got %+v", load.Metrics)
	}
	for i := 0; i < int(load.Metrics["subscribers"]); i++ {
		key := subscriberMetricKey(i)
		if load.Metrics[key] != 2000 {
			t.Fatalf("subscriber %d missing events: %+v", i, load.Metrics)
		}
	}

	if _, err := json.Marshal(report); err != nil {
		t.Fatalf("report should serialize: %v", err)
	}
}

func TestHarnessLoadScenarioHandlesTenThousandEvents(t *testing.T) {
	router := messagebus.NewRouter()
	harness := New(router)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report, err := harness.Run(ctx)
	if err != nil {
		t.Fatalf("run harness: %v", err)
	}
	if !report.Passed {
		t.Fatalf("expected harness to pass scenarios, report: %#v", report)
	}
	load := scenarioByName(t, report, scenarioLoadFanOut)
	if load.Metrics["events_dispatched"] < 10000 {
		t.Fatalf("expected load scenario >= 10k, metrics: %+v", load.Metrics)
	}
	expected := load.Metrics["events_dispatched"]
	for i := 0; i < int(load.Metrics["subscribers"]); i++ {
		key := subscriberMetricKey(i)
		if load.Metrics[key] != expected {
			t.Fatalf("subscriber %d lost events: %+v", i, load.Metrics)
		}
	}
}

func scenarioByName(t *testing.T, report Report, name string) ScenarioResult {
	t.Helper()
	for _, scenario := range report.Scenarios {
		if scenario.Name == name {
			return scenario
		}
	}
	t.Fatalf("scenario %s not found in report: %+v", name, report.Scenarios)
	return ScenarioResult{}
}

func subscriberMetricKey(idx int) string {
	return "subscriber_" + strconv.Itoa(idx) + "_received"
}
