# Pruning

Setting Pruning fields in the IAVL tree can optimize performance by only writing versions to disk if they are meant to be persisted indefinitely. Versions that are known to be deleted eventually are temporarily held in memory until they are ready to be pruned. This greatly reduces the I/O load of IAVL.

We can set custom pruning fields in IAVL using: `NewMutableTreeWithOpts`

## Current design

### NodeDB
NodeDB has extra fields:

```go
recentDB    dbm.DB     // Memory node storage.
recentBatch dbm.Batch  // Batched writing buffer for memDB.

// Pruning fields
keepEvery  int64n // Saves version to disk periodically
keepRecent int64  // Saves recent versions in memory
```

If version is not going to be persisted to disk, the version is simply saved in `recentDB` (typically a `memDB`)
If version is persisted to disk, the version is written to `recentDB` **and** `snapshotDB` (typically `levelDB`)

For example, setting keepEvery to 1 and keepRecent to 0 (which is the default setting) will persist every version to snapshot and skip storing
anything in memDB. Setting keepEvery to 10000 and keepRecent to 100 (default Cosmos SDK pruning strategy) is snapshotting every 10000th version and will keep the last 100 versions
in memDB.

#### Orphans:

Save orphan to `memDB` under `o|toVersion|fromVersion`.

If there exists snapshot version `snapVersion` s.t. `fromVersion < snapVersion < toVersion`, save orphan to disk as well under `o|snapVersion|fromVersion`.
NOTE: in unlikely event, that two snapshot versions exist between `fromVersion` and `toVersion`, we use closest snapshot version that is less than `toVersion`

Can then simply use the old delete algorithm with some minor simplifications/optimizations

### MutableTree

MutableTree can be instantiated with a pruning-aware NodeDB.

When `MutableTree` saves a new Version, it also calls `PruneRecentVersions` on nodeDB which causes oldest version in recentDB (`latestVersion - keepRecent`) to get pruned.
