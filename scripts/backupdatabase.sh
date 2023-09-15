_now=$(date +"%Y_%m_%d")
_file="./archive/nanoymousdb_$_now.dump"

pg_dump -U postgres -Fc nanonymousdb > "$_file"

# Delete everything older than a month but still keep one a month
find ./archive/* -type f -not -path './archive/*01.dump' -mtime +30 -exec rm {} \;
