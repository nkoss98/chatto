-- name: MigrationMessage :one
SELECT message
FROM initial_migration
         LIMIT 1;
