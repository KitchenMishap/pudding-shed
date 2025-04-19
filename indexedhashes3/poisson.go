package indexedhashes3

import (
	"fmt"
	"math"
)

func poissonApproximation(lambda float64, x float64) float64 {
	// Normal distribution with the following parameters mu and sigma,
	// is a usable approximation to Poisson distribution for lambda > 20
	mu := lambda
	sigma := math.Sqrt(lambda)
	return normalDistribution(mu, sigma, x)
}

func normalDistribution(mu float64, sigma float64, x float64) float64 {
	sigmaSquared := sigma * sigma
	result := math.Exp(-(x-mu)*(x-mu)/(2*sigmaSquared)) / math.Sqrt(2*math.Pi*sigmaSquared)
	if math.IsNaN(result) {
		panic("normalDistribution: NaN")
	}
	return result
}

func poissonExact(lambda float64, k int64) float64 {
	kFactorial := float64(1.0)
	for i := int64(1); i <= k; i++ {
		kFactorial *= float64(i)
	}
	result := math.Pow(lambda, float64(k)) * math.Exp(-lambda) / kFactorial
	if math.IsNaN(result) {
		panic("poissonExact: NaN")
	}
	return result
}

func poissonBest(lambda float64, k int64) float64 {
	// Google AI overview
	// The largest value of 'k' for which k! (k factorial) can be precisely represented as a double-precision floating-point number is 170.
	// So we put a limit at k=50 (say) (We FIND that high values of kLimit will produce other NaN's due to very big interim numbers)
	kLimit := int64(35)
	// Below this limit we can use the Poissom Exact formula
	if k < kLimit {
		return poissonExact(lambda, k)
	} else {
		// Above this limit we have to resort to the approximation, aa k factorial is too big for a double
		return poissonApproximation(lambda, float64(k))
	}
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
			fraction += poissonBest(float64(lambda), x)
		}
	}
	lambdaResult = lambda
	percentAchieved = fraction * 100.0
	return lambdaResult, percentAchieved
}

func xLimitBigEnoughForForPoissonCumulativeExceedsPercentageAtX(lambda float64, percentage float64) (xLimitResult int64) {
	fraction := 0.0
	for x := int64(0); true; x++ {
		poisson := poissonBest(lambda, x)
		fraction += poisson
		if fraction >= percentage/100.0 {
			return x
		}
		if float64(x) > 1000000*lambda {
			fmt.Println("High accuracy copout occurring") // Sometimes we have to give up on reaching a high percentage!
			fmt.Println("So far... (giving up)... fraction = ", fraction)
			return x
		}
	}
	return 1
}
