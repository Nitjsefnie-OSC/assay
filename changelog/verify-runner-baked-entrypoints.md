### Added
- the verify-runner image now ships the cluster backend's three fixed entrypoints,
`stage-source.sh`, `stage-cache.sh` and `witness.sh`, at `/usr/local/bin`, root-owned
and mode 0755, so an admission policy that pins each container's command to one
baked script path can admit the stage and witness Jobs. The image's ENTRYPOINT and
CMD are unchanged. The release probe (`images/verify-runner/check.py`) now fails the
build unless each script is a root-owned executable regular file that neither its
group nor others can write and,
executed directly with no argument as the non-root user, prints its usage and exits 2.
