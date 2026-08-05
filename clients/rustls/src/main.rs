use reqwest::Client;
use std::env;
use std::net::ToSocketAddrs;
use std::time::Duration;

#[tokio::main]
async fn main() {
    let id = env::var("TARGET_ID").unwrap_or_else(|_| "rust-rustls".into());
    let dial = env::var("DIAL_HOST").unwrap_or_else(|_| "capture:8443".into());
    let sni = format!("{}.fp.lab.local", id);
    let url = format!("https://{}/", sni);

    let addr = dial
        .to_socket_addrs()
        .expect("resolve DIAL_HOST")
        .next()
        .expect("no addr for DIAL_HOST");

    let client = Client::builder()
        .danger_accept_invalid_certs(true)
        .timeout(Duration::from_secs(20))
        .resolve(&sni, addr)
        .build()
        .expect("client");

    match client
        .get(&url)
        .header("X-Target-Id", &id)
        .send()
        .await
    {
        Ok(resp) => {
            let status = resp.status();
            let body = resp.text().await.unwrap_or_default();
            let n = body.len().min(200);
            println!("{} {}", status, &body[..n]);
        }
        Err(e) => eprintln!("request error (CH may still be saved): {e}"),
    }
}
