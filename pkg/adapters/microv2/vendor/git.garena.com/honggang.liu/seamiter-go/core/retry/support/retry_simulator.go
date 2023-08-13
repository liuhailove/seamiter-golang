package support

import (
	"errors"
	"git.garena.com/honggang.liu/seamiter-go/core/retry"
	"git.garena.com/honggang.liu/seamiter-go/core/retry/backoff"
	"sort"
	"strconv"
)

// *{@link RetrySimulator} 是一种用于执行重试 + 退避操作的工具。
//
// * 在校准一组重试 + 退避对时，了解各种场景的重试行为很有用。
//
// * 你可能想知道的事情：
// * - 我的回退中 5000 毫秒的“maxInterval”是否重要？ （当重试次数很低时通常会出现这种情况——那么为什么要将最大间隔设置为无法实现的值呢？）
// - 重试中线程的典型睡眠持续时间是多少
// - 任何重试序列的最长睡眠持续时间是多少
//
// * 模拟器通过执行重试+退避对直到失败提供此信息
// （即所有重试都用完了）。有关每次重试的信息作为 {@link RetrySimulation} 的一部分提供。
//
// * 请注意，这门课的动机是公开可能的时间安排
// * {@link org.springframework.retry.backoff.ExponentialRandomBackOffPolicy}，它提供
//随机值，必须通过一系列试验来观察。

type RetrySimulator struct {
	// 回退策略
	BackOffPolicy backoff.SleepingBackOffPolicy
	// 重试策略
	RtyPolicy retry.RtyPolicy
}

func NewRetrySimulator(backoffPolicy backoff.SleepingBackOffPolicy, rtyPolicy retry.RtyPolicy) *RetrySimulator {
	var simulation = new(RetrySimulator)
	simulation.BackOffPolicy = backoffPolicy
	simulation.RtyPolicy = rtyPolicy
	return simulation
}

// ExecuteSimulation 按照给定的迭代次数numSimulations，进行重试
func (r *RetrySimulator) ExecuteSimulation(numSimulations int) (*RetrySimulation, error) {
	var simulation = new(RetrySimulation)
	for i := 0; i < numSimulations; i++ {
		var sleeps, err = r.ExecuteSingleSimulation()
		if err != nil {
			return nil, err
		}
		simulation.AddSequence(sleeps)
	}
	return simulation, nil
}

// ExecuteSingleSimulation 执行单个语义
func (r *RetrySimulator) ExecuteSingleSimulation() ([]int64, error) {
	var stealingSleeper = &StealingSleeper{}
	var stealingBackoff = r.BackOffPolicy.WithSleeper(stealingSleeper)

	var rtyTemplate = new(RetryTemplate)
	rtyTemplate.BackOffPolicy = stealingBackoff.(backoff.BackOffPolicy)
	rtyTemplate.RetryPolicy = r.RtyPolicy
	var _, err = rtyTemplate.Execute(&FailingRetryCallback{})
	if err != nil {
		return nil, errors.New("Unexpected exception" + err.Error())
	}
	return stealingSleeper.GetSleeps(), nil
}

type StealingSleeper struct {
	Sleeps []int64
}

func (s *StealingSleeper) Sleep(backOffPeriodInMs int64) {
	s.Sleeps = append(s.Sleeps, backOffPeriodInMs)
}

func (s *StealingSleeper) GetSleeps() []int64 {
	return s.Sleeps
}

type FailingRetryCallback struct {
}

func (f FailingRetryCallback) DoWithRetry(content retry.RtyContext) interface{} {
	panic("FailingRetryException")
}

type RetrySimulation struct {
	SleepSequences []*SleepSequence
	SleepHistogram []int64
}

func NewRetrySimulation() *RetrySimulation {
	return &RetrySimulation{}
}

func (r *RetrySimulation) AddSequence(sleeps []int64) {
	for _, sleep := range sleeps {
		r.SleepHistogram = append(r.SleepHistogram, sleep)
	}
	r.SleepSequences = append(r.SleepSequences, NewSleepSequence(sleeps))
}

func (r *RetrySimulation) GetPercentiles() []float64 {
	var res []float64
	var percentiles = []float64{10, 20, 30, 40, 50, 60, 70, 80, 90}
	for _, percentile := range percentiles {
		res = append(res, r.GetPercentile(percentile/100))
	}
	return res
}

func (r *RetrySimulation) GetPercentile(p float64) float64 {
	sort.Slice(r.SleepHistogram, func(i, j int) bool {
		return i < j
	})
	var size = len(r.SleepSequences)
	var pos = p * float64(size-1)
	var i0 = int32(pos)
	var i1 = i0 + 1
	var weight = pos - float64(i0)
	return float64(r.SleepHistogram[i0])*(1-weight) + float64(r.SleepHistogram[i1])*weight
}

func (r *RetrySimulation) GetLongestTotalSleepSequence() *SleepSequence {
	var longest *SleepSequence
	for _, sequence := range r.SleepSequences {
		if longest == nil || sequence.GetTotalSleep() > longest.GetTotalSleep() {
			longest = sequence
		}
	}
	return longest
}

type SleepSequence struct {
	Sleeps       []int64
	LongestSleep int64
	TotalSleep   int64
}

func NewSleepSequence(sleeps []int64) *SleepSequence {
	var sleepSequence = new(SleepSequence)
	sleepSequence.Sleeps = sleeps
	var longestSleep int64
	var totalSleep int64
	for _, sleep := range sleeps {
		if sleep > longestSleep {
			longestSleep = sleep
		}
		totalSleep += sleep
	}
	sleepSequence.LongestSleep = longestSleep
	sleepSequence.TotalSleep = totalSleep
	return sleepSequence
}

func (s *SleepSequence) GetSleeps() []int64 {
	return s.Sleeps
}
func (s *SleepSequence) GetLongestSleep() int64 {
	return s.LongestSleep
}

func (s *SleepSequence) GetTotalSleep() int64 {
	return s.TotalSleep
}

func (s *SleepSequence) String() string {
	var str string
	for _, sleep := range s.Sleeps {
		str += strconv.FormatInt(sleep, 10) + ","
	}
	return "totalSleep=" + strconv.FormatInt(s.GetTotalSleep(), 10) + ": " + str
}
