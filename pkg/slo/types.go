package slo

type SLOConfig struct {
	SLOs []SLO `json:"slos" yaml:"slos"`
}

// Service Level Objective -> target for the SLI
// What we want to achieve and what should be happening
type SLO struct { // go conventions for acronyms
	Name           string          `json:"name" yaml:"name"`
	Service        string          `json:"service" yaml:"service"`
	SLI            SLI             `json:"sli" yaml:"sli"`
	Objective      float64         `json:"objective" yaml:"objective"` // 99.9 requests to succeed
	Window         string          `json:"window" yaml:"window"`       // measured over a rolling 30-day window
	BurnRateAlerts []BurnRateAlert `json:"burn_rate_alerts" yaml:"burn_rate_alerts"`
	Policies       []Policy        `json:"policies" yaml:"policies"`
}

// Service Level Indicator -> it's a measurement that describes the service behavior
type SLI struct {
	Source string `json:"source" yaml:"source"` // where data come from
	Query  string `json:"query" yaml:"query"`   // the formula used
}

// how fast you're consuming your error budget relative to normal
// 1 exactly sustainable
type BurnRateAlert struct {
	Window    string  `json:"window" yaml:"window"`       // how long to measure the burn rate over
	Threshold float64 `json:"threshold" yaml:"threshold"` // burn rate number that triggers the alert
	Severity  string  `json:"severity" yaml:"severity"`   // critical vs warning
}

// when the budget is low or is burning fast
// a policy is triggered
type Policy struct {
	When   string `json:"when" yaml:"when"`
	Action string `json:"action" yaml:"action"`
	Target string `json:"target" yaml:"target"`
}
