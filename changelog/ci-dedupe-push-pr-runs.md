### Changed

- `ci` and `winparity` now run once per pull-request commit instead of twice: push triggers are limited to `main`, and a newer push to a pull request cancels the superseded run. Check names are unchanged.
