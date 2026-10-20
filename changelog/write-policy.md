### Changed
- Extract shared write decisions into pure `deskcore/writepolicy` and retain writeguard as the hook observation/protocol adapter, with inherited behavior, production mutation checks and offline external-consumer fixtures. Managed activation, publication and the staged CI/release test hunk remain separate gates.
