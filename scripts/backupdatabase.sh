_now=$(date +"%Hh-%Y_%m_%d")
_file="./backups/nanoymousdb_$_now.dump"

pg_dump -U go -Fc nanonymousdb > "$_file"

# Delete everything older than a month but still keep one a month
find ./backups/* -type f -not -path './backups/*01.dump' -mtime +30 -exec rm {} \;
