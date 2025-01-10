package indexedhashes3

import "math"

func poissonApproximation(lambda float64, x float64) float64 {
	// Normal distribution with the following parameters mu and sigma,
	// is a usable approximation to Poisson distribution for lambda > 20
	mu := lambda
	sigma := math.Sqrt(lambda)
	return normalDistribution(mu, sigma, x)
}

func normalDistribution(mu float64, sigma float64, x float64) float64 {
	sigmaSquared := sigma * sigma
	return math.Exp(-(x-mu)*(x-mu)/(2*sigmaSquared)) / math.Sqrt(2*math.Pi*sigmaSquared)
}
