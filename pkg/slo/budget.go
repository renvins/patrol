package slo

func BurnRate(errorRate float64, objective float64) float64 {
	return errorRate / (1 - (objective / 100))
}

func BudgetConsumed(errorRate float64, objective float64) float64 {
	return BurnRate(errorRate, objective)
}

func BudgetRemaining(errorRate float64, objective float64) float64 {
	return (1 - BudgetConsumed(errorRate, objective)) * 100
}
