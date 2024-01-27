* Create a backup:

    ```
    pg_dump watcher > watcher-$(date +%s)_backup_dump.sql
    ```
    
* Determine dependencies:

    ```
    apt-cache show watcher=20.06 | egrep 'Depends|Suggests:'
    ```
