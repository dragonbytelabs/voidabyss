package editor

import (
	"time"
)

// PerformanceStats tracks performance metrics for the editor
type PerformanceStats struct {
	// Line calculation metrics
	LineStartsCalls     int
	LineStartsTotalTime time.Duration

	// Rendering metrics
	RenderCalls     int
	RenderTotalTime time.Duration

	// Search metrics
	SearchCalls     int
	SearchTotalTime time.Duration

	// Buffer operation metrics
	BufferOpCalls     int
	BufferOpTotalTime time.Duration
}

// Global performance stats (disabled by default)
var globalPerfStats *PerformanceStats
var perfEnabled = false

// EnablePerformanceTracking enables performance tracking
func EnablePerformanceTracking() {
	perfEnabled = true
	globalPerfStats = &PerformanceStats{}
}

// DisablePerformanceTracking disables performance tracking
func DisablePerformanceTracking() {
	perfEnabled = false
	globalPerfStats = nil
}

// GetPerformanceStats returns a copy of current performance stats
func GetPerformanceStats() PerformanceStats {
	if globalPerfStats == nil {
		return PerformanceStats{}
	}
	return *globalPerfStats
}

// ResetPerformanceStats resets all performance counters
func ResetPerformanceStats() {
	if globalPerfStats != nil {
		*globalPerfStats = PerformanceStats{}
	}
}

// trackLineStarts measures time spent in lineStarts()
func (e *Editor) trackLineStarts(fn func() []int) []int {
	if !perfEnabled || globalPerfStats == nil {
		return fn()
	}

	start := time.Now()
	result := fn()
	elapsed := time.Since(start)

	globalPerfStats.LineStartsCalls++
	globalPerfStats.LineStartsTotalTime += elapsed

	return result
}

// trackRender measures time spent in draw()
func (e *Editor) trackRender(fn func()) {
	if !perfEnabled || globalPerfStats == nil {
		fn()
		return
	}

	start := time.Now()
	fn()
	elapsed := time.Since(start)

	globalPerfStats.RenderCalls++
	globalPerfStats.RenderTotalTime += elapsed
}

// trackSearch measures time spent in search operations
func (e *Editor) trackSearch(fn func()) {
	if !perfEnabled || globalPerfStats == nil {
		fn()
		return
	}

	start := time.Now()
	fn()
	elapsed := time.Since(start)

	globalPerfStats.SearchCalls++
	globalPerfStats.SearchTotalTime += elapsed
}

// trackBufferOp measures time spent in buffer operations
func (e *Editor) trackBufferOp(fn func()) {
	if !perfEnabled || globalPerfStats == nil {
		fn()
		return
	}

	start := time.Now()
	fn()
	elapsed := time.Since(start)

	globalPerfStats.BufferOpCalls++
	globalPerfStats.BufferOpTotalTime += elapsed
}

// GetAverageLineStartsTime returns average time per lineStarts() call
func (s *PerformanceStats) GetAverageLineStartsTime() time.Duration {
	if s.LineStartsCalls == 0 {
		return 0
	}
	return s.LineStartsTotalTime / time.Duration(s.LineStartsCalls)
}

// GetAverageRenderTime returns average time per render call
func (s *PerformanceStats) GetAverageRenderTime() time.Duration {
	if s.RenderCalls == 0 {
		return 0
	}
	return s.RenderTotalTime / time.Duration(s.RenderCalls)
}

// GetAverageSearchTime returns average time per search call
func (s *PerformanceStats) GetAverageSearchTime() time.Duration {
	if s.SearchCalls == 0 {
		return 0
	}
	return s.SearchTotalTime / time.Duration(s.SearchCalls)
}

// GetAverageBufferOpTime returns average time per buffer operation
func (s *PerformanceStats) GetAverageBufferOpTime() time.Duration {
	if s.BufferOpCalls == 0 {
		return 0
	}
	return s.BufferOpTotalTime / time.Duration(s.BufferOpCalls)
}
