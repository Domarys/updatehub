// Copyright (C) 2020 O.S. Systems Sofware LTDA
//
// SPDX-License-Identifier: Apache-2.0

use super::{Error, Result};
use crate::{
    object::{Info, Installer},
    utils::{self, definitions::TargetTypeExt},
};

use pkg_schema::{definitions, objects};
use slog_scope::info;

impl Installer for objects::Bita {
    fn check_requirements(&self) -> Result<()> {
        info!("'bita' handle checking requirements");
        utils::fs::is_executable_in_path("bita")?;

        if let definitions::TargetType::Device(dev) = self.target.valid()? {
            utils::fs::ensure_disk_space(&dev, self.required_install_size())?;
            return Ok(());
        }

        Err(Error::InvalidTargetType(self.target.clone()))
    }

    fn install(&self, download_dir: &std::path::Path) -> Result<()> {
        info!("'bita' handler Install {} ({})", self.filename, self.sha256sum);

        let target = self.target.get_target()?;
        let url = &self.archive;
        // FIXME: Do we need a source?
        let _source = download_dir.join(self.sha256sum());

        easy_process::run(&format!(
            "bita clone --seed-output {} {}",
            url,
            target.to_string_lossy()
        ))?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::object::installer::tests::create_echo_bins;
    use pretty_assertions::assert_eq;
    use std::env;

    fn fake_bita_obj() -> objects::Bita {
        objects::Bita {
            filename: "etc/passwd".to_string(),
            sha256sum: "cfe2be1c64b03875008".to_string(),
            target: definitions::TargetType::Device(std::path::PathBuf::from("/dev/sda1")),

            archive: "https://foo.bar/bita_archive".to_string(),
            size: 1024,
        }
    }

    #[test]
    fn check_requirements_with_missing_binary() {
        let bita_obj = fake_bita_obj();

        env::set_var("PATH", "");
        let (_handle, _) = create_echo_bins(&["bita"]).unwrap();
        assert!(bita_obj.check_requirements().is_err());
    }

    #[test]
    #[ignore]
    fn install_commands() {
        let bita_obj = fake_bita_obj();
        let download_dir = tempfile::tempdir().unwrap();

        let (_handle, calls) = create_echo_bins(&["bita"]).unwrap();

        bita_obj.check_requirements().unwrap();
        bita_obj.install(download_dir.path()).unwrap();

        let expected =
            String::from("bita clone --seed-output https://foo.bar/bita_archive /dev/sda1\n");

        assert_eq!(std::fs::read_to_string(calls).unwrap(), expected);
    }
}
