"""8-step bidirectional sync algorithm."""
from __future__ import annotations

from cask.sync.protocol import ResourceSync, SyncOptions, SyncStats
from cask.executor.protocol import Executor


async def run_sync(
    sync: ResourceSync,
    config: object,
    exec: Executor,
    options: SyncOptions,
) -> SyncStats:
    """Run the 8-step bidirectional sync algorithm."""
    stats = SyncStats()

    # Step 1: Gather
    host_resources = await sync.get_host_resources(exec)
    config_resources = sync.get_config_resources(config)

    # Step 2: Build lookup maps
    host_map = {sync.resource_id(r): r for r in host_resources}
    config_map = {sync.resource_id(r): r for r in config_resources}

    # Step 3: Categorize
    host_ids = set(host_map.keys())
    config_ids = set(config_map.keys())
    to_apply = config_ids - host_ids
    common = config_ids & host_ids
    undeclared = host_ids - config_ids

    # Step 4: Check common for updates
    to_update = {
        rid for rid in common
        if sync.needs_update(host_map[rid], config_map[rid])
    }

    # Step 5: Handle undeclared
    to_remove: set[str] = set()
    if options.no:
        to_remove = undeclared
    elif options.yes:
        stats.kept = len(undeclared)
    else:
        # Interactive mode — for now, keep all (CLI will handle prompting)
        stats.kept = len(undeclared)

    # Step 6: Apply + Update
    for rid in to_apply:
        result = await sync.apply(config_map[rid], exec)
        if result.ok:
            stats.applied += 1
        else:
            stats.failed += 1

    for rid in to_update:
        result = await sync.apply(config_map[rid], exec)
        if result.ok:
            stats.updated += 1
        else:
            stats.failed += 1

    # Step 7: Remove
    for rid in to_remove:
        result = await sync.remove(rid, exec)
        if result.ok:
            stats.removed += 1
        else:
            stats.failed += 1

    # Step 8: Return stats
    return stats
