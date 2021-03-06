// Copyright (C) 2017, 2018 O.S. Systems Sofware LTDA
//
// SPDX-License-Identifier: Apache-2.0
use argh::FromArgs;
use slog_scope::info;
use std::path::PathBuf;
use lettre::transport::smtp::authentication::Credentials;
use lettre::{Message, SmtpTransport, Transport};
//add sdk support
use sdk::listener;

#[derive(FromArgs)]
/// Top-level command.
struct TopLevel {
    #[argh(subcommand)]
    entry_point: EntryPoints,
}

#[derive(FromArgs)]
#[argh(subcommand)]
enum EntryPoints {
    Client(ClientOptions),
    Server(ServerOptions),
}

#[derive(FromArgs)]
/// Client subcommand
#[argh(subcommand, name = "client")]
struct ClientOptions {
    #[argh(subcommand)]
    commands: ClientCommands,

    /// change the client socket to listen
    #[argh(option, default = "String::from(\"localhost:8080\")")]
    listen_socket: String,

    /// print the output in json format
    #[argh(switch)]
    json: bool,
}

#[derive(FromArgs)]
#[argh(subcommand)]
enum ClientCommands {
    Info(Info),
    Log(Log),
    Probe(Probe),
    AbortDownload(AbortDownload),
    LocalInstall(LocalInstall),
    RemoteInstall(RemoteInstall),
}

#[derive(FromArgs)]
/// Fetches information about the current state of the agent
#[argh(subcommand, name = "info")]
struct Info {}

#[derive(FromArgs)]
/// Fetches the available log entries for the last update cycle
#[argh(subcommand, name = "log")]
struct Log {}

#[derive(FromArgs)]
/// Checks if the server has a new update for this device.
///
/// A custom server for the update cycle can be specified via the ´--server´
#[argh(subcommand, name = "probe")]
struct Probe {
    /// custom address to try probe
    #[argh(option)]
    server: Option<String>,
}

#[derive(FromArgs)]
/// Abort current running download
#[argh(subcommand, name = "abort-download")]
struct AbortDownload {}

#[derive(FromArgs)]
/// Request agent to install a local update package
#[argh(subcommand, name = "local-install")]
struct LocalInstall {
    /// path to the update package
    #[argh(positional)]
    file: PathBuf,
}

#[derive(FromArgs)]
/// Request agent to download and install a package from a direct URL
#[argh(subcommand, name = "remote-install")]
struct RemoteInstall {
    /// the URL to get the update package
    #[argh(positional)]
    url: String,
}

#[derive(FromArgs)]
/// Server subcommand
#[argh(subcommand, name = "server")]
struct ServerOptions {
    /// increase the verboseness level
    #[argh(option, short = 'v', from_str_fn(verbosity_level), default = "slog::Level::Info")]
    verbosity: slog::Level,

    /// configuration file to use (defaults to "/etc/updatehub.conf")
    #[argh(option, short = 'c', default = "PathBuf::from(\"/etc/updatehub.conf\")")]
    config: PathBuf,
}

fn verbosity_level(value: &str) -> Result<slog::Level, String> {
    use std::str::FromStr;
    slog::Level::from_str(value).map_err(|_| format!("failed to parse verbosity level: {}", value))
}

async fn server_main(cmd: ServerOptions) -> updatehub::Result<()> {
    let _guard = updatehub::logger::init(cmd.verbosity);
    info!("starting UpdateHub Agent {}", updatehub::version());

    updatehub::run(&cmd.config).await?;

    Ok(())
}

async fn client_main(client_options: ClientOptions) -> updatehub::Result<()> {
    let client = sdk::Client::new(&client_options.listen_socket);

    match client_options.commands {
        ClientCommands::Info(_) => {
            let response = client.info().await?;

            if client_options.json {
                println!("{}", serde_json::to_string(&response)?);
            } else {
                println!("{:#?}", response);
            }
        }
        ClientCommands::Log(_) => {
            let response = client.log().await?;

            if client_options.json {
                println!("{}", serde_json::to_string(&response)?);
            } else {
                println!("{}", response);
            }
        }
        ClientCommands::Probe(Probe { server }) => {
            let response = client.probe(server).await?;

            if client_options.json {
                println!("{}", serde_json::to_string(&response)?);
            } else {
                match response {
                    sdk::api::probe::Response::Updating => {
                        println!("Update available. The update is running in background.")
                    }
                    sdk::api::probe::Response::NoUpdate => {
                        println!("There are no updates available.")
                    }
                    sdk::api::probe::Response::TryAgain(t) => {
                        println!("Server replied asking us to try again in {} seconds", t);
                    }
                }
            }
        }
        ClientCommands::AbortDownload(_) => {
            let response = client.abort_download().await?;

            if client_options.json {
                println!("{}", serde_json::to_string(&response)?);
            } else {
                println!("{:#?}", response);
            }
        }
        ClientCommands::LocalInstall(LocalInstall { file }) => {
            let file =
                if file.is_absolute() { file } else { std::env::current_dir().unwrap().join(file) };
            let response = client.local_install(&file).await?;

            if client_options.json {
                println!("{}", serde_json::to_string(&response)?);
            } else {
                println!("{:#?}", response);
            }
        }
        ClientCommands::RemoteInstall(RemoteInstall { url }) => {
            let response = client.remote_install(&url).await?;

            if client_options.json {
                println!("{}", serde_json::to_string(&response)?);
            } else {
                println!("{:#?}", response);
            }
        }
    }

    Ok(())
}

#[async_std::main]
async fn main() {

    let cmd: TopLevel = argh::from_env();

    let mut listener = listener::StateChange::default();
    listener.on_state(listener::State::Probe, |mut handler| async move {
        // email function
        let email = Message::builder()
            .from("Test <domarystest@gmail.com>".parse().unwrap())
            .reply_to("Teste <domarystest@gmail.com>".parse().unwrap())
            .to("Domy <domaryscorrea@gmail.com>".parse().unwrap())
            .subject("internal! in main, it's working")
            .body("callback!")
            .unwrap();

        let creds = Credentials::new("domarystest@gmail.com".to_string(), "@test1234".to_string());

        let mailer = SmtpTransport::relay("smtp.gmail.com")
            .unwrap()
            .credentials(creds)
            .build();

         let _result = mailer.send(&email);
        // keeping - stop the state
        handler.cancel().await
    });

    let res = match cmd.entry_point {
        EntryPoints::Client(client) => client_main(client).await,
        EntryPoints::Server(cmd) => server_main(cmd).await,
    };


    if let Err(e) = res {
        eprintln!("{}", e);
        std::process::exit(1);
    }
}
