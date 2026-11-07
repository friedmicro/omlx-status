pub mod activetask;
pub mod config;
pub mod types;

use crate::config::config;
use crate::types::*;
use std::io::{BufRead, BufReader, Read, Write};
use std::os::unix::net::UnixStream;

const DAEMON_SOCKET: &str = "/tmp/waybar-damon";
const DAEMON_KEY: &str = "omlxstatus";

const MACHINE_BUSY_SYMBOL: &str = "🔴";
const MACHINE_FREE_SYMBOL: &str = "🟢";

fn request_key() -> Result<String, String> {
    let mut conn =
        UnixStream::connect(DAEMON_SOCKET).map_err(|e| format!("connect to daemon: {e}"))?;
    conn.write_all(format!("{DAEMON_KEY}\n").as_bytes())
        .map_err(|e| format!("write request: {e}"))?;

    let mut reader = BufReader::new(&mut conn);
    let mut line = String::new();
    reader
        .read_line(&mut line)
        .map_err(|e| format!("read status: {e}"))?;
    if !line.starts_with("OK") {
        return Err(format!("daemon: {}", line.trim()));
    }
    let mut body = String::new();
    reader
        .read_to_string(&mut body)
        .map_err(|e| format!("read body: {e}"))?;
    Ok(body)
}

fn fetch_status() -> Result<APIResponse, String> {
    let body = request_key()?;
    serde_json::from_str(&body).map_err(|e| format!("parse status: {e}"))
}

fn process_icon_status(machine: &Machine) -> &'static str {
    if machine.maintenance || machine.busy {
        MACHINE_BUSY_SYMBOL
    } else {
        MACHINE_FREE_SYMBOL
    }
}

fn process_status(machine: &Machine) -> &'static str {
    if machine.maintenance {
        "maintenance"
    } else if machine.busy {
        "busy"
    } else {
        "ok"
    }
}

fn get_active_task(machine: &Machine) -> Option<&ActiveTask> {
    machine
        .prefilling
        .as_ref()
        .and_then(|tasks| tasks.first())
        .or_else(|| machine.waiting.as_ref().and_then(|tasks| tasks.first()))
        .or_else(|| machine.generating.as_ref().and_then(|tasks| tasks.first()))
}

fn format_eta(seconds: f64) -> String {
    if seconds <= 0.0 {
        return "N/A".to_string();
    }
    let hours = (seconds / 3600.0) as i64;
    let minutes = (seconds.rem_euclid(3600.0)) as i64 / 60;
    let secs = (seconds.rem_euclid(60.0)) as i64;

    if hours > 0 {
        format!("{hours}h{minutes}m")
    } else if minutes > 0 {
        format!("{minutes}m{secs}s")
    } else {
        format!("{secs}s")
    }
}

fn format_memory(bytes: f64) -> String {
    let gb = bytes / (1024.0 * 1024.0 * 1024.0);
    if gb >= 1.0 {
        format!("{gb:.1} GB")
    } else {
        let mb = bytes / (1024.0 * 1024.0);
        format!("{mb:.1} MB")
    }
}

fn process_active_task(active_task: &ActiveTask) -> String {
    let eta = active_task.eta();
    let mut tooltip = String::new();
    if eta > 0.0 {
        tooltip += &format!("\n  eta: {}", format_eta(eta));
    }
    tooltip += &format!("\n  phase: {}", active_task.phase());
    let (processed, total, pct) = active_task.progress();
    tooltip += &format!("\n  progress: {processed}/{total} ({pct:.1}%)");
    tooltip += &format!("\n  speed: {:.1} tok/s", active_task.speed());
    return tooltip;
}

fn process_tooltip(machine: &Machine, key: &str) -> String {
    let status = process_status(machine);
    let mut tooltip = String::new();
    tooltip += &format!("{key}: {status}");
    tooltip += &format!("\n  model: {}", machine.model);

    if let Some(active_task) = get_active_task(machine) {
        tooltip += &process_active_task(active_task);
    }

    tooltip += &format!(
        "\n  memory: {} / {}",
        format_memory(machine.model_memory_used),
        format_memory(machine.model_memory_max)
    );
    return tooltip;
}

fn process_css(machines: &APIResponse) -> &'static str {
    for key in &config().backends {
        match machines.get(key) {
            None => return "unhealthy",
            Some(m) if m.maintenance => return "unhealthy",
            _ => {}
        }
    }
    return "ok";
}

fn render(machines: &APIResponse) -> WaybarOutput {
    let mut out_text = String::new();
    let mut tooltip = String::new();
    for key in &config().backends {
        match machines.get(key) {
            Some(machine) => {
                let icon = process_icon_status(machine);
                tooltip += &process_tooltip(machine, key);
                out_text += icon;
            }
            None => {
                out_text += "⚪";
                if !tooltip.is_empty() {
                    tooltip += "\n";
                }
                tooltip += &format!("{key}: not available");
            }
        }
    }

    WaybarOutput {
        text: out_text,
        tooltip: tooltip,
        class: process_css(&machines),
    }
}

fn handle_socket_error(e: String) {
    let out = WaybarOutput {
        text: String::from("FATAL ERROR PARSING RESPONSE FROM UNIX SOCKET"),
        class: "unhealthy",
        tooltip: e,
    };
    println!("{}", serde_json::to_string(&out).unwrap());
    std::process::exit(1);
}

fn main() {
    match fetch_status() {
        Ok(s) => println!("{}", serde_json::to_string(&render(&s)).unwrap()),
        Err(e) => handle_socket_error(e),
    }
}
