# General-purpose Minecraft dashboard implementation plan

Approved direction: separate administrator and viewer accounts; Minecraft-changing capabilities disabled by default. Preserve the existing React/Go stack, artwork and read-only Minecraft integration. Never modify or restart the Minecraft container, execute commands inside it, or write its files.

## Deliverables

1. Account boundary: provision named accounts with Argon2id password hashes in a private file; role-bearing signed sessions, expiry, logout, login throttling and same-origin mutation checks. No shared-secret login fallback. Test invalid credentials, viewer denial and session tampering before implementing.
2. Reliable read-only sources: bounded incremental log reads with rotation support; preserve timestamps during privacy redaction; independent Minecraft status ping; cached filesystem measurements; discover backup metadata without implying integrity. Test fake protocol peers and temporary fixture files, never production mutations.
3. Player details: preserve existing statistics, support vanilla directory layouts, expose UUID and available stats/advancements, unknown missing values, historical join dates, search/filter/sort and profile views. Verify using synthetic fixtures.
4. Dashboard: configurable generic identity, live status and diagnostics, useful World/Console/Backups/Settings pages, console search/severity/pause/copy/export/clear, parsed chat separated from logs, local preferences. Unsupported inventory, maps, metrics and operations have explicit capability explanations. Browser-check all six pages, login/logout, viewer permissions and desktop/mobile containment.
5. Documentation and release: update setup/security/migration/backup guidance, remove family-specific branding, retain original artwork without adding source sheets to Git. Run Go tests/vet and frontend tests/build, review changes, create focused commits and push to the existing GitHub remote.

## Boundaries and verification

- Account provisioning is dashboard-only; credentials never enter Git or API responses.
- Admin role alone does not enable chat, commands, world changes, lifecycle or backup writes.
- Backup listing is read-only and integrity remains Unknown without verification.
- Monitoring errors must not masquerade as zero players or an offline Minecraft server.
- File scans are bounded/cached; log timestamps and duplicate messages must remain meaningful.
- No test sends an administrative command to the live server.
- Existing SSE may be retained for monitoring; any unimplemented specification items must be disclosed, not claimed as finished.

## Progress

- Initial inspection: overview and player summary exist; four other pages are placeholders. Shared access-secret authentication has no roles. Status incorrectly equates failed RCON with server offline. Minecraft mutations are currently disabled.
- Implemented the scoped deliverables above; remaining product-vision integrations are explicitly listed in `feature-coverage.md`.
- Read-only review identified and resolved startup-history gaps, snapshot/stream ordering, unavailable-presence filtering and protected history-storage paths, with regression tests.
- Browser verification covered all six pages, administrator/viewer sign-in, player search/profile, disabled chat/commands, desktop internal scrolling and 390px mobile reflow. Browser inspection also found and fixed narrow metric labels and partial-history loss from an oversized archive.
- Final verification: Go tests and vet pass; 12 frontend tests and production build pass. Deployed private admin/viewer logins return 204; audit access is 200 for admin and 403 for viewer. Chat/console capabilities remain false. Temporary browser-test accounts were removed; only dashboard containers were rebuilt.
