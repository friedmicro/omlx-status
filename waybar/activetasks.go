package main

func (t ActiveTask) ETA() float64 {
	if t.PreETA > 0 {
		return t.PreETA
	}
	if t.GeneratedTokens == 0 || t.TokensPerSecond == 0 {
		return 0
	}
	remaining := float64(t.MaxTokens-t.GeneratedTokens) / t.TokensPerSecond
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (t ActiveTask) Speed() float64 {
	if t.PreSpeed > 0 {
		return t.PreSpeed
	}
	return t.TokensPerSecond
}

func (t ActiveTask) Phase() string {
	if t.PrePhase != "" {
		return t.PrePhase
	}
	return "generating"
}

func (t ActiveTask) Progress() (int, int, float64) {
	if t.PreTotal > 0 {
		pct := float64(t.PreProcessed) / float64(t.PreTotal) * 100
		if pct > 100 {
			pct = 100
		}
		return t.PreProcessed, t.PreTotal, pct
	}
	if t.MaxTokens <= 0 {
		return 0, 100, 0
	}
	processed := t.GeneratedTokens
	total := t.MaxTokens
	pct := float64(processed) / float64(total) * 100
	if pct > 100 {
		pct = 100
	}
	return processed, total, pct
}
