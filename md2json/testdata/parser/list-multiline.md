* Create a backup:

    ```
    pg_dump watcher > watcher-$(date +%s)_backup_dump.sql
    ```

* Determine dependencies:

    !!! warning
        admonition warning
