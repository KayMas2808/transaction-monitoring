# Use the official PostgreSQL image as the base
FROM postgres:15-alpine
# Copy the migration script into the Docker entrypoint directory
# This ensures the script is run when the container is first created
COPY ./backend/database/migrations.sql /docker-entrypoint-initdb.d/init.sql
# Ensure the script has correct permissions (optional, but good practice)
RUN chmod 644 /docker-entrypoint-initdb.d/init.sql \
    && chown postgres:postgres /docker-entrypoint-initdb.d/init.sql
