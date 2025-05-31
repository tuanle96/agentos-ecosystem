-- Week 6 Day 1: Database Performance Analysis
-- AgentOS Performance Optimization Implementation

-- Enable timing for all queries
\timing on

-- Set output format
\pset format wrapped
\pset columns 120

-- Create performance analysis report
\echo '🗄️  AgentOS Database Performance Analysis'
\echo '========================================'
\echo ''

-- 1. Database Overview
\echo '📊 Database Overview'
\echo '-------------------'

SELECT 
    pg_database.datname as database_name,
    pg_size_pretty(pg_database_size(pg_database.datname)) as size,
    (SELECT count(*) FROM pg_stat_activity WHERE datname = pg_database.datname) as active_connections
FROM pg_database 
WHERE datname = current_database();

\echo ''

-- 2. Table Statistics
\echo '📋 Table Statistics'
\echo '------------------'

SELECT 
    schemaname,
    tablename,
    n_live_tup as live_rows,
    n_dead_tup as dead_rows,
    n_tup_ins as inserts,
    n_tup_upd as updates,
    n_tup_del as deletes,
    last_vacuum,
    last_autovacuum,
    last_analyze,
    last_autoanalyze
FROM pg_stat_user_tables 
ORDER BY n_live_tup DESC;

\echo ''

-- 3. Index Usage Statistics
\echo '🔍 Index Usage Statistics'
\echo '------------------------'

SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes 
ORDER BY idx_scan DESC;

\echo ''

-- 4. Unused Indexes (potential for removal)
\echo '🚫 Unused Indexes'
\echo '----------------'

SELECT 
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes 
WHERE idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC;

\echo ''

-- 5. Query Performance Statistics (requires pg_stat_statements)
\echo '⚡ Query Performance Statistics'
\echo '------------------------------'

-- Check if pg_stat_statements is available
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements') THEN
        RAISE NOTICE 'pg_stat_statements extension is available';
    ELSE
        RAISE NOTICE 'pg_stat_statements extension is NOT available - install for detailed query stats';
    END IF;
END $$;

-- Show top queries by total time (if pg_stat_statements is available)
SELECT 
    substring(query, 1, 100) as short_query,
    calls,
    round(total_exec_time::numeric, 2) as total_time_ms,
    round(mean_exec_time::numeric, 2) as avg_time_ms,
    round((100 * total_exec_time / sum(total_exec_time) OVER ())::numeric, 2) as percent_total_time,
    rows as total_rows
FROM pg_stat_statements 
WHERE query NOT LIKE '%pg_stat_statements%'
ORDER BY total_exec_time DESC 
LIMIT 20;

\echo ''

-- 6. Connection Statistics
\echo '🔌 Connection Statistics'
\echo '-----------------------'

SELECT 
    state,
    count(*) as connection_count,
    round(avg(extract(epoch from (now() - state_change)))::numeric, 2) as avg_duration_seconds
FROM pg_stat_activity 
WHERE datname = current_database()
GROUP BY state
ORDER BY connection_count DESC;

\echo ''

-- 7. Lock Statistics
\echo '🔒 Lock Statistics'
\echo '-----------------'

SELECT 
    mode,
    locktype,
    count(*) as lock_count
FROM pg_locks 
GROUP BY mode, locktype
ORDER BY lock_count DESC;

\echo ''

-- 8. Buffer Cache Hit Ratio
\echo '💾 Buffer Cache Hit Ratio'
\echo '------------------------'

SELECT 
    'Buffer Cache Hit Ratio' as metric,
    round(
        (sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read) + 1))::numeric * 100, 
        2
    ) as hit_ratio_percent
FROM pg_statio_user_tables;

\echo ''

-- 9. Table Sizes
\echo '📏 Table Sizes'
\echo '-------------'

SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) as index_size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

\echo ''

-- 10. Slow Queries Analysis
\echo '🐌 Slow Queries Analysis'
\echo '-----------------------'

-- Show current long-running queries
SELECT 
    pid,
    now() - pg_stat_activity.query_start as duration,
    query,
    state
FROM pg_stat_activity 
WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes'
AND state = 'active';

\echo ''

-- 11. Vacuum and Analyze Status
\echo '🧹 Vacuum and Analyze Status'
\echo '----------------------------'

SELECT 
    schemaname,
    tablename,
    last_vacuum,
    last_autovacuum,
    vacuum_count,
    autovacuum_count,
    last_analyze,
    last_autoanalyze,
    analyze_count,
    autoanalyze_count
FROM pg_stat_user_tables
ORDER BY last_autovacuum DESC NULLS LAST;

\echo ''

-- 12. Database Configuration Check
\echo '⚙️  Database Configuration Check'
\echo '-------------------------------'

SELECT 
    name,
    setting,
    unit,
    context,
    short_desc
FROM pg_settings 
WHERE name IN (
    'shared_buffers',
    'effective_cache_size',
    'maintenance_work_mem',
    'checkpoint_completion_target',
    'wal_buffers',
    'default_statistics_target',
    'random_page_cost',
    'effective_io_concurrency',
    'work_mem',
    'max_connections'
)
ORDER BY name;

\echo ''

-- 13. Performance Recommendations
\echo '💡 Performance Recommendations'
\echo '------------------------------'

-- Check for tables that might need VACUUM
WITH vacuum_stats AS (
    SELECT 
        schemaname,
        tablename,
        n_dead_tup,
        n_live_tup,
        CASE 
            WHEN n_live_tup > 0 THEN round((n_dead_tup::float / n_live_tup::float) * 100, 2)
            ELSE 0 
        END as dead_tuple_percent
    FROM pg_stat_user_tables
    WHERE n_live_tup > 1000  -- Only consider tables with significant data
)
SELECT 
    'VACUUM NEEDED' as recommendation,
    schemaname || '.' || tablename as table_name,
    dead_tuple_percent || '%' as dead_tuples,
    'Consider running VACUUM on this table' as action
FROM vacuum_stats 
WHERE dead_tuple_percent > 20
ORDER BY dead_tuple_percent DESC;

-- Check for missing indexes on foreign keys
WITH foreign_keys AS (
    SELECT 
        tc.table_name,
        kcu.column_name,
        tc.constraint_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu 
        ON tc.constraint_name = kcu.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
),
indexed_columns AS (
    SELECT 
        t.relname as table_name,
        a.attname as column_name
    FROM pg_index i
    JOIN pg_class t ON t.oid = i.indrelid
    JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(i.indkey)
    WHERE t.relkind = 'r'
    AND t.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = 'public')
)
SELECT 
    'INDEX NEEDED' as recommendation,
    fk.table_name,
    fk.column_name,
    'Consider adding index on foreign key column' as action
FROM foreign_keys fk
LEFT JOIN indexed_columns ic 
    ON fk.table_name = ic.table_name 
    AND fk.column_name = ic.column_name
WHERE ic.column_name IS NULL;

\echo ''
\echo '✅ Database Performance Analysis Complete'
\echo ''

-- Save timing information
\echo 'Analysis completed at:' 
SELECT now() as analysis_timestamp;
