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

func poissonExact(lambda float64, k int64) float64 {
	kFactorial := float64(1.0)
	for i := int64(1); i <= k; i++ {
		kFactorial *= float64(i)
	}
	return math.Pow(lambda, float64(k)) * math.Exp(-lambda) / kFactorial
}

func lambdaSmallEnoughForForPoissionCumulativeExceedsPercentageAtXLimit(percentage float64, xLimit int64) (lambdaResult int64, percentAchieved float64) {
	// Start with lambda = xLimit to give about 50% percentage
	// (The peak of the Poisson distribution is positioned horizontally at the top limit xLimit,
	// with half to the left of xLimit, and half to the right,
	// tailing down towards zero at x=infinity and x=-infinity)
	lambda := xLimit
	fraction := float64(0.5) // 50%
	for fraction < percentage/100.0 {
		// Decrease lambda to "squish" distribution leftwards,
		// bringing more of the area under the distribution to the left of xLimit
		// (increasing fraction)
		lambda--
		// Fraction is the cumulative sum of the Poisson Distribution up to xLimit
		fraction = 0.0
		for x := int64(0); x < xLimit; x++ {
			fraction += poissonApproximation(float64(lambda), float64(x))
		}
	}
	lambdaResult = lambda
	percentAchieved = fraction * 100.0
	return lambdaResult, percentAchieved
}

func xLimitBigEnoughForForPoissonCumulativeExceedsPercentageAtX(lambda float64, percentage float64) (xLimitResult int64) {
	fraction := 0.0
	for x := int64(0); true; x++ {
		poisson := poissonApproximation(lambda, float64(x))
		if math.IsNaN(poisson) {
			abc := 123
			abc++
		}
		fraction += poisson
		if fraction >= percentage/100.0 {
			return x
		}
	}
	return 1
}
