# Bootstrap

`lab bootstrap` prepares a supported Ubuntu 24.04 host for AOSP development.

## Commands

```bash
lab bootstrap
lab bootstrap plan
lab bootstrap --apply
```

- `plan` is the default. It runs the same toolchain checks as `lab doctor`,
  then displays available installation methods; it does not modify the host.
- `--apply` runs `sudo apt-get update` and installs only missing tools with an
  `apt` method. Tools with a `manual` method remain visible for the operator.

The shared [tool registry](../config/tools.conf) follows the AOSP Ubuntu
18.04-and-later setup requirements, with OpenJDK 17, repo, and ccache added
for ASRL. After installation, run:

```bash
lab doctor
```

Do not run `--apply` until the plan output has been reviewed.
