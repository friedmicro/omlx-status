use serde::Deserialize;
use std::collections::HashMap;

#[derive(serde::Serialize)]
pub struct WaybarOutput {
    pub text: String,
    pub tooltip: String,
    pub class: &'static str,
}

#[derive(Default)]
pub struct Config {
    pub backends: Vec<String>,
}

#[derive(Default, Deserialize)]
pub struct ActiveTask {
    #[serde(default)]
    pub generated_tokens: i64,
    #[serde(default)]
    pub tokens_per_second: f64,
    #[serde(default)]
    pub max_tokens: i64,
    #[serde(default)]
    pub processed: i64,
    #[serde(default)]
    pub total: i64,
    #[serde(default)]
    pub speed: f64,
    #[serde(default)]
    pub eta: f64,
    #[serde(default)]
    pub phase: String,
}

#[derive(Default, Deserialize)]
pub struct Machine {
    #[serde(default)]
    pub model: String,
    #[serde(default)]
    pub model_memory_used: f64,
    #[serde(default)]
    pub model_memory_max: f64,
    #[serde(default)]
    pub waiting: Option<Vec<ActiveTask>>,
    #[serde(default)]
    pub prefilling: Option<Vec<ActiveTask>>,
    #[serde(default)]
    pub generating: Option<Vec<ActiveTask>>,
    #[serde(default)]
    pub busy: bool,
    #[serde(default)]
    pub maintenance: bool,
}

pub type APIResponse = HashMap<String, Machine>;
