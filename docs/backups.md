# Backup catalog and restoration boundary

The dashboard lists regular archive files from an explicitly configured, read-only backup directory. It skips symbolic links and limits directory enumeration to 1000 entries. Modification time is not guaranteed creation time. Existence is not proof of completeness, checksum integrity or a consistent world snapshot: every discovered archive starts with **Unknown** integrity.

For Compose, add a dedicated read-only bind mount to the **dashboard backend only** and set `DASHBOARD_BACKUP_PATH` to its container path. No backup path is guessed from the world directory. Downloads, deletion, schedules, retention enforcement and automatic verification are not provided in this release.

## Safe external workflow

1. Choose a separately authorized backup service with an explicit consistency mechanism (storage snapshot or coordinated server save). Do not treat an ordinary copy of live files as guaranteed consistent.
2. Write archives only to its dedicated backup destination, never the live world path.
3. Verify the archive checksum against a trusted creation-time manifest and test extraction into a new, isolated directory. A checksum created after discovering a file cannot prove its original correctness.
4. Test an isolated Minecraft instance using the extracted copy; never mount it over production data.
5. Plan any production restore as a separate maintenance operation: obtain approval, arrange downtime, preserve a pre-restore copy and define rollback. The dashboard will not execute these steps or overwrite the live world.

No backup creation, deletion or restoration operation is run as part of dashboard verification.
