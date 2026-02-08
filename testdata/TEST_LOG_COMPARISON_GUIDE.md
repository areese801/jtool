# Test Log Comparison Guide

## Files Created

1. **baseline.log** - Left/Baseline file (older version)
2. **comparison.log** - Right/Comparison file (newer version)

## Expected Comparison Results

### 🟢 ADDED Paths (only in comparison.log)

These paths will show as **green** with "ADDED" status:

- `.record.email` - New email field added to users
- `.record.department` - New department field added to users
- `.value.last_sync` - New timestamp in STATE records
- `.record.status` - New status field added to orders (replaces deprecated_status)
- `.stream` = "products" - Entire new stream type
- `.record.product_id` - Fields from new products stream
- `.record.name` (in products context)
- `.record.price`

### 🔴 REMOVED Paths (only in baseline.log)

These paths will show as **red** with "REMOVED" status:

- `.record.legacy_field` - Removed from users
- `.record.deprecated_status` - Removed from orders

### 🟠 CHANGED Paths (in both, but different counts)

These paths will show as **orange** with "CHANGED" status:

- `.type` - Count changed (6 → 12)
- `.stream` - Count changed (6 → 12)
- `.record.id` - Count changed (3 → 5 users)
- `.record.name` - Count changed (3 → 7 total, different contexts)
- `.record.role` - Count changed (3 → 5)
- `.record.order_id` - Count changed (2 → 4)
- `.record.user_id` - Count changed (2 → 4)
- `.record.total` - Count changed (2 → 4)
- `.value.count` - Count changed (1 → 1, value changed 3 → 5)

### ⚪ EQUAL Paths (unchanged)

Unlikely to have many truly equal paths given the different data volumes.

## What to Look For in the UI

1. **Sorting**: Paths should be ordered: Removed → Added → Changed → Equal
2. **Color Coding**:
   - Green background for added rows
   - Red background for removed rows
   - Orange background for changed rows
3. **Delta Values**:
   - Positive deltas (e.g., +2) in green
   - Negative deltas (e.g., -3) in red
   - Zero deltas grayed out
4. **Empty Cells**:
   - Added paths will have "—" in Left columns
   - Removed paths will have "—" in Right columns
5. **Top Values** (if enabled):
   - Should show side-by-side comparison of top values
   - Added paths show only right values
   - Removed paths show only left values

## Stats Summary Expected

- **Total Paths**: ~20-25 paths
- **Added**: ~8-10 paths
- **Removed**: ~2 paths
- **Changed**: ~8-10 paths
- **Equal**: ~0-2 paths

## Testing Interactions

1. **Click path cell**: Should copy jq command for left (baseline) file
2. **Filter "Show only changes"**: Should hide any equal paths
3. **Show top values**: Should display side-by-side value lists
4. **Hover over rows**: Should highlight with darker background
