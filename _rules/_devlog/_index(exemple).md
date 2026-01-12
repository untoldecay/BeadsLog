# Development Log Index

This index provides a concise record of all development work for easy scanning and pattern recognition across sessions.

## Nomenclature Rules:
- **[fix]** - Bug fixes and error resolution
- **[feature]** - New feature implementation
- **[enhance]** - Improvements to existing functionality
- **[rationalize]** - Code cleanup and consolidation
- **[deploy]** - Deployment activities and version releases
- **[security]** - Security fixes and vulnerability patches
- **[debug]** - Troubleshooting and investigation
- **[test]** - Testing and validation activities

## Work Index

| Subject | Problems | Date | Devlog |
|---------|----------|------|---------|
| [rationalize] Export endpoint system | 5 redundant export endpoints causing confusion and maintenance issues | 2025-11-22 | [2025-11-22_export-rationalization-and-vector-fixes.md](2025-11-22_export-rationalization-and-vector-fixes.md) |
| [fix] Vector export format detection | Added intelligent format selection for vector vs regular tables | 2025-11-22 | [2025-11-22_export-rationalization-and-vector-fixes.md](2025-11-22_export-rationalization-and-vector-fixes.md) |
| [feature] Remote MCP SSE | Implemented Remote SSE transport for MCP with absolute URLs | 2026-01-10 | [2026-01-10_mcp-remote-sse-integration.md](2026-01-10_mcp-remote-sse-integration.md) |
| [fix] MCP Connection Stability | Resolved Nginx buffering and 400 Bad Request issues | 2026-01-10 | [2026-01-10_mcp-remote-sse-integration.md](2026-01-10_mcp-remote-sse-integration.md) |
| [enhance] MCP Schema Cache | Fixed internal 401 errors during cache warmup with system tokens | 2026-01-10 | [2026-01-10_mcp-remote-sse-integration.md](2026-01-10_mcp-remote-sse-integration.md) |
| [rationalize] Backend ETA | Simplified backend feature log into clear 1-row index | 2026-01-10 | [2026-01-10_mcp-remote-sse-integration.md](2026-01-10_mcp-remote-sse-integration.md) |
| [enhance] API client export support | Frontend needed to use rationalized export endpoints with format parameter | 2025-11-22 | [2025-11-22_export-rationalization-and-vector-fixes.md](2025-11-22_export-rationalization-and-vector-fixes.md) |
| [deploy] Export rationalization v4.1.141 | Successfully deployed unified export system to staging | 2025-11-22 | [2025-11-22_export-rationalization-and-vector-fixes.md](2025-11-22_export-rationalization-and-vector-fixes.md) |
| [security] Import endpoint authentication vulnerabilities | Critical security gaps in public import endpoints requiring JWT protection | 2025-11-23 | [2025-11-23_session_summary.md](2025-11-23_session_summary.md) |
| [enhance] File upload size limits and validation | Increased from 50MB to 1GB with dual-layer MIME and extension validation | 2025-11-23 | [2025-11-23_session_summary.md](2025-11-23_session_summary.md) |
| [security] SQL content scanning and pattern detection | Real-time threat detection with advanced pattern recognition for malicious SQL | 2025-11-23 | [2025-11-23_session_summary.md](2025-11-23_session_summary.md) |
| [docs] API Connection Guide security updates | Updated documentation for new authentication requirements and security features | 2025-11-23 | [2025-11-23_session_summary.md](2025-11-23_session_summary.md) |
| [fix] Authentication system mismatch | Fixed legacy token verification to use database instead of in-memory Map storage | 2025-11-23 | [2025-11-23_session_summary.md](2025-11-23_session_summary.md) |
| [resolve] JWT authentication system confirmed working | Frontend correctly configured, usage error identified between old and new endpoints | 2025-11-23 | [2025-11-23_session_summary.md](2025-11-23_session_summary.md) |
| [deploy] Cloudron deployment version mismatch and multi-arch build fix | Nuclear deployment with v4.1.156 to staging, fixed ARM64 build failures and version synchronization | 2025-11-23 | [2025-11-23_session_summary.md](2025-11-23_session_summary.md) |
| [docs] API Development Contract completion | Comprehensive API guidelines deployed with contract standards and development checklist | 2025-11-23 | [2025-11-23_session_summary.md](2025-11-23_session_summary.md) |
| [fix] AddColumnModal normalization | Standardized CRUD patterns implemented using apiClient and useAuthenticatedApi hook | 2025-11-23 | [2025-11-23_session_summary.md](2025-11-23_session_summary.md) |
| [fix] Column management debugging and layout alignment | Fixed "Missing required fields" bug with comprehensive debugging, aligned ManageColumnsModal layout with structure tab pattern | 2025-11-23 | [2025-11-23_column-management-fixes-and-layout-rationalization.md](2025-11-23_column-management-fixes-and-layout-rationalization.md) |
| [docs] API Development Contract specification | Created comprehensive 11-section contract establishing mandatory standards for API development to prevent future misalignments | 2025-11-23 | [../requirements/API_DEVELOPMENT_CONTRACT.md](../requirements/API_DEVELOPMENT_CONTRACT.md) |
| [docs] API Development Checklist | Created pre-flight validation checklist for developers ensuring compliance with API Development Contract standards | 2025-11-23 | [../requirements/API_DEVELOPMENT_CHECKLIST.md](../requirements/API_DEVELOPMENT_CHECKLIST.md) |
| [enhance] AddColumnModal normalization | Normalized AddColumnModal with standardized CRUD patterns using apiClient and useAuthenticatedApi hook per contract standards | 2025-11-23 | [../requirements/API_DEVELOPMENT_CONTRACT.md](../requirements/API_DEVELOPMENT_CONTRACT.md) |
| [fix] Double slash URL construction bug | Critical production bug where BASE_URL='/' + url='/admin/endpoint' = '//admin/endpoint' causing fetch failures | 2025-11-26 | [2025-11-26_double-slash-url-bug-fix.md](2025-11-26_double-slash-url-bug-fix.md) |
| [deploy] Production fix deployment v4.1.162 | Successfully deployed URL construction fix to boiler-alpha.decaylab.com, restored import and column deletion functionality | 2025-11-26 | [2025-11-26_double-slash-url-bug-fix.md](2025-11-26_double-slash-url-bug-fix.md) |
| [docs] End-of-session documentation update | Updated backend ETA, roadmap, and current plan with critical bug fix completion status | 2025-11-26 | [2025-11-26_double-slash-url-bug-fix.md](2025-11-26_double-slash-url-bug-fix.md) |
| [fix] Manage Columns modal JavaScript error | Fixed "id is not defined" error by correcting useSortable hook parameter and cleaning up component structure | 2025-11-29 | [2025-11-29_manage-columns-modal-javascript-fix.md](2025-11-29_manage-columns-modal-javascript-fix.md) |
| [fix] SQL Export Corruption | Relaxed aggressive cleaning of SQL exports to preserve COPY data rows with backslashes | 2025-12-23 | [2025-12-23_sql-export-fix-and-staging-deployment.md](2025-12-23_sql-export-fix-and-staging-deployment.md) |
| [fix] Table Export CSV | Implemented missing convertToCSV helper and enhanced UI with CSV/SQL choice | 2025-12-23 | [2025-12-23_sql-export-fix-and-staging-deployment.md](2025-12-23_sql-export-fix-and-staging-deployment.md) |
| [deploy] Staging boiler-stagging | Created new staging environment at v4.1.174 for verification | 2025-12-23 | [2025-12-23_sql-export-fix-and-staging-deployment.md](2025-12-23_sql-export-fix-and-staging-deployment.md) |
| [doc] PG_DUMP Guide | Added comprehensive guide for manual database exports via Cloudron CLI | 2025-12-23 | [2025-12-23_sql-export-fix-and-staging-deployment.md](2025-12-23_sql-export-fix-and-staging-deployment.md) |
| [feature] Migration System | Implemented automatic delta patching system with tracking table | 2025-12-23 | [2025-12-23_migration-system-implementation.md](2025-12-23_migration-system-implementation.md) |
| [rationalize] Legacy Migrations | Cleaned up merged migration files to start fresh with v1 schema | 2025-12-23 | [2025-12-23_migration-system-implementation.md](2025-12-23_migration-system-implementation.md) |
| [fix] Database Persistence | Relocated PGDATA to /app/data/postgres to ensure data survives updates | 2025-12-23 | [2025-12-23_overview-fix-and-migration-test.md](2025-12-23_overview-fix-and-migration-test.md) |
| [fix] Overview Row Count | Fixed numeric summation of total row counts across all tables | 2025-12-23 | [2025-12-23_overview-fix-and-migration-test.md](2025-12-23_overview-fix-and-migration-test.md) |
| [test] Migration System | Verified automatic schema update via 001_migration_system_test.sql | 2025-12-23 | [2025-12-23_overview-fix-and-migration-test.md](2025-12-23_overview-fix-and-migration-test.md) |
| [feature] Table Pagination | Integrated top-aligned pagination with direct page jump and true record counting | 2025-12-23 | [2025-12-23_pagination-and-ui-refactor.md](2025-12-23_pagination-and-ui-refactor.md) |
| [fix] Disappearing Rows bug | Fixed redundant local slicing causing empty tables on next pages | 2025-12-23 | [2025-12-23_pagination-and-ui-refactor.md](2025-12-23_pagination-and-ui-refactor.md) |
| [enhance] UI Reorganization | Split action bar into three functional ButtonGroups (Nav, Density, Management) | 2025-12-23 | [2025-12-23_pagination-and-ui-refactor.md](2025-12-23_pagination-and-ui-refactor.md) |
| [docs] Token System Documentation | Created comprehensive guide for Access/Refresh token architecture and auto-refresh logic | 2025-12-24 | [2025-12-24_token-system-documentation.md](2025-12-24_token-system-documentation.md) |
| [fix] Sidebar User Profile Display | Fixed issue where sidebar showed generic "User" info on page reload by fetching user profile during auth check | 2025-12-24 | [2025-12-24_sidebar-user-profile-fix.md](2025-12-24_sidebar-user-profile-fix.md) |
| [feature] Table sorting implementation | Added backend-driven sorting with context-aware UI and Null handling | 2026-01-01 | [2026-01-01_session_summary.md](2026-01-01_session_summary.md) |
| [fix] Infinite request loop | Fixed React useEffect dependency issue causing recursive API calls | 2026-01-01 | [2026-01-01_session_summary.md](2026-01-01_session_summary.md) |
| [fix] Sort persistence | Added logic to reset sorting state when switching tables | 2026-01-01 | [2026-01-01_session_summary.md](2026-01-01_session_summary.md) |
| [enhance] API error handling | Improved response parsing to handle non-JSON errors (429s) by cloning responses | 2026-01-01 | [2026-01-01_session_summary.md](2026-01-01_session_summary.md) |
| [enhance] Table View UI optimizations | Improved pagination borders, Sort button styling with Badge, added Switch for Null toggle | 2026-01-01 | [2026-01-01_session_summary.md](2026-01-01_session_summary.md) |
| [fix] Sort state persistence bug | Fixed issue where sort configuration persisted across table switches | 2026-01-01 | [2026-01-01_session_summary.md](2026-01-01_session_summary.md) |
| [feature] Supabase import implementation | Implemented backend and frontend for importing databases directly from Supabase | 2026-01-03 | [2026-01-03_supabase_import_and_fixes.md](2026-01-03_supabase_import_and_fixes.md) |
| [fix] Neon tab crash | Fixed undefined variable crash in ImportExportView Neon tab | 2026-01-03 | [2026-01-03_supabase_import_and_fixes.md](2026-01-03_supabase_import_and_fixes.md) |
| [fix] Supabase service syntax error | Fixed SyntaxError caused by nested comments in supabaseService.js | 2026-01-03 | [2026-01-03_supabase_import_and_fixes.md](2026-01-03_supabase_import_and_fixes.md) |
| [enhance] Supabase import robustness | Handled pg_restore exit code 1 as success with warnings for Supabase imports | 2026-01-03 | [2026-01-03_supabase_import_and_fixes.md](2026-01-03_supabase_import_and_fixes.md) |
| [fix] API Token Generation Endpoint | Refactored POST /api/tokens to use verifyAuthToken instead of raw database credentials | 2026-01-05 | [2026-01-05_api-token-fix-and-auth-docs.md](2026-01-05_api-token-fix-and-auth-docs.md) |
| [docs] Authentication Workflow Clarification | Overhauled API and Token guides to explicitly state Access Token prerequisite | 2026-01-05 | [2026-01-05_api-token-fix-and-auth-docs.md](2026-01-05_api-token-fix-and-auth-docs.md) |
| [fix] API Token Persistence & Integration | Implemented database-backed token storage and fixed frontend integration/loop bugs | 2026-01-05 | [2026-01-05_api-token-fix-and-auth-docs.md](2026-01-05_api-token-fix-and-auth-docs.md) |
| [feature] MCP Server Implementation | Added initial MCP server codebase for AI agent integration | 2026-01-05 | [2026-01-05_api-token-fix-and-auth-docs.md](2026-01-05_api-token-fix-and-auth-docs.md) |
| [feature] Boiler MCP Evolution Roadmap | Added a 4-day plan for a schema-driven, intention-focused MCP server | 2026-01-05 | [2026-01-05_api-token-fix-and-auth-docs.md](2026-01-05_api-token-fix-and-auth-docs.md) |
| [enhance] MCP Server Phase 3 data writing tools | Converted insert/update/delete tools to project-scoped pattern with auto-cleanup | 2026-01-08 | [2026-01-08_mcp-phase3-data-writing-and-auto-cleanup.md](2026-01-08_mcp-phase3-data-writing-and-auto-cleanup.md) |
| [enhance] Auto-cleanup for deleted databases/tables | Implemented automatic removal of deleted tables from config/projects.json | 2026-01-08 | [2026-01-08_mcp-phase3-data-writing-and-auto-cleanup.md](2026-01-08_mcp-phase3-data-writing-and-auto-cleanup.md) |
| [debug] Perplexity tool discoverability investigation | Investigated why write tools don't appear in Perplexity's tool list (client-side filtering confirmed) | 2026-01-08 | [2026-01-08_mcp-phase3-data-writing-and-auto-cleanup.md](2026-01-08_mcp-phase3-data-writing-and-auto-cleanup.md) |

---

*This index is automatically updated when devlogs are created via the generation prompt. All work subjects must be referenced in this index following the established nomenclature rules.*