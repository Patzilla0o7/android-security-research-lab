# Bootstrap

`lab bootstrap` prepares a supported Ubuntu 24.04 host for AOSP development.

## Commands

```bash
lab bootstrap
lab bootstrap plan
lab bootstrap --apply
```

- `plan` is the default. It checks `config/bootstrap.conf` with `dpkg-query`
  and `apt-cache`; it does not modify the host.
- `--apply` runs `sudo apt-get update` and installs missing packages. It is the
  only mode that changes the host and requires an interactive sudo session.

The package list follows the AOSP Ubuntu 18.04-and-later setup requirements,
with OpenJDK 17, repo, and ccache added for ASRL. After installation, run:

```bash
lab doctor
```

Do not run `--apply` until the plan output has been reviewed.
