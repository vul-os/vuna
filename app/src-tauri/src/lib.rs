mod commands;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![
            commands::search,
            commands::node_status,
            commands::stats,
        ])
        .run(tauri::generate_context!())
        .expect("error while running the Vuna tauri application");
}
