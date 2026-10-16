### Changed
- Activate the release-owned container pipeline: umbrella releases build and verify all nine images and publish their immutable references with the binary artifacts.
- The tag-push release resolves its authorizer from the selected pull request detail endpoint, since the commit pull-request list omits the merger. Both API responses must match the merged release marker and commit; missing or malformed details refuse publication. Offline tests execute the live and staged resolve shells against the actual list/detail response shapes.
