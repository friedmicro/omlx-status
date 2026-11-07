use crate::types::*;
use serde::Deserialize;
use std::fs;
use std::path::PathBuf;
use std::sync::OnceLock;

const CONFIG_FILE: &str = "omlx-waybar.yaml";
static CONFIG: OnceLock<Config> = OnceLock::new();

pub fn config() -> &'static Config {
    CONFIG.get_or_init(load_config)
}

fn config_path() -> Option<PathBuf> {
    std::env::current_exe()
        .ok()
        .and_then(|exe| exe.parent().map(|dir| dir.join(CONFIG_FILE)))
}

fn load_config() -> Config {
    let mut cfg = Config {
        backends: vec!["small".to_string(), "mid".to_string(), "big".to_string()],
    };

    let Some(path) = config_path() else {
        return cfg;
    };
    let Ok(contents) = fs::read_to_string(&path) else {
        return cfg;
    };

    #[derive(Deserialize, Default)]
    struct FileConfig {
        #[serde(default)]
        backends: Vec<String>,
    }

    let file_cfg: FileConfig = match serde_yaml::from_str(&contents) {
        Ok(f) => f,
        Err(e) => {
            eprintln!("error parsing {CONFIG_FILE}: {e}");
            return cfg;
        }
    };

    if !file_cfg.backends.is_empty() {
        cfg.backends = file_cfg.backends;
    }
    cfg
}
