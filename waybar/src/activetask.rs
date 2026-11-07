use crate::types::*;

impl ActiveTask {
    pub fn eta(&self) -> f64 {
        if self.eta > 0.0 {
            return self.eta;
        }
        if self.generated_tokens == 0 || self.tokens_per_second == 0.0 {
            return 0.0;
        }
        let remaining = (self.max_tokens - self.generated_tokens) as f64 / self.tokens_per_second;
        if remaining < 0.0 {
            return 0.0;
        } else {
            return remaining;
        }
    }

    pub fn speed(&self) -> f64 {
        if self.speed > 0.0 {
            self.speed
        } else {
            self.tokens_per_second
        }
    }

    pub fn phase(&self) -> &str {
        if !self.phase.is_empty() {
            &self.phase
        } else {
            "generating"
        }
    }

    pub fn progress(&self) -> (i64, i64, f64) {
        if self.total > 0 {
            let pct = self.processed as f64 / self.total as f64 * 100.0;
            return (self.processed, self.total, pct.min(100.0));
        }
        if self.max_tokens <= 0 {
            return (0, 100, 0.0);
        }
        let processed = self.generated_tokens;
        let total = self.max_tokens;
        let pct = processed as f64 / total as f64 * 100.0;
        return (processed, total, pct.min(100.0));
    }
}
